package sqlite

import (
	"QueryRaccoon/internal/drivers"
	"database/sql"

	_ "modernc.org/sqlite"
)

type SQLiteDriver struct {
	sql *sql.DB
}

func (d *SQLiteDriver) Connect(config drivers.ConnectionConfig) error {
	db, err := sql.Open("sqlite", config.Database)
	if err != nil {
		return err
	}
	d.sql = db
	return nil
}

func (d *SQLiteDriver) Disconnect() error {
	return d.sql.Close()
}

func (d *SQLiteDriver) Ping() error {
	return d.sql.Ping()
}

func (d *SQLiteDriver) Execute(query string) (*drivers.QueryResult, error) {
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
