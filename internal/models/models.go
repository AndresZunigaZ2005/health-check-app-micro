package models

import "time"

type ServiceStatus string

const (
	StatusUnknown ServiceStatus = "UNKNOWN"
	StatusUp      ServiceStatus = "UP"
	StatusDown    ServiceStatus = "DOWN"
	StatusAlarm   ServiceStatus = "ALARM"
)

type MonitoredService struct {
	Name        string        `json:"name"`
	Endpoint    string        `json:"endpoint"`
	Frequency   int           `json:"frequency_seconds"` // segundos
	Emails      []string      `json:"emails"`
	LastStatus  ServiceStatus `json:"last_status"`
	LastChecked time.Time     `json:"last_checked"`
}

type RegisterRequest struct {
	Name      string   `json:"name"`
	Endpoint  string   `json:"endpoint"`
	Frequency int      `json:"frequency_seconds"`
	Emails    []string `json:"emails"`
}
