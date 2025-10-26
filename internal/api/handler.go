package api

import (
	"net/http"
	"time"

	"health-check-app-micro/internal/models"
	"health-check-app-micro/internal/store"
	"health-check-app-micro/pkg/utils"

	"github.com/gin-gonic/gin"
)

func RegisterHandler(storage *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var service models.Microservice
		if err := c.ShouldBindJSON(&service); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
			return
		}
		service.Status = "UNKNOWN"
		service.LastCheck = time.Now().Format(time.RFC3339)
		storage.RegisterService(service)
		utils.LogInfo("✅ Servicio registrado: " + service.Name)
		c.JSON(http.StatusCreated, gin.H{"message": "Microservicio registrado exitosamente"})
	}
}

func HealthAllHandler(storage *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, storage.GetAll())
	}
}

func HealthOneHandler(storage *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		service := storage.Get(name)
		if service == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Microservicio no encontrado"})
			return
		}
		c.JSON(http.StatusOK, service)
	}
}
