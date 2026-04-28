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

func (s *SchemaService) getInspector(connectionID string) (schema.Inspector, error) {
	conn, ok := s.Manager.Get(connectionID)
	if !ok {
		return nil, fmt.Errorf("connection %s not found", connectionID)
	}
	if conn.Inspector == nil {
		return nil, fmt.Errorf("not connected")
	}
	return conn.Inspector, nil
}

func (s *SchemaService) GetDatabases(connectionID string) ([]string, error) {
	inspector, err := s.getInspector(connectionID)
	if err != nil {
		return nil, err
	}
	return inspector.GetDatabases()
}

func (s *SchemaService) GetSchemas(connectionID string, database string) ([]string, error) {
	inspector, err := s.getInspector(connectionID)
	if err != nil {
		return nil, err
	}
	return inspector.GetSchemas(database)
}

func (s *SchemaService) GetTables(connectionID string, schemaName string) ([]string, error) {
	inspector, err := s.getInspector(connectionID)
	if err != nil {
		return nil, err
	}
	return inspector.GetTables(schemaName)
}

func (s *SchemaService) GetViews(connectionID string, schemaName string) ([]string, error) {
	inspector, err := s.getInspector(connectionID)
	if err != nil {
		return nil, err
	}
	return inspector.GetViews(schemaName)
}

func (s *SchemaService) GetColumns(connectionID string, schemaName string, table string) ([]schema.Column, error) {
	inspector, err := s.getInspector(connectionID)
	if err != nil {
		return nil, err
	}
	return inspector.GetColumns(schemaName, table)
}
