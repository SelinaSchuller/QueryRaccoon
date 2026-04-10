package drivers

type Driver interface {
	Connect(ConnectionConfig) error
	Disconnect() error
	Ping() error
	Execute(string) (*QueryResult, error)
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
