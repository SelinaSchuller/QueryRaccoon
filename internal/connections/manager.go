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
	"QueryRaccoon/internal/storage"
	"fmt"

	"github.com/google/uuid"
)

type Manager struct {
	connections map[uuid.UUID]Connection
}

type Connection struct {
	Name      string
	config    drivers.ConnectionConfig
	Driver    drivers.Driver
	Inspector schema.Inspector
	Connected bool
}

type ConnectionInfo struct {
	ID        string
	Name      string
	Config    drivers.ConnectionConfig
	Connected bool
}

func NewManager() *Manager {
	m := &Manager{
		connections: make(map[uuid.UUID]Connection),
	}

	saved, err := storage.Load()
	if err == nil && len(saved) > 0 {
		m.restore(saved)
	} else if err == nil && len(saved) == 0 {
		if seed, seedErr := storage.LoadDevSeed(); seedErr == nil && len(seed) > 0 {
			m.restore(seed)
			_ = m.save()
		}
	}

	return m
}

func (m *Manager) restore(saved []storage.SavedConnection) {
	for _, s := range saved {
		id, err := uuid.Parse(s.ID)
		if err != nil {
			continue
		}
		d := driverFor(s.Config.DriverType)
		if d == nil {
			continue
		}
		m.connections[id] = Connection{
			Name:      s.Name,
			config:    s.Config,
			Driver:    d,
			Connected: false,
		}
	}
}

func (m *Manager) save() error {
	saved := make([]storage.SavedConnection, 0, len(m.connections))
	for id, conn := range m.connections {
		saved = append(saved, storage.SavedConnection{
			ID:     id.String(),
			Name:   conn.Name,
			Config: conn.config,
		})
	}
	return storage.Save(saved)
}

func driverFor(dt drivers.DriverType) drivers.Driver {
	switch dt {
	case drivers.PostgreSQL:
		return &postgres.PostgresDriver{}
	case drivers.MySQL:
		return &mysql.MySQLDriver{}
	case drivers.SQLite:
		return &sqlite.SQLiteDriver{}
	case drivers.MSSQL:
		return &mssql.MSSQLDriver{}
	default:
		return nil
	}
}

func (m *Manager) Add(name string, config drivers.ConnectionConfig) (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	d := driverFor(config.DriverType)
	if d == nil {
		return "", fmt.Errorf("unknown driver type: %s", config.DriverType)
	}

	m.connections[id] = Connection{
		Name:      name,
		config:    config,
		Driver:    d,
		Connected: false,
	}
	_ = m.save()
	return id.String(), nil
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

	if err = conn.Driver.Connect(conn.config); err != nil {
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

	if err = conn.Driver.Disconnect(); err != nil {
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
	_ = m.save()
}

func (m *Manager) Update(id, name string, config drivers.ConnectionConfig) error {
	parsedId, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	conn, ok := m.connections[parsedId]
	if !ok {
		return fmt.Errorf("connection %s not found", id)
	}
	if conn.Connected {
		_ = conn.Driver.Disconnect()
	}
	d := driverFor(config.DriverType)
	if d == nil {
		return fmt.Errorf("unknown driver type: %s", config.DriverType)
	}
	m.connections[parsedId] = Connection{
		Name:   name,
		config: config,
		Driver: d,
	}
	_ = m.save()
	return nil
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

func (m *Manager) GetAll() []ConnectionInfo {
	result := make([]ConnectionInfo, 0, len(m.connections))
	for id, conn := range m.connections {
		result = append(result, ConnectionInfo{
			ID:        id.String(),
			Name:      conn.Name,
			Config:    conn.config,
			Connected: conn.Connected,
		})
	}
	return result
}
