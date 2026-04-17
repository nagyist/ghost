package common

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/util"
)

// DatabaseSchema holds complete schema information for a database
type DatabaseSchema struct {
	ID                string
	Name              string
	Tables            []TableSchema
	Views             []ViewSchema
	MaterializedViews []ViewSchema
	Enums             []EnumSchema
}

// TableSchema holds schema information for a table
type TableSchema struct {
	Name        string
	Columns     []TableColumnSchema
	Constraints []TableConstraint // PK, UK, FK constraints (single and multi-column)
	Indexes     []IndexSchema
	Checks      []CheckConstraint
	Exclusions  []ExclusionConstraint
}

// ViewSchema holds schema information for a view or materialized view
type ViewSchema struct {
	Name    string
	Columns []ViewColumnSchema
	Indexes []IndexSchema // only for materialized views
}

// ViewColumnSchema holds column info for views (simpler than table columns)
type ViewColumnSchema struct {
	Name string
	Type string
}

// TableColumnSchema holds schema information for a table column
type TableColumnSchema struct {
	Name         string
	Type         string
	NotNull      bool
	Default      string // empty if no default
	IsSerial     bool   // true if SERIAL/BIGSERIAL/SMALLSERIAL (has sequence, not identity)
	IdentityType string // 'a' = ALWAYS, 'd' = BY DEFAULT, '' = not identity
}

// TableConstraint describes a constraint (single or multi-column)
type TableConstraint struct {
	Type       ConstraintType
	Name       string
	Columns    []string
	RefTable   string   // for FK
	RefColumns []string // for FK
}

// ConstraintType represents the type of a table constraint
type ConstraintType string

const (
	ConstraintPrimaryKey ConstraintType = "PRIMARY KEY"
	ConstraintUnique     ConstraintType = "UNIQUE"
	ConstraintForeignKey ConstraintType = "FOREIGN KEY"
)

// IndexSchema describes an index
type IndexSchema struct {
	Name        string
	Columns     string // column expressions, e.g. "status" or "created_at DESC"
	IsUnique    bool
	WhereClause string // for partial indexes, empty if not partial
}

// CheckConstraint describes a check constraint
type CheckConstraint struct {
	Name       string
	Columns    []string // columns involved in the check (from conkey)
	Expression string   // full constraint def from pg_get_constraintdef, e.g. "CHECK ((age > 0))"
}

// ExclusionConstraint describes an exclusion constraint
type ExclusionConstraint struct {
	Name       string
	Definition string // full constraint def from pg_get_constraintdef, e.g. "EXCLUDE USING gist (circle WITH &&)"
}

// EnumSchema describes an enum type
type EnumSchema struct {
	Name   string
	Values []string
}

// Row types for scanning query results

type relationColumnRow struct {
	RelationName string  `db:"relation_name"`
	RelationType string  `db:"relation_type"`
	ColumnName   string  `db:"column_name"`
	DataType     string  `db:"data_type"`
	NotNull      bool    `db:"not_null"`
	DefaultValue *string `db:"default_value"`
	ColumnOrder  int16   `db:"column_order"`
	SequenceName *string `db:"sequence_name"`
	IdentityType string  `db:"identity_type"`
}

type constraintRow struct {
	TableName      string   `db:"table_name"`
	ConstraintName string   `db:"constraint_name"`
	ConstraintType string   `db:"constraint_type"`
	Columns        []string `db:"columns"`
	RefTable       *string  `db:"ref_table"`
	RefColumns     []string `db:"ref_columns"`
	ConstraintDef  string   `db:"constraint_def"`
}

type indexRow struct {
	TableName   string  `db:"table_name"`
	IndexName   string  `db:"index_name"`
	IsUnique    bool    `db:"is_unique"`
	ColumnsDef  string  `db:"columns_def"`
	WhereClause *string `db:"where_clause"`
}

type enumRow struct {
	EnumName   string   `db:"enum_name"`
	EnumValues []string `db:"enum_values"`
}

// SQL queries for fetching schema information

