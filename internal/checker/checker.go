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

func StartHealthCheckLoop(storage *store.Store, interval time.Duration) {
	for {
		for _, service := range storage.GetAll() {
			go checkHealth(storage, service)
		}
		time.Sleep(interval)
	}
}

func checkHealth(storage *store.Store, service *models.Microservice) {
	resp, err := http.Get(service.Endpoint)
	status := "DOWN"
	if err == nil && resp.StatusCode == 200 {
		status = "UP"
	}
	service.Status = status
	service.LastCheck = time.Now().Format(time.RFC3339)

	if status == "DOWN" {
		notifier.Notify(service)
		utils.LogError("⚠️ Servicio caído: " + service.Name)
	} else {
		utils.LogInfo("🟢 " + service.Name + " está " + status)
	}

	// parse JSON si el endpoint /health devuelve más info
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	var hs models.HealthStatus
	json.NewDecoder(resp.Body).Decode(&hs)
}
