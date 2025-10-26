package store

import (
	"health-check-app-micro/internal/models"
	"sync"
)

type Store struct {
	mu            sync.Mutex
	Microservices map[string]*models.Microservice
}

func NewStore() *Store {
	return &Store{
		Microservices: make(map[string]*models.Microservice),
	}
}

func (s *Store) RegisterService(service models.Microservice) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Microservices[service.Name] = &service
}

func (s *Store) GetAll() map[string]*models.Microservice {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Microservices
}

func (s *Store) Get(name string) *models.Microservice {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Microservices[name]
}
