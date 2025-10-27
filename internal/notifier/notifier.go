package notifier

import (
	"fmt"
	"health-check-app-micro/internal/models"
	"health-check-app-micro/pkg/utils"
	"net/smtp"
	"os"
)

func Notify(service *models.Microservice) {
	sendNotification(service, fmt.Sprintf("⚠️ ALERTA: El microservicio %s está CAÍDO", service.Name),
		fmt.Sprintf("El microservicio %s está actualmente DOWN.\nEndpoint: %s\nÚltimo check: %s",
			service.Name, service.Endpoint, service.LastCheck))
}

func NotifyRecovery(service *models.Microservice) {
	sendNotification(service, fmt.Sprintf("✅ RECUPERADO: El microservicio %s está UP", service.Name),
		fmt.Sprintf("El microservicio %s ha recuperado su estado normal.\nEndpoint: %s\nÚltimo check: %s",
			service.Name, service.Endpoint, service.LastCheck))
}

func sendNotification(service *models.Microservice, subject, body string) {
	// Obtener configuración SMTP de variables de entorno
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	
	if smtpHost == "" || smtpPort == "" {
		// Si no hay SMTP configurado, solo log en consola
		for _, email := range service.Emails {
			utils.LogInfo(fmt.Sprintf("📧 [CONSOLE] %s -> %s: %s", subject, email, body))
		}
		return
	}
	
	// Implementación real de envío por SMTP con formato RFC 822 correcto
	for _, email := range service.Emails {
		// Formato correcto del mensaje según RFC 822
		msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s\r\n", smtpUser, email, subject, body)
		msgBytes := []byte(msg)
		
		auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
		addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
		
		err := smtp.SendMail(addr, auth, smtpUser, []string{email}, msgBytes)
		if err != nil {
			utils.LogError(fmt.Sprintf("❌ Error enviando email a %s: %v", email, err))
			// No lanza error, solo log para no bloquear el monitoreo
		} else {
			utils.LogInfo(fmt.Sprintf("📧 Email enviado exitosamente a %s sobre %s", email, service.Name))
		}
	}
}
