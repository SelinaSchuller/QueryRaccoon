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

func (d *MSSQLDriver) GetDB() *sql.DB { return d.sql }

func (d *MSSQLDriver) Disconnect() error {
	return d.sql.Close()
}

func (d *MSSQLDriver) Ping() error {
	return d.sql.Ping()
}

func (d *MSSQLDriver) Execute(query string) (*drivers.QueryResult, error) {
	return drivers.Execute(d.sql, query)
}
