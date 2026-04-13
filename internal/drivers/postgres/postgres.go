package postgres

import (
	"QueryRaccoon/internal/drivers"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresDriver struct {
	sql *sql.DB
}

func (d *PostgresDriver) Connect(config drivers.ConnectionConfig) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.Database)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	d.sql = db
	return nil
}

func (d *PostgresDriver) GetDB() *sql.DB { return d.sql }

func (d *PostgresDriver) Disconnect() error {
	return d.sql.Close()
}

func (d *PostgresDriver) Ping() error {
	return d.sql.Ping()
}

func (d *PostgresDriver) Execute(query string) (*drivers.QueryResult, error) {
	return drivers.Execute(d.sql, query)
}
