package registry

import (
	"encoding/json"
	"fmt"
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
	// Intentar encontrar el archivo de configuración
	foundPath := ""

	// Si se proporciona una ruta específica, verificar si existe
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			foundPath = configPath
		}
	}

	// Si no se encontró, intentar rutas por defecto
	if foundPath == "" {
		// Intentar primero en el directorio de trabajo actual
		testPath := "services-config.json"
		if _, err := os.Stat(testPath); err == nil {
			foundPath = testPath
		} else {
			// Si no existe, intentar en /app (donde se monta en Docker)
			testPath = "/app/services-config.json"
			if _, err2 := os.Stat(testPath); err2 == nil {
				foundPath = testPath
			}
		}
	}

	// Si no se encontró ningún archivo, usar servicios por defecto
	if foundPath == "" {
		utils.LogInfo("⚠️ No se encontró archivo de configuración, usando servicios por defecto")
		return registerDefaultServices(storage)
	}

	// Intentar leer el archivo de configuración
	configData, err := os.ReadFile(foundPath)
	if err != nil {
		// Si no se puede leer el archivo, usar configuración por defecto
		utils.LogInfo(fmt.Sprintf("⚠️ No se pudo leer el archivo de configuración en %s: %v, usando servicios por defecto", foundPath, err))
		return registerDefaultServices(storage)
	}

	utils.LogInfo(fmt.Sprintf("📋 Cargando configuración desde: %s", foundPath))

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

// FixEmptyEmails corrige los emails vacíos de servicios ya registrados
// leyendo desde el archivo de configuración
func FixEmptyEmails(storage *store.Store, configPath string) {
	// Intentar encontrar el archivo de configuración
	foundPath := ""

	// Si se proporciona una ruta específica, verificar si existe
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			foundPath = configPath
		}
	}

	// Si no se encontró, intentar rutas por defecto
	if foundPath == "" {
		// Intentar primero en el directorio de trabajo actual
		testPath := "services-config.json"
		if _, err := os.Stat(testPath); err == nil {
			foundPath = testPath
		} else {
			// Si no existe, intentar en /app (donde se monta en Docker)
			testPath = "/app/services-config.json"
			if _, err2 := os.Stat(testPath); err2 == nil {
				foundPath = testPath
			}
		}
	}

	// Si no se encontró ningún archivo, salir
	if foundPath == "" {
		utils.LogInfo("⚠️ No se encontró archivo de configuración para corregir emails")
		return
	}

	// Intentar leer el archivo de configuración
	configData, err := os.ReadFile(foundPath)
	if err != nil {
		utils.LogInfo(fmt.Sprintf("⚠️ No se pudo leer el archivo de configuración en %s para corregir emails: %v", foundPath, err))
		return
	}

	utils.LogInfo(fmt.Sprintf("📋 Corrigiendo emails desde: %s", foundPath))

	var services []ServiceConfig
	if err := json.Unmarshal(configData, &services); err != nil {
		utils.LogError("❌ Error parseando archivo de configuración para corregir emails: " + err.Error())
		return
	}

	// Corregir emails de cada servicio si están vacíos
	for _, svcConfig := range services {
		currentService := storage.Get(svcConfig.Name)
		if currentService != nil {
			// Verificar si los emails están vacíos o inválidos
			if len(currentService.Emails) == 0 || (len(currentService.Emails) == 1 && currentService.Emails[0] == "") {
				if len(svcConfig.Emails) > 0 && svcConfig.Emails[0] != "" {
					storage.UpdateServiceEmails(svcConfig.Name, svcConfig.Emails)
					utils.LogInfo(fmt.Sprintf("✅ Emails corregidos para servicio %s", svcConfig.Name))
				}
			}
		}
	}
}
