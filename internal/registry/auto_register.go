package registry

import (
	"encoding/json"
	"health-check-app-micro/internal/checker"
	"health-check-app-micro/internal/models"
	"health-check-app-micro/internal/store"
	"health-check-app-micro/pkg/utils"
	"os"
	"time"
)

// ServiceConfig representa la configuración de un servicio para registro automático
type ServiceConfig struct {
	Name      string   `json:"name"`
	Endpoint  string   `json:"endpoint"`
	Frequency int      `json:"frequency"`
	Emails    []string `json:"emails"`
}

// AutoRegisterServices registra automáticamente los servicios definidos en el archivo de configuración
func AutoRegisterServices(storage *store.Store, configPath string) error {
	// Si no hay archivo de configuración, usar servicios por defecto
	if configPath == "" {
		configPath = "services-config.json"
	}

	// Intentar leer el archivo de configuración
	configData, err := os.ReadFile(configPath)
	if err != nil {
		// Si no existe el archivo, usar configuración por defecto
		utils.LogInfo("⚠️ No se encontró archivo de configuración, usando servicios por defecto")
		return registerDefaultServices(storage)
	}

	var services []ServiceConfig
	if err := json.Unmarshal(configData, &services); err != nil {
		utils.LogError("❌ Error parseando archivo de configuración: " + err.Error())
		return registerDefaultServices(storage)
	}

	// Registrar cada servicio
	for _, svcConfig := range services {
		service := models.Microservice{
			Name:      svcConfig.Name,
			Endpoint:  svcConfig.Endpoint,
			Frequency: svcConfig.Frequency,
			Emails:    svcConfig.Emails,
			Status:    "UNKNOWN",
			LastCheck: time.Now().Format(time.RFC3339),
		}

		// Validar frecuencia mínima
		if service.Frequency < 10 {
			service.Frequency = 30
		}

		storage.RegisterService(service)
		checker.RegisterNewService(storage, &service)
		utils.LogInfo("✅ Servicio auto-registrado: " + service.Name)
	}

	return nil
}

// registerDefaultServices registra los servicios por defecto del sistema
func registerDefaultServices(storage *store.Store) error {
	defaultServices := []ServiceConfig{
		{
			Name:      "api-gateway",
			Endpoint:  "http://api-gateway:8085/actuator/health",
			Frequency: 30,
			Emails:    []string{os.Getenv("SMTP_TO")},
		},
		{
			Name:      "gestion-perfil",
			Endpoint:  "http://gestion-perfil:8084/actuator/health",
			Frequency: 30,
			Emails:    []string{os.Getenv("SMTP_TO")},
		},
		{
			Name:      "jwt-service",
			Endpoint:  "http://jwt-service:8081/v1/health",
			Frequency: 30,
			Emails:    []string{os.Getenv("SMTP_TO")},
		},
		{
			Name:      "notifications-service",
			Endpoint:  "http://notifications-service-micro:8080/health",
			Frequency: 30,
			Emails:    []string{os.Getenv("SMTP_TO")},
		},
		{
			Name:      "orquestador-solicitudes",
			Endpoint:  "http://orquestador-solicitudes-micro:3001/health",
			Frequency: 30,
			Emails:    []string{os.Getenv("SMTP_TO")},
		},
	}

	for _, svcConfig := range defaultServices {
		service := models.Microservice{
			Name:      svcConfig.Name,
			Endpoint:  svcConfig.Endpoint,
			Frequency: svcConfig.Frequency,
			Emails:    svcConfig.Emails,
			Status:    "UNKNOWN",
			LastCheck: time.Now().Format(time.RFC3339),
		}

		storage.RegisterService(service)
		checker.RegisterNewService(storage, &service)
		utils.LogInfo("✅ Servicio por defecto registrado: " + service.Name)
	}

	return nil
}

