package tests

// Archivo central de pruebas unitarias
// Este fichero agrupa pruebas de varios paquetes (store, api, checker, notifier)
// para que puedan ejecutarse desde un único punto.
//
// Cómo ejecutar:
//   En la raíz del módulo ejecutar: go test ./internal/tests -v
//
// Qué cubre:
// - Persistencia básica del store en `services.json`.
// - Registro vía API y validaciones de entrada.
// - Loop de checking y registro de nuevos servicios.
// - Fallback del notificador cuando no hay SMTP (se escribe en consola).

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"health-check-app-micro/internal/api"
	"health-check-app-micro/internal/checker"
	"health-check-app-micro/internal/models"
	"health-check-app-micro/internal/notifier"
	"health-check-app-micro/internal/store"
)

// para evitar dependencias de disco. Las pruebas ahora se basan en endpoints
// y comportamiento en memoria.

// Comprueba que POST /register registra el servicio y el checker en background actualiza su estado.
func TestAPI_RegisterAndBackgroundCheck(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"name":"ok","status":"UP"}`))
	}))
	defer ts.Close()

	tmp := t.TempDir()
	storage := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	router := api.SetupRouter(storage)

	svc := models.Microservice{
		Name:      "svc-api",
		Endpoint:  ts.URL,
		Frequency: 1,
		Emails:    []string{"x@y.com"},
	}

	body, _ := json.Marshal(svc)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d / body: %s", w.Code, w.Body.String())
	}

	timeout := time.After(3 * time.Second)
	tick := time.Tick(50 * time.Millisecond)
	for {
		select {
		case <-timeout:
			t.Fatalf("timed out waiting for background check; stored: %+v", storage.Get("svc-api"))
		case <-tick:
			s := storage.Get("svc-api")
			if s != nil && s.Status == "UP" {
				return
			}
		}
	}
}

// Verifica que RegisterNewService y el loop de checks actualizan el estado a UP.
func TestChecker_UpdatesStatus(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"name":"svc","status":"UP"}`))
	}))
	defer ts.Close()

	tmp := t.TempDir()
	storage := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	svc := models.Microservice{
		Name:      "svc-checker",
		Endpoint:  ts.URL,
		Frequency: 1,
		Emails:    []string{"a@b"},
		Status:    "UNKNOWN",
	}

	storage.RegisterService(svc)
	checker.RegisterNewService(storage, &svc)

	timeout := time.After(3 * time.Second)
	tick := time.Tick(50 * time.Millisecond)
	for {
		select {
		case <-timeout:
			t.Fatalf("timed out waiting for checker to set UP; got: %+v", storage.Get("svc-checker"))
		case <-tick:
			s := storage.Get("svc-checker")
			if s != nil && s.Status == "UP" {
				return
			}
		}
	}
}

// Nueva prueba: UpdateService actualiza status y lastCheck.
func TestStore_UpdateService(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	s := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	svc := models.Microservice{
		Name:      "to-update",
		Endpoint:  "http://example.local",
		Frequency: 10,
		Emails:    []string{"u@v.com"},
		Status:    "UNKNOWN",
	}
	s.RegisterService(svc)

	s.UpdateService("to-update", "DOWN", "2020-01-01T00:00:00Z")
	got := s.Get("to-update")
	if got == nil {
		t.Fatalf("expected service to exist")
	}
	if got.Status != "DOWN" || got.LastCheck != "2020-01-01T00:00:00Z" {
		t.Fatalf("unexpected update result: %+v", got)
	}
}

// Prueba de validaciones en el handler de registro: campos faltantes o endpoint inválido.
func TestAPI_Register_InvalidCases(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	storage := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	router := api.SetupRouter(storage)

	cases := []struct {
		name string
		body string
		want int
	}{
		{"missing name", `{"endpoint":"http://x","frequency":30}`, http.StatusBadRequest},
		{"bad endpoint", `{"name":"n","endpoint":"ftp://x","frequency":30}`, http.StatusBadRequest},
		{"invalid json", `not-json`, http.StatusBadRequest},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte(c.body)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != c.want {
				t.Fatalf("case %s: expected %d got %d body:%s", c.name, c.want, w.Code, w.Body.String())
			}
		})
	}
}

// Prueba HealthOneHandler devuelve 404 si no existe el servicio.
func TestAPI_HealthOne_NotFound(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	storage := store.NewStoreWithPath(filepath.Join(tmp, "services.json"))
	router := api.SetupRouter(storage)

	req := httptest.NewRequest("GET", "/health/not-exist", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing service, got %d", w.Code)
	}
}

// Prueba del notificador: cuando no hay SMTP configurado, debe usar el log en consola.
func TestNotifier_ConsoleFallback(t *testing.T) {
	t.Parallel()
	// Asegurar que no haya vars SMTP
	os.Unsetenv("SMTP_HOST")
	os.Unsetenv("SMTP_PORT")
	os.Unsetenv("SMTP_USER")
	os.Unsetenv("SMTP_PASSWORD")

	// Capturar la salida del logger global
	var buf strings.Builder
	prev := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prev)

	svc := &models.Microservice{
		Name:      "n1",
		Endpoint:  "http://x",
		Emails:    []string{"a@b"},
		LastCheck: "when",
	}

	notifier.Notify(svc)
	// Debe haberse escrito algo en el buffer (líneas con [INFO])
	out := buf.String()
	if !strings.Contains(out, "[INFO]") && !strings.Contains(out, "[ERROR]") {
		t.Fatalf("esperaba salida por consola en fallback, got: %q", out)
	}
}