const queryRelationsAndColumns = `
SELECT
    c.relname AS relation_name,
    CASE c.relkind
        WHEN 'r' THEN 'table'
        WHEN 'v' THEN 'view'
        WHEN 'm' THEN 'materialized_view'
    END AS relation_type,
    a.attname AS column_name,
    pg_catalog.format_type(a.atttypid, a.atttypmod) AS data_type,
    a.attnotnull AS not_null,
    pg_get_expr(d.adbin, d.adrelid) AS default_value,
    a.attnum AS column_order,
    pg_get_serial_sequence(c.relname, a.attname) AS sequence_name,
    a.attidentity::text AS identity_type
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
JOIN pg_attribute a ON a.attrelid = c.oid
LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE n.nspname = 'public'
  AND c.relkind IN ('r', 'v', 'm')
  AND a.attnum > 0
  AND NOT a.attisdropped
ORDER BY c.relname, a.attnum`

const queryConstraints = `
SELECT
    c.relname AS table_name,
    con.conname AS constraint_name,
    con.contype::text AS constraint_type,
    (
        SELECT array_agg(a.attname ORDER BY x.n)
        FROM unnest(con.conkey) WITH ORDINALITY AS x(key, n)
        JOIN pg_attribute a ON a.attrelid = con.conrelid AND a.attnum = x.key
    ) AS columns,
    confrel.relname AS ref_table,
    (
        SELECT array_agg(a.attname ORDER BY x.n)
        FROM unnest(con.confkey) WITH ORDINALITY AS x(key, n)
        JOIN pg_attribute a ON a.attrelid = con.confrelid AND a.attnum = x.key
    ) AS ref_columns,
    pg_get_constraintdef(con.oid) AS constraint_def
FROM pg_constraint con
JOIN pg_class c ON c.oid = con.conrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
LEFT JOIN pg_class confrel ON confrel.oid = con.confrelid
WHERE n.nspname = 'public'
  AND con.contype IN ('p', 'u', 'f', 'c', 'x')
ORDER BY c.relname, con.contype, con.conname`

const queryIndexes = `
SELECT
    t.relname AS table_name,
    i.relname AS index_name,
    ix.indisunique AS is_unique,
    (
        -- Build column expressions with sort direction from indoption
        -- indoption is an int2vector with bit flags per column:
        -- bit 0 (1) = DESC, bit 1 (2) = NULLS FIRST
        -- Default: ASC NULLS LAST, DESC NULLS FIRST
        -- Show non-default nulls ordering explicitly
        SELECT string_agg(
            pg_get_indexdef(ix.indexrelid, k.n, false) ||
            CASE (ix.indoption[k.n - 1] & 3)
                WHEN 0 THEN ''                      -- ASC NULLS LAST (default)
                WHEN 1 THEN ' DESC NULLS LAST'      -- DESC with non-default nulls
                WHEN 2 THEN ' NULLS FIRST'          -- ASC with non-default nulls
                WHEN 3 THEN ' DESC'                 -- DESC NULLS FIRST (default nulls)
            END,
            ', ' ORDER BY k.n
        )
        FROM generate_series(1, ix.indnkeyatts) AS k(n)
    ) AS columns_def,
    pg_get_expr(ix.indpred, ix.indrelid) AS where_clause
FROM pg_index ix
JOIN pg_class t ON t.oid = ix.indrelid
JOIN pg_class i ON i.oid = ix.indexrelid
JOIN pg_namespace n ON n.oid = t.relnamespace
WHERE n.nspname = 'public'
  AND t.relkind IN ('r', 'm')
  AND NOT EXISTS (
      SELECT 1 FROM pg_constraint con
      WHERE con.conindid = ix.indexrelid
  )
ORDER BY t.relname, i.relname`

const queryEnums = `
SELECT
    t.typname AS enum_name,
    array_agg(e.enumlabel ORDER BY e.enumsortorder) AS enum_values
FROM pg_type t
JOIN pg_namespace n ON n.oid = t.typnamespace
JOIN pg_enum e ON e.enumtypid = t.oid
WHERE n.nspname = 'public'
GROUP BY t.typname
ORDER BY t.typname`

type FetchDatabaseSchemaArgs struct {
	Client      api.ClientWithResponsesInterface
	ProjectID   string
	DatabaseRef string
}

