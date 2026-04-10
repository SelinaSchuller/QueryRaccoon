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
	id, err := s.Manager.Add(config)
	if err != nil {
		return "", err
	}
	if err = s.Manager.Connect(id); err != nil {
		s.Manager.Remove(id)
		return "", err
	}
	return id, nil
}

func (s *ConnectionService) Disconnect(id string) error {
	return s.Manager.Disconnect(id)
}
