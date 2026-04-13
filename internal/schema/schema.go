package schema

type Inspector interface {
	GetDatabases() ([]string, error)
	GetSchemas(database string) ([]string, error)
	GetTables(schema string) ([]string, error)
	GetViews(schema string) ([]string, error)
	GetColumns(schema, table string) ([]Column, error)
}

type Column struct {
	Name     string
	Type     string
	Nullable bool
	Default  string
}
