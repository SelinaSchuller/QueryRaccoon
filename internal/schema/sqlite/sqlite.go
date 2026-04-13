package sqlite

import (
	"QueryRaccoon/internal/schema"
	"database/sql"
	"fmt"
)

type SQLiteInspector struct {
	db *sql.DB
}

func NewSQLiteInspector(db *sql.DB) *SQLiteInspector {
	return &SQLiteInspector{db: db}
}

func (i *SQLiteInspector) GetDatabases() ([]string, error) {
	return []string{"main"}, nil
}

func (i *SQLiteInspector) GetSchemas(database string) ([]string, error) {
	return []string{"main"}, nil
}

func (i *SQLiteInspector) GetTables(schemaParam string) ([]string, error) {
	rows, err := i.db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
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

func (i *SQLiteInspector) GetViews(schemaParam string) ([]string, error) {
	rows, err := i.db.Query("SELECT name FROM sqlite_master WHERE type='view' ORDER BY name")
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

func (i *SQLiteInspector) GetColumns(schemaParam, table string) ([]schema.Column, error) {
	rows, err := i.db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []schema.Column
	for rows.Next() {
		var cid int
		var name, colType, dfltValue string
		var notNull, pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return nil, err
		}
		columns = append(columns, schema.Column{
			Name:     name,
			Type:     colType,
			Nullable: notNull == 0,
			Default:  dfltValue,
		})
	}
	return columns, nil
}