// FetchDatabaseSchema fetches the complete schema information for a database.
func FetchDatabaseSchema(ctx context.Context, args FetchDatabaseSchemaArgs) (*DatabaseSchema, error) {
	// Fetch database info to get name and connection details
	database, err := fetchDatabase(ctx, args.Client, args.ProjectID, args.DatabaseRef)
	if err != nil {
		return nil, err
	}

	// Check if the database is ready to accept connections
	if err := CheckReady(database); err != nil {
		return nil, err
	}

	// Connect to the database
	conn, err := connectToDatabase(ctx, database)
	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())

	schema := &DatabaseSchema{
		ID:   database.Id,
		Name: database.Name,
	}

	// Run queries in sequence (they're fast and we need all of them)
	if err := fetchRelationsAndColumns(ctx, conn, schema); err != nil {
		return nil, fmt.Errorf("failed to fetch relations: %w", err)
	}

	if err := fetchConstraints(ctx, conn, schema); err != nil {
		return nil, fmt.Errorf("failed to fetch constraints: %w", err)
	}

	if err := fetchIndexes(ctx, conn, schema); err != nil {
		return nil, fmt.Errorf("failed to fetch indexes: %w", err)
	}

	if err := fetchEnums(ctx, conn, schema); err != nil {
		return nil, fmt.Errorf("failed to fetch enums: %w", err)
	}

	return schema, nil
}

// connectToDatabase establishes a connection to the given database.
func connectToDatabase(ctx context.Context, database api.Database) (*pgx.Conn, error) {
	const role = "tsdbadmin"

	password, err := GetPassword(database, role)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve password: %w", err)
	}

	connStr, err := BuildConnectionString(ConnectionStringArgs{
		Database: database,
		Role:     role,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build connection string: %w", err)
	}

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return conn, nil
}

func fetchRelationsAndColumns(ctx context.Context, conn *pgx.Conn, schema *DatabaseSchema) error {
	rows, err := conn.Query(ctx, queryRelationsAndColumns)
	if err != nil {
		return err
	}

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[relationColumnRow])
	if err != nil {
		return err
	}

	// Maps to build schema objects
	tables := make(map[string]*TableSchema)
	views := make(map[string]*ViewSchema)
	matViews := make(map[string]*ViewSchema)

	for _, row := range results {
		switch row.RelationType {
		case "table":
			if _, exists := tables[row.RelationName]; !exists {
				tables[row.RelationName] = &TableSchema{Name: row.RelationName}
			}
			col := TableColumnSchema{
				Name:         row.ColumnName,
				Type:         row.DataType,
				NotNull:      row.NotNull,
				Default:      util.DerefStr(row.DefaultValue),
				IsSerial:     row.SequenceName != nil && row.IdentityType == "",
				IdentityType: row.IdentityType,
			}
			tables[row.RelationName].Columns = append(tables[row.RelationName].Columns, col)

		case "view":
			if _, exists := views[row.RelationName]; !exists {
				views[row.RelationName] = &ViewSchema{Name: row.RelationName}
			}
			col := ViewColumnSchema{
				Name: row.ColumnName,
				Type: row.DataType,
			}
			views[row.RelationName].Columns = append(views[row.RelationName].Columns, col)

		case "materialized_view":
			if _, exists := matViews[row.RelationName]; !exists {
				matViews[row.RelationName] = &ViewSchema{Name: row.RelationName}
			}
			col := ViewColumnSchema{
				Name: row.ColumnName,
				Type: row.DataType,
			}
			matViews[row.RelationName].Columns = append(matViews[row.RelationName].Columns, col)
		}
	}

	// Convert maps to slices and sort by name
	for _, t := range tables {
		schema.Tables = append(schema.Tables, *t)
	}
	sort.Slice(schema.Tables, func(i, j int) bool {
		return schema.Tables[i].Name < schema.Tables[j].Name
	})

	for _, v := range views {
		schema.Views = append(schema.Views, *v)
	}
	sort.Slice(schema.Views, func(i, j int) bool {
		return schema.Views[i].Name < schema.Views[j].Name
	})

	for _, mv := range matViews {
		schema.MaterializedViews = append(schema.MaterializedViews, *mv)
	}
	sort.Slice(schema.MaterializedViews, func(i, j int) bool {
		return schema.MaterializedViews[i].Name < schema.MaterializedViews[j].Name
	})

	return nil
}

