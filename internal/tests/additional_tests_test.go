package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"health-check-app-micro/internal/api"
	"health-check-app-micro/internal/models"
	"health-check-app-micro/internal/store"
)

// Test que verifica el endpoint GET /health devuelve todos los servicios
func TestAPI_HealthAll_Success(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	storage := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	router := api.SetupRouter(storage)

	// Registrar algunos servicios
	svc1 := models.Microservice{
		Name:      "service1",
		Endpoint:  "http://example1.com",
		Frequency: 30,
		Emails:    []string{"test1@example.com"},
		Status:    "UP",
	}
	svc2 := models.Microservice{
		Name:      "service2",
		Endpoint:  "http://example2.com",
		Frequency: 60,
		Emails:    []string{"test2@example.com"},
		Status:    "DOWN",
	}
	storage.RegisterService(svc1)
	storage.RegisterService(svc2)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var services []models.Microservice
	if err := json.Unmarshal(w.Body.Bytes(), &services); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(services))
	}
}

// Test que verifica el endpoint GET /health/:name devuelve un servicio específico
func TestAPI_HealthOne_Success(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	storage := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	router := api.SetupRouter(storage)

	svc := models.Microservice{
		Name:      "test-service",
		Endpoint:  "http://example.com",
		Frequency: 30,
		Emails:    []string{"test@example.com"},
		Status:    "UP",
		LastCheck: time.Now().Format(time.RFC3339),
	}
	storage.RegisterService(svc)

	req := httptest.NewRequest("GET", "/health/test-service", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var returned models.Microservice
	if err := json.Unmarshal(w.Body.Bytes(), &returned); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if returned.Name != "test-service" {
		t.Fatalf("expected service name 'test-service', got '%s'", returned.Name)
	}
	if returned.Status != "UP" {
		t.Fatalf("expected status 'UP', got '%s'", returned.Status)
	}
}

// Test que verifica registro con frecuencia menor a 10 usa default de 30
func TestAPI_Register_FrequencyDefault(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	storage := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	router := api.SetupRouter(storage)

	svc := models.Microservice{
		Name:      "low-freq-service",
		Endpoint:  "http://example.com",
		Frequency: 5, // Menor a 10
		Emails:    []string{"test@example.com"},
	}

	body, _ := json.Marshal(svc)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", w.Code)
	}

	stored := storage.Get("low-freq-service")
	if stored == nil {
		t.Fatal("service not stored")
	}
	if stored.Frequency < 10 {
		t.Fatalf("expected frequency >= 10, got %d", stored.Frequency)
	}
}

// Test que verifica que el status inicial es UNKNOWN
func TestAPI_Register_InitialStatusUnknown(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	storage := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	router := api.SetupRouter(storage)

	svc := models.Microservice{
		Name:      "new-service",
		Endpoint:  "http://example.com",
		Frequency: 30,
		Emails:    []string{"test@example.com"},
	}

	body, _ := json.Marshal(svc)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", w.Code)
	}

	stored := storage.Get("new-service")
	if stored == nil {
		t.Fatal("service not stored")
	}
	if stored.Status != "UNKNOWN" {
		t.Fatalf("expected initial status 'UNKNOWN', got '%s'", stored.Status)
	}
}

// Test que verifica que se puede registrar un servicio con HTTPS
func TestAPI_Register_HTTPSEndpoint(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	storage := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	router := api.SetupRouter(storage)

	svc := models.Microservice{
		Name:      "secure-service",
		Endpoint:  "https://secure.example.com",
		Frequency: 30,
		Emails:    []string{"test@example.com"},
	}

	body, _ := json.Marshal(svc)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", w.Code)
	}

	stored := storage.Get("secure-service")
	if stored == nil {
		t.Fatal("service not stored")
	}
	if stored.Endpoint != "https://secure.example.com" {
		t.Fatalf("expected https endpoint, got '%s'", stored.Endpoint)
	}
}

