package checker

import (
	"encoding/json"
	"net/http"
	"time"

	"health-check-app-micro/internal/models"
	"health-check-app-micro/internal/notifier"
	"health-check-app-micro/internal/store"
	"health-check-app-micro/pkg/utils"
)

func StartHealthCheckLoop(storage *store.Store) {
	for _, service := range storage.GetAll() {
		go checkHealthWithFrequency(storage, service)
	}
}

func checkHealthWithFrequency(storage *store.Store, service *models.Microservice) {
	frequency := time.Duration(service.Frequency) * time.Second
	if frequency == 0 {
		frequency = 30 * time.Second // default
	}
	
	ticker := time.NewTicker(frequency)
	defer ticker.Stop()
	
	// Ejecutar primera verificación inmediatamente
	checkHealth(storage, service)
	
	for range ticker.C {
		checkHealth(storage, service)
	}
}

// Nueva función para registrar servicios después del loop inicial
func RegisterNewService(storage *store.Store, service *models.Microservice) {
	utils.LogInfo("🆕 Registrando nuevo servicio para monitoreo: " + service.Name)
	go checkHealthWithFrequency(storage, service)
}

func checkHealth(storage *store.Store, service *models.Microservice) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	oldStatus := service.Status
	resp, err := client.Get(service.Endpoint)
	status := "DOWN"
	
	if err == nil && resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			status = "UP"
			// Intentar parsear respuesta JSON para status detallado
			var hs models.HealthStatus
			if json.NewDecoder(resp.Body).Decode(&hs) == nil && hs.Status != "" {
				status = hs.Status
			}
		}
	}
	
	lastCheck := time.Now().Format(time.RFC3339)
	storage.UpdateService(service.Name, status, lastCheck)

	// Notificar cambio de estado
	if oldStatus != status {
		if status == "DOWN" {
			notifier.Notify(service)
			utils.LogError("⚠️ Servicio caído: " + service.Name)
		} else if oldStatus == "DOWN" {
			notifier.NotifyRecovery(service)
			utils.LogInfo("✅ " + service.Name + " recuperado")
		} else {
			utils.LogInfo("🟢 " + service.Name + " está " + status)
		}
	}
}
