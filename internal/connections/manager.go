package connections

import (
	"QueryRaccoon/internal/drivers"
	"QueryRaccoon/internal/drivers/mssql"
	"QueryRaccoon/internal/drivers/mysql"
	"QueryRaccoon/internal/drivers/postgres"
	"QueryRaccoon/internal/drivers/sqlite"
	"QueryRaccoon/internal/schema"
	schemamssql "QueryRaccoon/internal/schema/mssql"
	schemamysql "QueryRaccoon/internal/schema/mysql"
	schemapg "QueryRaccoon/internal/schema/postgres"
	schemasqlite "QueryRaccoon/internal/schema/sqlite"
	"fmt"

	"github.com/google/uuid"
)

type Manager struct {
	connections map[uuid.UUID]Connection
}

type Connection struct {
	config    drivers.ConnectionConfig
	Driver    drivers.Driver
	Inspector schema.Inspector
	Connected bool
}

func NewManager() *Manager {
	return &Manager{
		connections: make(map[uuid.UUID]Connection),
	}
}

func (m *Manager) Add(config drivers.ConnectionConfig) (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	var d drivers.Driver
	switch config.DriverType {
	case drivers.PostgreSQL:
		d = &postgres.PostgresDriver{}
	case drivers.MySQL:
		d = &mysql.MySQLDriver{}
	case drivers.SQLite:
		d = &sqlite.SQLiteDriver{}
	case drivers.MSSQL:
		d = &mssql.MSSQLDriver{}
	default:
		return "", fmt.Errorf("unknown driver type: %s", config.DriverType)
	}

	driver := Connection{
		config:    config,
		Driver:    d,
		Connected: false,
	}
	m.connections[uuid] = driver
	return uuid.String(), nil
}

func (m *Manager) Connect(id string) error {
	parsedId, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	conn, ok := m.connections[parsedId]
	if !ok {
		return fmt.Errorf("connection %s not found", id)
	}

	err = conn.Driver.Connect(conn.config)
	if err != nil {
		return err
	}

	if err = conn.Driver.Ping(); err != nil {
		conn.Driver.Disconnect()
		return fmt.Errorf("could not reach database: %w", err)
	}

	db := conn.Driver.GetDB()
	switch conn.config.DriverType {
	case drivers.PostgreSQL:
		conn.Inspector = schemapg.NewPostgresInspector(db)
	case drivers.MySQL:
		conn.Inspector = schemamysql.NewMySQLInspector(db)
	case drivers.SQLite:
		conn.Inspector = schemasqlite.NewSQLiteInspector(db)
	case drivers.MSSQL:
		conn.Inspector = schemamssql.NewMSSQLInspector(db)
	}

	conn.Connected = true
	m.connections[parsedId] = conn
	return nil
}

func (m *Manager) Disconnect(id string) error {
	parsedId, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	conn, ok := m.connections[parsedId]
	if !ok {
		return fmt.Errorf("connection %s not found", id)
	}

	err = conn.Driver.Disconnect()
	if err != nil {
		return err
	}

	conn.Connected = false
	m.connections[parsedId] = conn
	return nil
}

func (m *Manager) Remove(id string) {
	parsedId, err := uuid.Parse(id)
	if err != nil {
		return
	}
	delete(m.connections, parsedId)
}

func (m *Manager) Get(id string) (*Connection, bool) {
	parsedId, err := uuid.Parse(id)
	if err != nil {
		return nil, false
	}
	conn, ok := m.connections[parsedId]
	if !ok {
		return nil, false
	}

	return &conn, true
}
