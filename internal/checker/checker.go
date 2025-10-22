package checker

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/example/monitor/internal/models"
	"github.com/example/monitor/internal/notifier"
	"github.com/example/monitor/internal/store"
)

// Start arranca el proceso que sincroniza servicios desde el store
// y lanza un worker por cada servicio para chequear su health periódicamente.
func Start(ctx context.Context, st *store.Store, n *notifier.Notifier) {
	go func() {
		workers := map[string]context.CancelFunc{}
		for {
			select {
			case <-ctx.Done():
				// cancelar todos los workers
				for _, cancel := range workers {
					cancel()
				}
				return
			case <-time.After(2 * time.Second):
				services, err := st.GetAll()
				if err != nil {
					log.Println("checker: error getting services:", err)
					continue
				}

				// lanzar worker para servicios nuevos
				for _, s := range services {
					if _, ok := workers[s.Name]; !ok {
						wctx, cancel := context.WithCancel(ctx)
						workers[s.Name] = cancel
						go serviceWorker(wctx, s, st, n)
					}
				}

				// cancelar workers para servicios eliminados
				for name, cancel := range workers {
					found := false
					for _, s := range services {
						if s.Name == name {
							found = true
							break
						}
					}
					if !found {
						cancel()
						delete(workers, name)
					}
				}
			}
		}
	}()
}

func serviceWorker(ctx context.Context, s *models.MonitoredService, st *store.Store, n *notifier.Notifier) {
	// calcular frecuencia
	freq := time.Duration(s.Frequency) * time.Second
	if freq <= 0 {
		freq = 30 * time.Second
	}
	ticker := time.NewTicker(freq)
	defer ticker.Stop()

	// primer check inmediato
	performCheck(s, st, n)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			performCheck(s, st, n)
		}
	}
}

func performCheck(s *models.MonitoredService, st *store.Store, n *notifier.Notifier) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(s.Endpoint)

	newStatus := models.StatusUnknown
	if err != nil {
		newStatus = models.StatusDown
	} else {
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			newStatus = models.StatusUp
		} else if resp.StatusCode >= 500 {
			newStatus = models.StatusAlarm
		} else {
			// 3xx/4xx tratarlos como ALARM según política (puedes ajustarlo)
			newStatus = models.StatusAlarm
		}
	}

	// Persistir nuevo estado
	if err := st.UpdateStatus(s.Name, newStatus, time.Now()); err != nil {
		log.Printf("checker: error updating status for %s: %v", s.Name, err)
	}

	// Recuperar el registro actual para comparar estado anterior.
	cur, err := st.Get(s.Name)
	if err != nil {
		log.Printf("checker: error fetching service %s after update: %v", s.Name, err)
		return
	}
	// Si cur es nil, nada que comparar
	if cur == nil {
		log.Printf("checker: service %s not found in store after update", s.Name)
		return
	}

	// Si cambió el estado y es DOWN o ALARM -> notificar
	if cur.LastStatus != newStatus {
		// Notificar solo en caídas o alarmas; si quieres notificar también en UP, añade condición.
		if newStatus == models.StatusDown || newStatus == models.StatusAlarm {
			// Construir subject/body sencillos
			subject := "ALERTA: " + s.Name + " -> " + string(newStatus)
			body := "Servicio: " + s.Name + "\nEndpoint: " + s.Endpoint + "\nEstado: " + string(newStatus) + "\nTiempo: " + time.Now().Format(time.RFC3339)
			n.Notify(s.Emails, subject, body)
		}
	}

	log.Printf("checker: checked %s -> %s", s.Name, newStatus)
}
