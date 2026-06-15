package serve

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timescale/ghost/internal/log"
	"github.com/timescale/ghost/internal/serve/api"
)

func logRequests(logger *slog.Logger) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Stash the parsed query params so downstream middleware and
			// handlers can read them via queryParamFromContext without
			// re-parsing the raw query string on each access.
			ctx = context.WithValue(ctx, queryParamsKey{}, r.URL.Query())

			// Use the client-provided request ID, or generate one.
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Create a request-scoped logger.
			ctx, logger := log.NewContext(ctx, logger.With(
				slog.String("method", r.Method),
				slog.String("from", r.RemoteAddr),
				slog.String("url", r.URL.String()),
				slog.String("requestId", requestID),
			))

			// Log the request duration.
			defer func(start time.Time) {
				logger.Debug("Request handled",
					slog.Duration("elapsed", time.Since(start)),
				)
			}(time.Now())

			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func handlePanics() Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := log.FromContext(ctx)

			defer func() {
				if r := recover(); r != nil {
					stack := string(debug.Stack())
					logger.Error(fmt.Sprintf("Request handler panic: %s\n%s", r, stack))
					writeError(w, http.StatusInternalServerError, api.ErrInternalServer, logger)
				}
			}()

			handler.ServeHTTP(w, r)
		})
	}
}

func contentTypeJSON() Middleware {
	return contentType("application/json")
}

func contentType(required string) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := log.FromContext(ctx)

			header := r.Header["Content-Type"]
			if len(header) != 1 || header[0] != required {
				logger.Warn("Invalid content type",
					slog.Any("header", header),
				)
				writeError(w, http.StatusBadRequest, &InvalidContentTypeError{
					Required: required,
				}, logger)
				return
			}

			handler.ServeHTTP(w, r)
		})
	}
}

type queryParamsKey struct{}

// queryParamsFromContext returns the parsed query params stashed by the
// logRequests middleware. It panics if that middleware was not configured,
// since that's a programmer error.
func queryParamsFromContext(ctx context.Context) url.Values {
	params, ok := ctx.Value(queryParamsKey{}).(url.Values)
	if !ok {
		panic(errors.New("logRequests middleware not configured"))
	}
	return params
}

// queryParamFromContext returns the named query param from the parsed query
// param map stashed by the logRequests middleware.
func queryParamFromContext(ctx context.Context, name string) string {
	return queryParamsFromContext(ctx).Get(name)
}

// requiredQueryParam validates that the named query param is present and
// non-empty, returning 400 otherwise. It doesn't stash anything: handlers read
// the value via queryParamFromContext like any other param.
func requiredQueryParam(name string) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := log.FromContext(ctx)

			if queryParamFromContext(ctx, name) == "" {
				logger.Warn("Missing required query param", slog.String("name", name))
				writeError(w, http.StatusBadRequest, &RequiredQueryParamError{ParamName: name}, logger)
				return
			}

			handler.ServeHTTP(w, r)
		})
	}
}

type boolQueryParamKey struct {
	name string
}

// boolQueryParamFromContext returns the value of a query param parsed by the
// boolQueryParam middleware. It panics if that middleware was not configured
// for the given name, since that's a programmer error.
func boolQueryParamFromContext(ctx context.Context, name string) bool {
	key := boolQueryParamKey{name: name}
	val, ok := ctx.Value(key).(bool)
	if !ok {
		panic(fmt.Errorf("boolQueryParam middleware not configured for %q", name))
	}
	return val
}

// boolQueryParam parses the named query param as a boolean and stashes it in
// the context for retrieval via boolQueryParamFromContext. A missing or empty
// value yields defaultValue; a present but unparseable value returns 400.
func boolQueryParam(name string, defaultValue bool) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := log.FromContext(ctx)

			val := defaultValue
			if raw := queryParamFromContext(ctx, name); raw != "" {
				parsed, err := strconv.ParseBool(raw)
				if err != nil {
					logger.Warn("Invalid bool query param",
						slog.String("name", name),
						slog.String("value", raw),
					)
					writeError(w, http.StatusBadRequest, &InvalidBoolQueryParamError{
						ParamName: name,
						Value:     raw,
					}, logger)
					return
				}
				val = parsed
			}

			key := boolQueryParamKey{name: name}
			ctx = context.WithValue(ctx, key, val)

			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type requestKey struct{}

func requestFromContext(ctx context.Context) any {
	key := requestKey{}
	req := ctx.Value(key)
	if req == nil {
		panic(errors.New("unmarshalRequestBody middleware not configured"))
	}
	return req
}

func unmarshalRequest[T any]() Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := log.FromContext(ctx)

			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Warn("Error reading request body", slog.Any("error", err))
				writeError(w, http.StatusBadRequest, &InvalidJSONBodyError{}, logger)
				return
			}

			var req T
			if err := json.Unmarshal(body, &req); err != nil {
				logger.Warn("Error unmarshalling request body", slog.Any("error", err))
				writeError(w, http.StatusBadRequest, unmarshalError(err), logger)
				return
			}

			key := requestKey{}
			ctx = context.WithValue(ctx, key, &req)

			ctx, _ = log.NewContext(ctx, logger.With(
				slog.Any("request", req),
			))

			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func unmarshalError(err error) error {
	var syntaxError *json.SyntaxError
	if errors.As(err, &syntaxError) {
		return &InvalidJSONBodyError{
			Err: syntaxError,
		}
	}

	var typeError *json.UnmarshalTypeError
	if errors.As(err, &typeError) {
		return &InvalidJSONBodyError{
			Err: typeError,
		}
	}

	return &InvalidJSONBodyError{}
}

type validator interface {
	Validate() error
}

func validateRequest() Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := log.FromContext(ctx)
			req, ok := requestFromContext(ctx).(validator)
			if !ok {
				panic(errors.New("request type does not implement validator interface"))
			}

			if err := req.Validate(); err != nil {
				logger.Warn("Invalid request", slog.Any("error", err))
				writeError(w, http.StatusBadRequest, err, logger)
				return
			}

			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
