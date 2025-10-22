package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/example/monitor/internal/models"
	"github.com/example/monitor/internal/store"
)

type API struct {
	Store *store.Store
}

func NewAPI(s *store.Store) *API { return &API{Store: s} }

// RegisterHandler registra o actualiza un servicio a monitorear.
// Request JSON:
//
//	{
//	  "name":"svc-a",
//	  "endpoint":"http://svc-a:9001/health",
//	  "frequency_seconds":10,
//	  "emails":["you@example.com"]
//	}
func (a *API) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Endpoint == "" || req.Frequency <= 0 {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	ms := &models.MonitoredService{
		Name:      req.Name,
		Endpoint:  req.Endpoint,
		Frequency: req.Frequency,
		Emails:    req.Emails,
		// LastStatus y LastChecked serán manejados por el checker
		LastStatus:  models.StatusUnknown,
		LastChecked: time.Time{},
	}

	if err := a.Store.AddOrUpdate(ms); err != nil {
		http.Error(w, "failed to persist service", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"result": "registered"})
}

// GetAllHealth devuelve el estado actual (según lo guardado en store) de todos los servicios.
func (a *API) GetAllHealth(w http.ResponseWriter, r *http.Request) {
	services, err := a.Store.GetAll()
	if err != nil {
		http.Error(w, "failed to read services", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// GetOneHealth devuelve el estado de un servicio por nombre (ruta /health/{name})
func (a *API) GetOneHealth(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/health/")
	if name == "" {
		http.Error(w, "missing service name", http.StatusBadRequest)
		return
	}
	ms, err := a.Store.Get(name)
	if err != nil {
		http.Error(w, "failed to read service", http.StatusInternalServerError)
		return
	}
	if ms == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ms)
}
