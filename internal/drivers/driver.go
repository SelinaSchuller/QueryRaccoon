package drivers

import (
	"database/sql"
	"strings"
)

type Driver interface {
	Connect(ConnectionConfig) error
	Disconnect() error
	Ping() error
	Execute(string) (*QueryResult, error)
	GetDB() *sql.DB
}

type DriverType string

const (
	PostgreSQL DriverType = "postgresql"
	MSSQL      DriverType = "mssql"
	MySQL      DriverType = "mysql"
	SQLite     DriverType = "sqlite"
)

type ConnectionConfig struct {
	Host       string
	Port       int
	User       string
	Password   string
	DriverType DriverType
	Database   string
}

type QueryResult struct {
	Columns []string
	Rows    [][]any
}

// Execute routes SELECT-like queries through Query() and DML through Exec().
// Using Query() for UPDATE/INSERT/DELETE can cause drivers (especially MSSQL)
// to hang waiting for result sets that never come.
func Execute(db *sql.DB, query string) (*QueryResult, error) {
	first := strings.ToUpper(strings.Fields(strings.TrimSpace(query))[0])
	switch first {
	case "SELECT", "WITH", "SHOW", "EXPLAIN", "DESCRIBE", "DESC", "PRAGMA", "EXEC", "CALL":
		return queryRows(db, query)
	default:
		return execStatement(db, query)
	}
}

func queryRows(db *sql.DB, query string) (*QueryResult, error) {
	var res QueryResult
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res.Columns, err = rows.Columns()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		cols := make([]interface{}, len(res.Columns))
		pointers := make([]interface{}, len(res.Columns))
		for i := range cols {
			pointers[i] = &cols[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		res.Rows = append(res.Rows, cols)
	}
	return &res, nil
}

func execStatement(db *sql.DB, query string) (*QueryResult, error) {
	result, err := db.Exec(query)
	if err != nil {
		return nil, err
	}
	affected, _ := result.RowsAffected()
	return &QueryResult{
		Columns: []string{"rows affected"},
		Rows:    [][]any{{affected}},
	}, nil
}