func fetchConstraints(ctx context.Context, conn *pgx.Conn, schema *DatabaseSchema) error {
	rows, err := conn.Query(ctx, queryConstraints)
	if err != nil {
		return err
	}

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[constraintRow])
	if err != nil {
		return err
	}

	// Build a map for quick table lookup
	tableMap := make(map[string]*TableSchema)
	for i := range schema.Tables {
		tableMap[schema.Tables[i].Name] = &schema.Tables[i]
	}

	for _, row := range results {
		table, exists := tableMap[row.TableName]
		if !exists {
			continue
		}

		switch row.ConstraintType {
		case "p": // primary key
			table.Constraints = append(table.Constraints, TableConstraint{
				Type:    ConstraintPrimaryKey,
				Name:    row.ConstraintName,
				Columns: row.Columns,
			})
		case "u": // unique
			table.Constraints = append(table.Constraints, TableConstraint{
				Type:    ConstraintUnique,
				Name:    row.ConstraintName,
				Columns: row.Columns,
			})
		case "f": // foreign key
			table.Constraints = append(table.Constraints, TableConstraint{
				Type:       ConstraintForeignKey,
				Name:       row.ConstraintName,
				Columns:    row.Columns,
				RefTable:   util.DerefStr(row.RefTable),
				RefColumns: row.RefColumns,
			})
		case "c": // check
			table.Checks = append(table.Checks, CheckConstraint{
				Name:       row.ConstraintName,
				Columns:    row.Columns,
				Expression: row.ConstraintDef,
			})
		case "x": // exclusion
			table.Exclusions = append(table.Exclusions, ExclusionConstraint{
				Name:       row.ConstraintName,
				Definition: row.ConstraintDef,
			})
		}
	}

	return nil
}

func fetchIndexes(ctx context.Context, conn *pgx.Conn, schema *DatabaseSchema) error {
	rows, err := conn.Query(ctx, queryIndexes)
	if err != nil {
		return err
	}

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[indexRow])
	if err != nil {
		return err
	}

	// Build maps for quick lookup
	tableMap := make(map[string]*TableSchema)
	for i := range schema.Tables {
		tableMap[schema.Tables[i].Name] = &schema.Tables[i]
	}
	matViewMap := make(map[string]*ViewSchema)
	for i := range schema.MaterializedViews {
		matViewMap[schema.MaterializedViews[i].Name] = &schema.MaterializedViews[i]
	}

	for _, row := range results {
		idx := IndexSchema{
			Name:        row.IndexName,
			Columns:     row.ColumnsDef,
			IsUnique:    row.IsUnique,
			WhereClause: util.DerefStr(row.WhereClause),
		}

		if table, exists := tableMap[row.TableName]; exists {
			table.Indexes = append(table.Indexes, idx)
		} else if mv, exists := matViewMap[row.TableName]; exists {
			mv.Indexes = append(mv.Indexes, idx)
		}
	}

	return nil
}

func fetchEnums(ctx context.Context, conn *pgx.Conn, schema *DatabaseSchema) error {
	rows, err := conn.Query(ctx, queryEnums)
	if err != nil {
		return err
	}

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[enumRow])
	if err != nil {
		return err
	}

	for _, row := range results {
		schema.Enums = append(schema.Enums, EnumSchema{
			Name:   row.EnumName,
			Values: row.EnumValues,
		})
	}

	return nil
}

// FormatSchema formats a DatabaseSchema into a human-readable string
func FormatSchema(schema *DatabaseSchema) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("DATABASE: %s (%s)\n", schema.Name, schema.ID))

	for _, table := range schema.Tables {
		buf.WriteString(fmt.Sprintf("\nTABLE: %s\n", table.Name))
		formatTableContents(&buf, table)
	}

	for _, view := range schema.Views {
		buf.WriteString(fmt.Sprintf("\nVIEW: %s\n", view.Name))
		formatViewContents(&buf, view)
	}

	for _, mv := range schema.MaterializedViews {
		buf.WriteString(fmt.Sprintf("\nMATERIALIZED VIEW: %s\n", mv.Name))
		formatViewContents(&buf, mv)
	}

	for _, enum := range schema.Enums {
		buf.WriteString(fmt.Sprintf("\nENUM: %s\n", enum.Name))
		formatEnumContents(&buf, enum)
	}

	return buf.String()
}

