package mysql

import (
	"QueryRaccoon/internal/drivers"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLDriver struct {
	sql *sql.DB
}

func (d *MySQLDriver) Connect(config drivers.ConnectionConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.User, config.Password, config.Host, config.Port, config.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	d.sql = db
	return nil
}

func (d *MySQLDriver) GetDB() *sql.DB { return d.sql }

func (d *MySQLDriver) Disconnect() error {
	return d.sql.Close()
}

func (d *MySQLDriver) Ping() error {
	return d.sql.Ping()
}

func (d *MySQLDriver) Execute(query string) (*drivers.QueryResult, error) {
	return drivers.Execute(d.sql, query)
}
