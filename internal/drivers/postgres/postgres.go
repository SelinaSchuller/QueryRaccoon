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

func (d *PostgresDriver) Disconnect() error {
	return d.sql.Close()
}

func (d *PostgresDriver) Ping() error {
	return d.sql.Ping()
}

func (d *PostgresDriver) Execute(query string) (*drivers.QueryResult, error) {
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
