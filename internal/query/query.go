package query

import (
	"QueryRaccoon/internal/connections"
	"QueryRaccoon/internal/drivers"
	"fmt"
)

type Service struct {
	manager *connections.Manager
}

func NewService(manager *connections.Manager) *Service {
	return &Service{manager: manager}
}

func (s *Service) Execute(connectionID string, query string) (*drivers.QueryResult, error) {
	connection, ok := s.manager.Get(connectionID)
	if !ok {
		return nil, fmt.Errorf("connection %s not found", connectionID)
	}

	if !connection.Connected {
		return nil, fmt.Errorf("connection %s is not connected", connectionID)
	}

	return connection.Driver.Execute(query)
}