func formatTableContents(buf *strings.Builder, table TableSchema) {
	// Build single-column constraint lookups for inlining
	// Note: if a column has multiple FKs or CHECKs, show them all separately (not inlined)
	singlePK := ""
	singleUnique := make(map[string]bool)
	singleFK := make(map[string]TableConstraint)
	singleCheck := make(map[string]CheckConstraint)
	fkCount := make(map[string]int)
	checkCount := make(map[string]int)
	var nonInlinedConstraints []TableConstraint
	var nonInlinedChecks []CheckConstraint

	for _, con := range table.Constraints {
		if len(con.Columns) == 1 {
			colName := con.Columns[0]
			switch con.Type {
			case ConstraintPrimaryKey:
				singlePK = colName
			case ConstraintUnique:
				singleUnique[colName] = true
			case ConstraintForeignKey:
				fkCount[colName]++
				if fkCount[colName] == 1 {
					singleFK[colName] = con
				} else {
					// Multiple FKs on same column - show all separately
					if fkCount[colName] == 2 {
						// Add the first FK we stored
						nonInlinedConstraints = append(nonInlinedConstraints, singleFK[colName])
						delete(singleFK, colName)
					}
					nonInlinedConstraints = append(nonInlinedConstraints, con)
				}
			}
		} else {
			nonInlinedConstraints = append(nonInlinedConstraints, con)
		}
	}

	for _, chk := range table.Checks {
		if len(chk.Columns) == 1 {
			colName := chk.Columns[0]
			checkCount[colName]++
			if checkCount[colName] == 1 {
				singleCheck[colName] = chk
			} else {
				// Multiple CHECKs on same column - show all separately
				if checkCount[colName] == 2 {
					// Add the first CHECK we stored
					nonInlinedChecks = append(nonInlinedChecks, singleCheck[colName])
					delete(singleCheck, colName)
				}
				nonInlinedChecks = append(nonInlinedChecks, chk)
			}
		} else {
			// Multi-column or no-column CHECK - show separately
			nonInlinedChecks = append(nonInlinedChecks, chk)
		}
	}

	// Print columns with inline constraints
	maxNameLen := maxColumnNameLength(table.Columns)
	for _, col := range table.Columns {
		isPK := col.Name == singlePK
		isUnique := singleUnique[col.Name]
		fk, hasSingleFK := singleFK[col.Name]
		chk, hasSingleCheck := singleCheck[col.Name]
		line := formatTableColumn(col, maxNameLen, isPK, isUnique, hasSingleFK, fk, hasSingleCheck, chk)
		buf.WriteString(fmt.Sprintf("  %s\n", line))
	}

	// Add blank line before constraints/indexes if there are any
	if len(nonInlinedConstraints) > 0 ||
		len(nonInlinedChecks) > 0 ||
		len(table.Exclusions) > 0 ||
		len(table.Indexes) > 0 {
		buf.WriteString("\n")
	}

	// Print non-inlined constraints (multi-column or multiple on same column)
	for _, con := range nonInlinedConstraints {
		buf.WriteString(fmt.Sprintf("  %s\n", formatConstraint(con)))
	}

	// Print non-inlined check constraints
	for _, chk := range nonInlinedChecks {
		buf.WriteString(fmt.Sprintf("  %s\n", chk.Expression))
	}

	// Print exclusion constraints (always shown separately)
	for _, exc := range table.Exclusions {
		buf.WriteString(fmt.Sprintf("  %s\n", exc.Definition))
	}

	// Print indexes
	for _, idx := range table.Indexes {
		buf.WriteString(fmt.Sprintf("  %s\n", formatIndex(idx)))
	}
}