// Test que verifica que no se puede registrar un servicio sin emails
func TestAPI_Register_EmptyEmails(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	storage := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	router := api.SetupRouter(storage)

	svc := models.Microservice{
		Name:      "no-email-service",
		Endpoint:  "http://example.com",
		Frequency: 30,
		Emails:    []string{}, // Vacío
	}

	body, _ := json.Marshal(svc)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// El servicio se registra aunque no tenga emails
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", w.Code)
	}
}

// Test que verifica que GET /health retorna lista vacía cuando no hay servicios
func TestAPI_HealthAll_EmptyList(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	storage := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	router := api.SetupRouter(storage)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var services []models.Microservice
	if err := json.Unmarshal(w.Body.Bytes(), &services); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(services) != 0 {
		t.Fatalf("expected empty list, got %d services", len(services))
	}
}

// Test que verifica actualización de status DOWN
func TestStore_UpdateService_Down(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	s := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	svc := models.Microservice{
		Name:      "down-service",
		Endpoint:  "http://example.com",
		Frequency: 30,
		Emails:    []string{"test@example.com"},
		Status:    "UP",
	}
	s.RegisterService(svc)

	checkTime := time.Now().Format(time.RFC3339)
	s.UpdateService("down-service", "DOWN", checkTime)
	
	got := s.Get("down-service")
	if got == nil {
		t.Fatal("service not found")
	}
	if got.Status != "DOWN" {
		t.Fatalf("expected status 'DOWN', got '%s'", got.Status)
	}
	if got.LastCheck != checkTime {
		t.Fatalf("expected lastCheck '%s', got '%s'", checkTime, got.LastCheck)
	}
}

// Test que verifica GetAll retorna todos los servicios registrados
func TestStore_GetAll_MultipleServices(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	s := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))

	for i := 1; i <= 5; i++ {
		svc := models.Microservice{
			Name:      string(rune(i)) + "-service",
			Endpoint:  "http://example.com",
			Frequency: 30,
			Emails:    []string{"test@example.com"},
		}
		s.RegisterService(svc)
	}

	all := s.GetAll()
	if len(all) != 5 {
		t.Fatalf("expected 5 services, got %d", len(all))
	}
}

// Test que verifica que el checker detecta servicio DOWN
func TestChecker_DetectsDown(t *testing.T) {
	t.Parallel()

	// Servidor que siempre falla
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	tmp := t.TempDir()
	storage := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	svc := models.Microservice{
		Name:      "failing-service",
		Endpoint:  ts.URL,
		Frequency: 1,
		Emails:    []string{"test@example.com"},
		Status:    "UNKNOWN",
	}

	storage.RegisterService(svc)
	checker.RegisterNewService(storage, &svc)

	timeout := time.After(3 * time.Second)
	tick := time.Tick(50 * time.Millisecond)
	for {
		select {
		case <-timeout:
			t.Fatalf("timed out waiting for checker to set DOWN")
		case <-tick:
			s := storage.Get("failing-service")
			if s != nil && s.Status == "DOWN" {
				return
			}
		}
	}
}

// Test que verifica Content-Type de respuestas
func TestAPI_ContentType_JSON(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	storage := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	router := api.SetupRouter(storage)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if !bytes.Contains([]byte(contentType), []byte("application/json")) {
		t.Fatalf("expected JSON content type, got '%s'", contentType)
	}
}

// Test que verifica que un servicio sin nombre no se registra
func TestAPI_Register_EmptyName(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	storage := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	router := api.SetupRouter(storage)

	svc := models.Microservice{
		Name:      "", // Vacío
		Endpoint:  "http://example.com",
		Frequency: 30,
		Emails:    []string{"test@example.com"},
	}

	body, _ := json.Marshal(svc)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", w.Code)
	}
}

// Test que verifica que un servicio sin endpoint no se registra
func TestAPI_Register_EmptyEndpoint(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	storage := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	router := api.SetupRouter(storage)

	svc := models.Microservice{
		Name:      "no-endpoint",
		Endpoint:  "", // Vacío
		Frequency: 30,
		Emails:    []string{"test@example.com"},
	}

	body, _ := json.Marshal(svc)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", w.Code)
	}
}

