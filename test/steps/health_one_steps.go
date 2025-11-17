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
	res     *http.Response
	err     error
	body    []byte
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
	resp, err := http.Get("http://localhost:8082" + path)
	s.res = resp
	s.err = err

	if resp != nil {
		s.body, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
	}

	return nil
}

func (s *HealthOneState) BodyContainsName(name string) error {
	if !strings.Contains(string(s.body), name) {
		return fmt.Errorf("la respuesta no contiene '%s': %s", name, string(s.body))
	}
	return nil
}
