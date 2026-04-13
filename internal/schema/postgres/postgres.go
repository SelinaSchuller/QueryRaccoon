package postgres

import (
	"QueryRaccoon/internal/schema"
	"database/sql"
)

type PostgresInspector struct {
	sql *sql.DB
}

func NewPostgresInspector(sql *sql.DB) *PostgresInspector {
	return &PostgresInspector{sql: sql}
}

func (i *PostgresInspector) GetDatabases() ([]string, error) {
	rows, err := i.sql.Query("SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var db string
		if err := rows.Scan(&db); err != nil {
			return nil, err
		}
		databases = append(databases, db)
	}
	return databases, nil
}

func (i *PostgresInspector) GetSchemas(database string) ([]string, error) {
	rows, err := i.sql.Query("SELECT schema_name FROM information_schema.schemata WHERE catalog_name = $1 AND schema_name NOT LIKE 'pg_%' AND schema_name != 'information_schema' ORDER BY schema_name;", database)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, err
		}
		schemas = append(schemas, schema)
	}
	return schemas, nil
}

func (i *PostgresInspector) GetTables(schemaParam string) ([]string, error) {
	rows, err := i.sql.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = $1 AND table_type = 'BASE TABLE' ORDER BY table_name;", schemaParam)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (i *PostgresInspector) GetViews(schemaParam string) ([]string, error) {
	rows, err := i.sql.Query("SELECT table_name FROM information_schema.views WHERE table_schema = $1 ORDER BY table_name", schemaParam)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		views = append(views, v)
	}
	return views, nil
}

func (i *PostgresInspector) GetColumns(schemaParam, table string) ([]schema.Column, error) {
	rows, err := i.sql.Query("SELECT column_name, data_type, is_nullable, COALESCE(column_default, '') FROM information_schema.columns WHERE table_schema = $1 AND table_name = $2 ORDER BY ordinal_position;", schemaParam, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []schema.Column
	for rows.Next() {
		var column schema.Column
		var isNullable string
		if err := rows.Scan(&column.Name, &column.Type, &isNullable, &column.Default); err != nil {
			return nil, err
		}
		column.Nullable = isNullable == "YES"
		columns = append(columns, column)
	}
	return columns, nil
}
