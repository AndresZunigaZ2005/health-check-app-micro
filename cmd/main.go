package main

import (
	"health-check-app-micro/internal/api"
	"health-check-app-micro/internal/checker"
	"health-check-app-micro/internal/store"
	"health-check-app-micro/pkg/utils"
)

func main() {
	utils.InitLogger()
	utils.LogInfo("🚀 Iniciando microservicio health-check-app-micro...")

	storage := store.NewStore()
	go checker.StartHealthCheckLoop(storage) // inicia verificaciones periódicas individuales

	router := api.SetupRouter(storage)
	utils.LogInfo("🌐 Servidor iniciado en el puerto 8080")
	router.Run(":8080")
}
