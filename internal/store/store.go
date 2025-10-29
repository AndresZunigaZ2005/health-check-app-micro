package store

import (
	"encoding/json"
	"health-check-app-micro/internal/models"
	"os"
	"path/filepath"
	"sync"
)

const defaultStoreFile = "services.json"

type Store struct {
	mu            sync.Mutex
	Microservices map[string]*models.Microservice
	filePath      string
}

// NewStore creates a new store and attempts to load persisted services from disk.
func NewStore() *Store {
	wd, _ := os.Getwd()
	path := filepath.Join(wd, defaultStoreFile)
	return NewStoreWithPath(path)
}

// NewStoreWithPath crea un Store que persiste en la ruta indicada.
// Esto es útil para tests que quieran aislar el almacenamiento en un
// archivo temporal y evitar escribir en el directorio de trabajo del repo.
func NewStoreWithPath(path string) *Store {
	s := &Store{
		Microservices: make(map[string]*models.Microservice),
		filePath:      path,
	}
	// load existing services if file exists
	s.loadFromFile()
	return s
}

func (s *Store) RegisterService(service models.Microservice) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Microservices[service.Name] = &service
	// persist current state (best-effort)
	_ = s.persistLocked()
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

func (s *Store) UpdateService(name string, status string, lastCheck string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if service, exists := s.Microservices[name]; exists {
		service.Status = status
		service.LastCheck = lastCheck
		// persist change
		_ = s.persistLocked()
	}
}

// persistLocked writes the current services to the configured file.
// Caller MUST hold s.mu.
func (s *Store) persistLocked() error {
	list := make([]models.Microservice, 0, len(s.Microservices))
	for _, v := range s.Microservices {
		// copy value to avoid pointing to loop var
		m := *v
		list = append(list, m)
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	// ensure directory exists
	dir := filepath.Dir(s.filePath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// loadFromFile loads persisted services into the store. It acquires the lock
// while populating the internal map.
func (s *Store) loadFromFile() error {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var list []models.Microservice
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ms := range list {
		m := ms
		s.Microservices[m.Name] = &m
	}
	return nil
}
