package mssql

import (
	"QueryRaccoon/internal/schema"
	"database/sql"
)

type MSSQLInspector struct {
	db *sql.DB
}

func NewMSSQLInspector(db *sql.DB) *MSSQLInspector {
	return &MSSQLInspector{db: db}
}

func (i *MSSQLInspector) GetDatabases() ([]string, error) {
	rows, err := i.db.Query("SELECT name FROM sys.databases ORDER BY name")
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

func (i *MSSQLInspector) GetSchemas(database string) ([]string, error) {
	rows, err := i.db.Query("SELECT DISTINCT TABLE_SCHEMA FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_CATALOG = @p1 ORDER BY TABLE_SCHEMA", database)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		schemas = append(schemas, s)
	}
	return schemas, nil
}

func (i *MSSQLInspector) GetTables(schemaParam string) ([]string, error) {
	rows, err := i.db.Query("SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = @p1 AND TABLE_TYPE = 'BASE TABLE' ORDER BY TABLE_NAME", schemaParam)
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

func (i *MSSQLInspector) GetViews(schemaParam string) ([]string, error) {
	rows, err := i.db.Query("SELECT TABLE_NAME FROM INFORMATION_SCHEMA.VIEWS WHERE TABLE_SCHEMA = @p1 ORDER BY TABLE_NAME", schemaParam)
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

func (i *MSSQLInspector) GetColumns(schemaParam, table string) ([]schema.Column, error) {
	rows, err := i.db.Query("SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COALESCE(COLUMN_DEFAULT, '') FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = @p1 AND TABLE_NAME = @p2 ORDER BY ORDINAL_POSITION", schemaParam, table)
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
