package bindings

import (
	"QueryRaccoon/internal/connections"
	"QueryRaccoon/internal/drivers"
)

type ConnectionService struct {
	Manager *connections.Manager
}

func NewConnectionService(manager *connections.Manager) *ConnectionService {
	return &ConnectionService{Manager: manager}
}

func (s *ConnectionService) AddConnection(config drivers.ConnectionConfig) (string, error) {
	return s.Manager.Add(config)
}

func (s *ConnectionService) Connect(id string) error {
	return s.Manager.Connect(id)
}

func (s *ConnectionService) Disconnect(id string) error {
	return s.Manager.Disconnect(id)
}
