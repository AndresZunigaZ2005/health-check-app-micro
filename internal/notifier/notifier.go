package notifier

import (
	"fmt"
	"health-check-app-micro/internal/models"
)

func Notify(service *models.Microservice) {
	for _, email := range service.Emails {
		fmt.Printf("📧 Enviando notificación a %s: servicio %s está %s\n", email, service.Name, service.Status)
	}
}
