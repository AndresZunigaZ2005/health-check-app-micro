package steps

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type HealthOneState struct {
	Service string
	Res     *http.Response
	Err     error
	Body    []byte
}

func (s *HealthOneState) ExisteMicroservicio(name string) error {
	s.Service = name

	// Registrar microservicio automáticamente
	data := fmt.Sprintf(`
	{
		"name": "%s",
		"endpoint": "http://example.com",
		"frequency": 10,
		"emails": ["test@example.com"]
	}`, name)

	resp, err := http.Post("http://localhost:8082/register", "application/json", bytes.NewBuffer([]byte(data)))
	if err == nil {
		resp.Body.Close()
	}

	return nil
}

func (s *HealthOneState) HagoGETOne(path string) error {
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

func (s *HealthOneState) BodyContainsName(name string) error {
	if s.Err != nil {
		return fmt.Errorf("error en la petición: %v", s.Err)
	}
	if s.Res == nil {
		return fmt.Errorf("respuesta HTTP nula")
	}
	if !strings.Contains(string(s.Body), name) {
		return fmt.Errorf("la respuesta no contiene '%s': %s", name, string(s.Body))
	}
	return nil
}

func (s *HealthOneState) ResponseCodeShouldBe(expected int) error {
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
