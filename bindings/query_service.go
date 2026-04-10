package bindings

import (
	"QueryRaccoon/internal/drivers"
	"QueryRaccoon/internal/query"
)

type QueryService struct {
	Service *query.Service
}

func NewQueryService(service *query.Service) *QueryService {
	return &QueryService{Service: service}
}

func (s *QueryService) Execute(connectionID string, query string) (*drivers.QueryResult, error) {
	return s.Service.Execute(connectionID, query)
}
