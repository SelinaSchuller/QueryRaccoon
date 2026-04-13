package mysql

import (
	"QueryRaccoon/internal/schema"
	"database/sql"
	"fmt"
)

type MySQLInspector struct {
	db *sql.DB
}

func NewMySQLInspector(db *sql.DB) *MySQLInspector {
	return &MySQLInspector{db: db}
}

func (i *MySQLInspector) GetDatabases() ([]string, error) {
	rows, err := i.db.Query("SHOW DATABASES")
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

func (i *MySQLInspector) GetSchemas(database string) ([]string, error) {
	return []string{database}, nil
}

func (i *MySQLInspector) GetTables(schemaParam string) ([]string, error) {
	rows, err := i.db.Query(fmt.Sprintf("SHOW TABLES FROM `%s`", schemaParam))
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

func (i *MySQLInspector) GetViews(schemaParam string) ([]string, error) {
	rows, err := i.db.Query("SELECT table_name FROM information_schema.views WHERE table_schema = ? ORDER BY table_name", schemaParam)
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

func (i *MySQLInspector) GetColumns(schemaParam, table string) ([]schema.Column, error) {
	rows, err := i.db.Query("SELECT column_name, data_type, is_nullable, COALESCE(column_default, '') FROM information_schema.columns WHERE table_schema = ? AND table_name = ? ORDER BY ordinal_position", schemaParam, table)
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
