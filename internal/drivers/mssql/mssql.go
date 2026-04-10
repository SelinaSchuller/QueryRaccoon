package mssql

import (
	"QueryRaccoon/internal/drivers"
	"database/sql"
	"fmt"

	_ "github.com/microsoft/go-mssqldb"
)

type MSSQLDriver struct {
	sql *sql.DB
}

func (d *MSSQLDriver) Connect(config drivers.ConnectionConfig) error {
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		config.User, config.Password, config.Host, config.Port, config.Database)
	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return err
	}
	d.sql = db
	return nil
}

func (d *MSSQLDriver) Disconnect() error {
	return d.sql.Close()
}

func (d *MSSQLDriver) Ping() error {
	return d.sql.Ping()
}

func (d *MSSQLDriver) Execute(query string) (*drivers.QueryResult, error) {
	var res drivers.QueryResult
	rows, err := d.sql.Query(query)
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
