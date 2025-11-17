package main

import (
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

	go checker.StartHealthCheckLoop(storage) // inicia verificaciones periódicas individuales

	router := api.SetupRouter(storage)
	utils.LogInfo("🌐 Servidor iniciado en el puerto 8080")
	router.Run(":8082")
}
