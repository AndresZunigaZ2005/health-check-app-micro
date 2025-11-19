package checker

import (
	"encoding/json"
	"fmt"
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

	// Obtener el servicio actualizado del store para tener el estado correcto
	currentService := storage.Get(service.Name)
	if currentService == nil {
		currentService = service
	}
	oldStatus := currentService.Status

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

	// Obtener el servicio actualizado después de UpdateService para tener los emails correctos
	updatedService := storage.Get(service.Name)
	if updatedService == nil {
		updatedService = service
	}

	// Notificar cambio de estado
	if oldStatus != status {
		if status == "DOWN" {
			// Verificar que el servicio tenga emails configurados
			if len(updatedService.Emails) == 0 || (len(updatedService.Emails) == 1 && updatedService.Emails[0] == "") {
				utils.LogError(fmt.Sprintf("⚠️ Servicio caido: %s, pero no hay emails configurados para notificar", service.Name))
			} else {
				notifier.Notify(updatedService)
				utils.LogError("⚠️ Servicio caido: " + service.Name)
			}
		} else if oldStatus == "DOWN" {
			if len(updatedService.Emails) == 0 || (len(updatedService.Emails) == 1 && updatedService.Emails[0] == "") {
				utils.LogInfo(fmt.Sprintf("✅ %s recuperado, pero no hay emails configurados para notificar", service.Name))
			} else {
				notifier.NotifyRecovery(updatedService)
				utils.LogInfo("✅ " + service.Name + " recuperado")
			}
		} else {
			utils.LogInfo("🟢 " + service.Name + " está " + status)
		}
	}
}
