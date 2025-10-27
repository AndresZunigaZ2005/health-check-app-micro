package api

import (
	"net/http"
	"strings"
	"time"

	"health-check-app-micro/internal/checker"
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
		
		// Validaciones básicas
		if service.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El nombre es requerido"})
			return
		}
		if service.Endpoint == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El endpoint es requerido"})
			return
		}
		if !strings.HasPrefix(service.Endpoint, "http://") && !strings.HasPrefix(service.Endpoint, "https://") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El endpoint debe comenzar con http:// o https://"})
			return
		}
		if service.Frequency < 10 {
			service.Frequency = 30 // mínimo 10 segundos, default 30
		}
		
		service.Status = "UNKNOWN"
		service.LastCheck = time.Now().Format(time.RFC3339)
		storage.RegisterService(service)
		
		// Iniciar monitoreo del servicio
		checker.RegisterNewService(storage, &service)
		
		utils.LogInfo("✅ Servicio registrado: " + service.Name + " (check cada " + 
			time.Duration(service.Frequency).String() + ")")
		c.JSON(http.StatusCreated, gin.H{
			"message": "Microservicio registrado exitosamente",
			"service": service,
		})
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
