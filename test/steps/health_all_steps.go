package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HealthAllState struct {
	Res  *http.Response
	Err  error
	Body []byte
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
	url := "http://localhost:8082" + path
	resp, err := http.Get(url)
	s.Res = resp
	s.Err = err

	if err != nil {
		return fmt.Errorf("error al hacer GET a %s: %v", url, err)
	}

	if resp != nil {
		s.Body, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
	} else {
		return fmt.Errorf("respuesta HTTP nula para %s", url)
	}

	return nil
}

func (s *HealthAllState) LaRespuestaDebeSerLista() error {
	if s.Err != nil {
		return fmt.Errorf("error en la petición: %v", s.Err)
	}
	if s.Res == nil {
		return fmt.Errorf("respuesta HTTP nula")
	}

	var list []interface{}
	if err := json.Unmarshal(s.Body, &list); err != nil {
		return fmt.Errorf("la respuesta no es una lista JSON: %v", err)
	}

	if len(list) == 0 {
		return fmt.Errorf("lista de microservicios vacía")
	}

	return nil
}

func (s *HealthAllState) ResponseCodeShouldBe(expected int) error {
	if s.Err != nil {
		return fmt.Errorf("error en la petición: %v", s.Err)
	}
	if s.Res == nil {
		return fmt.Errorf("respuesta HTTP nula")
	}
	if s.Res.StatusCode != expected {
		return fmt.Errorf("status esperado %d, recibido %d", expected, s.Res.StatusCode)
	}
	return nil
}
