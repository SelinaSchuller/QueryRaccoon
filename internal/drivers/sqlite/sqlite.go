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

func (d *SQLiteDriver) GetDB() *sql.DB { return d.sql }

func (d *SQLiteDriver) Disconnect() error {
	return d.sql.Close()
}

func (d *SQLiteDriver) Ping() error {
	return d.sql.Ping()
}

func (d *SQLiteDriver) Execute(query string) (*drivers.QueryResult, error) {
	return drivers.Execute(d.sql, query)
}
