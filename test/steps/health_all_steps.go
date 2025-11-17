package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HealthAllState struct {
	res  *http.Response
	err  error
	body []byte
}

func (s *HealthAllState) HayServiciosRegistrados() error {
	// Registrar un servicio falso para pruebas
	payload := `{
		"name": "dummy-service",
		"endpoint": "http://example.com",
		"frequency": 5,
		"emails": ["test@example.com"]
	}`

	resp, err := http.Post("http://localhost:8082/register", "application/json", bytes.NewBuffer([]byte(payload)))
	if err == nil {
		resp.Body.Close()
	}
	return nil
}

func (s *HealthAllState) HagoGETA(path string) error {
	resp, err := http.Get("http://localhost:8082" + path)
	s.res = resp
	s.err = err

	if resp != nil {
		s.body, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
	}

	return nil
}

func (s *HealthAllState) LaRespuestaDebeSerLista() error {
	if s.err != nil {
		return fmt.Errorf("error en la petición: %v", s.err)
	}
	if s.res == nil {
		return fmt.Errorf("respuesta HTTP nula")
	}

	var list []interface{}
	if err := json.Unmarshal(s.body, &list); err != nil {
		return fmt.Errorf("la respuesta no es una lista JSON: %v", err)
	}

	if len(list) == 0 {
		return fmt.Errorf("lista de microservicios vacía")
	}

	return nil
}
