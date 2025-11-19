package main

import (
	"fmt"
	"health-check-app-micro/internal/api"
	"health-check-app-micro/internal/checker"
	"health-check-app-micro/internal/registry"
	"health-check-app-micro/internal/store"
	"health-check-app-micro/pkg/utils"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	utils.InitLogger()
	utils.LogInfo("🚀 Iniciando microservicio health-check-app-micro...")

	// Cargar variables de entorno desde .env
	if err := godotenv.Load(); err != nil {
		utils.LogInfo("⚠️ No se encontró el archivo .env, se usarán variables del entorno")
	}

	storage := store.NewStore()

	// Registrar servicios automáticamente
	configPath := os.Getenv("SERVICES_CONFIG_PATH")
	if err := registry.AutoRegisterServices(storage, configPath); err != nil {
		utils.LogError("❌ Error en registro automático: " + err.Error())
	}

	// Corregir emails vacíos desde el archivo de configuración
	registry.FixEmptyEmails(storage, configPath)

	go checker.StartHealthCheckLoop(storage) // inicia verificaciones periódicas individuales

	router := api.SetupRouter(storage)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	utils.LogInfo(fmt.Sprintf("🌐 Servidor iniciado en el puerto %s", port))
	router.Run(":" + port)
}