func formatIndex(idx IndexSchema) string {
	var buf strings.Builder
	if idx.IsUnique {
		buf.WriteString("UNIQUE INDEX ")
	} else {
		buf.WriteString("INDEX ")
	}
	buf.WriteString(idx.Name)
	buf.WriteString(" (")
	buf.WriteString(idx.Columns)
	buf.WriteString(")")
	if idx.WhereClause != "" {
		buf.WriteString(" WHERE ")
		buf.WriteString(idx.WhereClause)
	}
	return buf.String()
}

func formatTableColumn(col TableColumnSchema, width int, isPK, isUnique, hasFK bool, fk TableConstraint, hasCheck bool, chk CheckConstraint) string {
	var parts []string

	// Handle SERIAL and IDENTITY types
	displayType := strings.ToUpper(col.Type)
	showDefault := true
	isAutoGenerated := false

	if col.IdentityType != "" {
		// GENERATED AS IDENTITY column
		isAutoGenerated = true
		showDefault = false
		parts = append(parts, strings.ToUpper(col.Type))
		if col.IdentityType == "a" {
			parts = append(parts, "GENERATED ALWAYS AS IDENTITY")
		} else {
			parts = append(parts, "GENERATED BY DEFAULT AS IDENTITY")
		}
	} else if col.IsSerial {
		// SERIAL/BIGSERIAL/SMALLSERIAL column
		isAutoGenerated = true
		showDefault = false
		switch col.Type {
		case "integer":
			displayType = "SERIAL"
		case "bigint":
			displayType = "BIGSERIAL"
		case "smallint":
			displayType = "SMALLSERIAL"
		}
		parts = append(parts, displayType)
	} else {
		parts = append(parts, displayType)
	}

	if isPK {
		parts = append(parts, "PRIMARY KEY")
	}
	if col.NotNull && !isPK && !isAutoGenerated { // SERIAL and IDENTITY imply NOT NULL
		parts = append(parts, "NOT NULL")
	}
	if isUnique {
		parts = append(parts, "UNIQUE")
	}
	if hasFK {
		parts = append(parts, fmt.Sprintf("REFERENCES %s(%s)", fk.RefTable, fk.RefColumns[0]))
	}
	if showDefault && col.Default != "" {
		parts = append(parts, "DEFAULT "+col.Default)
	}
	if hasCheck {
		parts = append(parts, chk.Expression) // already includes "CHECK (...)"
	}

	return fmt.Sprintf("%-*s  %s", width, col.Name, strings.Join(parts, " "))
}

func formatConstraint(con TableConstraint) string {
	cols := strings.Join(con.Columns, ", ")
	switch con.Type {
	case ConstraintPrimaryKey:
		return fmt.Sprintf("PRIMARY KEY (%s)", cols)
	case ConstraintUnique:
		return fmt.Sprintf("UNIQUE (%s)", cols)
	case ConstraintForeignKey:
		refCols := strings.Join(con.RefColumns, ", ")
		return fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)", cols, con.RefTable, refCols)
	default:
		return ""
	}
}

func formatViewContents(buf *strings.Builder, view ViewSchema) {
	maxNameLen := 0
	for _, col := range view.Columns {
		if len(col.Name) > maxNameLen {
			maxNameLen = len(col.Name)
		}
	}

	for _, col := range view.Columns {
		buf.WriteString(fmt.Sprintf("  %-*s  %s\n", maxNameLen, col.Name, strings.ToUpper(col.Type)))
	}

	// Print indexes (materialized views only)
	if len(view.Indexes) > 0 {
		buf.WriteString("\n")
		for _, idx := range view.Indexes {
			buf.WriteString(fmt.Sprintf("  %s\n", formatIndex(idx)))
		}
	}
}

func formatEnumContents(buf *strings.Builder, enum EnumSchema) {
	values := make([]string, len(enum.Values))
	for i, v := range enum.Values {
		values[i] = fmt.Sprintf("'%s'", v)
	}
	buf.WriteString(fmt.Sprintf("  %s\n", strings.Join(values, ", ")))
}

func maxColumnNameLength(columns []TableColumnSchema) int {
	max := 0
	for _, col := range columns {
		if len(col.Name) > max {
			max = len(col.Name)
		}
	}
	return max
}
