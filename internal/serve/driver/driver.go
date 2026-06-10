package driver

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgconn/ctxwatch"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/timescale/ghost/internal/log"
	"github.com/timescale/ghost/internal/serve/api"
)

// applicationName identifies this application to the database via the
// application_name connection parameter.
const applicationName = "ghost"

// Source is the value reported in the Source field of an [api.NormalizedError]
// for errors originating from this driver. The driver only supports Postgres,
// so this is constant.
const Source = "postgresql"

// Driver executes arbitrary SQL queries against a Postgres database instance.
// It wraps [sql.DB] and [sql.Conn] and provides additional methods for
// normalizing query results and errors.
type Driver struct {
	db     *sql.DB
	conn   *sql.Conn
	pgConn *pgconn.PgConn
	tracer *postgresQueryTracer
}

// Open establishes a new Postgres [Driver] connection using the provided DSN.
// It opens the connection pool, acquires a single dedicated connection, and
// wires up the query tracer and cancellation handler. If an error occurs after
// a resource is acquired, that resource is closed before returning.
func Open(ctx context.Context, dsn string) (d *Driver, err error) {
	pgxCfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	pgxCfg.DefaultQueryExecMode = pgx.QueryExecModeExec
	pgxCfg.RuntimeParams["application_name"] = applicationName

	tracer := &postgresQueryTracer{}
	pgxCfg.Tracer = tracer

	pgxCfg.BuildContextWatcherHandler = newCancelHandler

	db := stdlib.OpenDB(*pgxCfg)
	defer closeDBOnErr(ctx, db, &err)

	db.SetMaxIdleConns(0)

	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer closeConnOnErr(ctx, conn, &err)

	var pgConn *pgconn.PgConn
	if err := conn.Raw(func(driverConn any) error {
		pgConn = driverConn.(*stdlib.Conn).Conn().PgConn()
		return nil
	}); err != nil {
		return nil, fmt.Errorf("error getting raw driver connection: %w", err)
	}

	return &Driver{
		db:     db,
		conn:   conn,
		pgConn: pgConn,
		tracer: tracer,
	}, nil
}

// cancelDeadlineDelay bounds how long a canceled query may keep running if the
// cancel request is lost or ignored; after this delay the connection deadline
// expires and breaks the connection so the query doesn't block until the run
// times out.
const cancelDeadlineDelay = 10 * time.Second

// newCancelHandler builds the query-cancellation handler for a connection.
// pgx's default handler cancels by closing the underlying connection, which
// would tear down the session — not acceptable here.
// [pgconn.CancelRequestContextWatcherHandler] instead sends the native
// Postgres cancel ([pgconn.PgConn.CancelRequest]) over a fresh connection,
// leaving the session intact. As a fallback, it also sets a deadline on the
// connection so a cancel request that never takes effect can't hang the query
// indefinitely.
func newCancelHandler(pgConn *pgconn.PgConn) ctxwatch.Handler {
	return &pgconn.CancelRequestContextWatcherHandler{
		Conn:          pgConn,
		DeadlineDelay: cancelDeadlineDelay,
	}
}

// Ping checks whether the underlying database connection is alive, returning an
// error if not. It is not safe to call concurrently with itself or Query.
func (d *Driver) Ping(ctx context.Context) error {
	return d.conn.PingContext(ctx)
}

// Query issues an SQL query and returns the results. Canceling the passed-in
// context cancels the in-progress query via the connection's context-watcher
// handler. It is not safe to call concurrently with itself or Ping.
func (d *Driver) Query(ctx context.Context, query string) (*Rows, error) {
	rows, err := d.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return &Rows{
		Rows:   rows,
		tracer: d.tracer,
	}, nil
}

// NormalizeError normalizes an error returned from the database driver. The
// returned error also carries information about whether the query was canceled,
// timed out, or executed on a broken database connection.
func (d *Driver) NormalizeError(ctx context.Context, err error) *api.NormalizedError {
	ctxErr := context.Cause(ctx)
	normalized := &api.NormalizedError{
		Message: errMessage(err),
		Source:  Source,
		Fatal:   fatal(err),
		Timeout: errors.Is(ctxErr, context.DeadlineExceeded),
		Cancel:  errors.Is(ctxErr, context.Canceled),
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if strings.EqualFold(pgErr.Severity, "FATAL") {
			normalized.Fatal = true
		}

		// Return a more user-friendly error if the user attempts to run
		// multiple statements in a single query. It's possible we will support
		// this in the future, but it's a little complicated - see #20.
		// NOTE: Error code 42601 is just the generic "syntax error" code, so
		// we also need to check the message.
		if pgErr.Code == "42601" && pgErr.Message == "cannot insert multiple commands into a prepared statement" {
			normalized.Message = ErrMultiStatement.Error()
			return normalized
		}

		normalized.Code = pgErr.Code
		normalized.Detail = pgErr.Detail
		normalized.Hint = pgErr.Hint
		normalized.Message = pgErr.Message
		normalized.Position = pgErr.Position
	}
	return normalized
}

// errMessage returns more user-friendly error messages for some low-level
// errors (such as [io.ErrUnexpectedEOF]), for use in [api.NormalizedError].
func errMessage(err error) string {
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return "the database connection was terminated unexpectedly"
	}
	return err.Error()
}

// Standard error types that indicate a broken/invalid database connection.
var fatalErrs = []error{
	driver.ErrBadConn,
	sql.ErrConnDone,
	io.ErrUnexpectedEOF,
	net.ErrClosed,
}

// fatal returns true if the provided err indicates a broken/invalid database
// connection.
func fatal(err error) bool {
	return slices.ContainsFunc(fatalErrs, func(target error) bool {
		return errors.Is(err, target)
	})
}

// Close waits for all in-progress queries to finish before closing the
// underlying database connection and pool.
func (d *Driver) Close() error {
	var errs []error
	if err := d.conn.Close(); err != nil && !errors.Is(err, sql.ErrConnDone) {
		errs = append(errs, fmt.Errorf("error closing database connection: %w", err))
	}
	if err := d.db.Close(); err != nil {
		errs = append(errs, fmt.Errorf("error closing database connection pool: %w", err))
	}
	return errors.Join(errs...)
}

// postgresQueryTracer implements the [pgx.QueryTracer] interface. It exists
// solely for the purpose of getting access to the [pgconn.CommandTag], which
// contains the number of rows affected (for INSERT/UPDATE/DELETE statements).
// Getting the number of rows affected is not possible via the database/sql
// Query methods (and we can't use the Exec methods, because we don't know for
// sure whether the user's queries will return rows or not).
type postgresQueryTracer struct {
	lastCommandTag *pgconn.CommandTag
}

func (t *postgresQueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	t.lastCommandTag = nil
	return ctx
}

func (t *postgresQueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	t.lastCommandTag = &data.CommandTag
}

func closeDBOnErr(ctx context.Context, db *sql.DB, err *error) {
	if err := closeOnErr(db, err); err != nil {
		logger := log.FromContext(ctx)
		logger.Error("Error closing database connection pool", slog.Any("error", err))
	}
}

func closeConnOnErr(ctx context.Context, conn *sql.Conn, err *error) {
	if err := closeOnErr(conn, err); err != nil {
		logger := log.FromContext(ctx)
		logger.Error("Error closing database connection", slog.Any("error", err))
	}
}

func closeOnErr(closer io.Closer, err *error) error {
	if err == nil || *err == nil {
		return nil
	}
	return closer.Close()
}
