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

func (s *ConnectionService) AddConnection(name string, config drivers.ConnectionConfig) (string, error) {
	id, err := s.Manager.Add(name, config)
	if err != nil {
		return "", err
	}
	if err = s.Manager.Connect(id); err != nil {
		s.Manager.Remove(id)
		return "", err
	}
	return id, nil
}

func (s *ConnectionService) GetConnections() []connections.ConnectionInfo {
	return s.Manager.GetAll()
}

func (s *ConnectionService) Connect(id string) error {
	return s.Manager.Connect(id)
}

func (s *ConnectionService) Disconnect(id string) error {
	return s.Manager.Disconnect(id)
}
