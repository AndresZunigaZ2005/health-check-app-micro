package api

import (
	"health-check-app-micro/internal/store"

	"github.com/gin-gonic/gin"
)

func SetupRouter(storage *store.Store) *gin.Engine {
	r := gin.Default()

	r.POST("/register", RegisterHandler(storage))
	r.GET("/health", HealthAllHandler(storage))
	r.GET("/health/:name", HealthOneHandler(storage))

	return r
}
