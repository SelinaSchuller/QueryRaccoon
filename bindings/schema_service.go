package bindings

import (
	"QueryRaccoon/internal/connections"
	"QueryRaccoon/internal/schema"
	"fmt"
)

type SchemaService struct {
	Manager *connections.Manager
}

func NewSchemaService(manager *connections.Manager) *SchemaService {
	return &SchemaService{Manager: manager}
}

func (s *SchemaService) GetDatabases(connectionID string) ([]string, error) {
	conn, ok := s.Manager.Get(connectionID)
	if !ok {
		return nil, fmt.Errorf("connection %s not found", connectionID)
	}
	return conn.Inspector.GetDatabases()
}

func (s *SchemaService) GetSchemas(connectionID string, database string) ([]string, error) {
	conn, ok := s.Manager.Get(connectionID)
	if !ok {
		return nil, fmt.Errorf("connection %s not found", connectionID)
	}
	return conn.Inspector.GetSchemas(database)
}

func (s *SchemaService) GetTables(connectionID string, schemaName string) ([]string, error) {
	conn, ok := s.Manager.Get(connectionID)
	if !ok {
		return nil, fmt.Errorf("connection %s not found", connectionID)
	}
	return conn.Inspector.GetTables(schemaName)
}

func (s *SchemaService) GetViews(connectionID string, schemaName string) ([]string, error) {
	conn, ok := s.Manager.Get(connectionID)
	if !ok {
		return nil, fmt.Errorf("connection %s not found", connectionID)
	}
	return conn.Inspector.GetViews(schemaName)
}

func (s *SchemaService) GetColumns(connectionID string, schemaName string, table string) ([]schema.Column, error) {
	conn, ok := s.Manager.Get(connectionID)
	if !ok {
		return nil, fmt.Errorf("connection %s not found", connectionID)
	}
	return conn.Inspector.GetColumns(schemaName, table)
}
