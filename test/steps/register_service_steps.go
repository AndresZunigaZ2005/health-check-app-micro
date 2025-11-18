package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/stretchr/testify/assert"
)

type RegisterState struct {
	Name      string
	Endpoint  string
	Frequency int
	Email     string

	Res  *http.Response
	Err  error
	Body []byte
}

func (s *RegisterState) TengoMicroservicio(name string) error {
	s.Name = name
	return nil
}

func (s *RegisterState) SuEndpoint(ep string) error {
	s.Endpoint = ep
	return nil
}

func (r *RegisterState) SuFrecuencia(freq string) error {
	f, err := strconv.Atoi(freq)
	if err != nil {
		return err
	}
	r.Frequency = f
	return nil
}

func (s *RegisterState) SuEmail(email string) error {
	s.Email = email
	return nil
}

func (s *RegisterState) HagoPOSTA(path string) error {
	data := map[string]interface{}{
		"name":      s.Name,
		"endpoint":  s.Endpoint,
		"frequency": s.Frequency,
		"emails":    []string{s.Email},
	}

	b, _ := json.Marshal(data)
	resp, err := http.Post("http://localhost:8082"+path, "application/json", bytes.NewBuffer(b))

	s.Res = resp
	s.Err = err

	if resp != nil {
		s.Body, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
	}

	return nil
}

func (s *RegisterState) ResponseCodeShouldBe(expected int) error {
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

func (s *RegisterState) BodyShouldContain(text string) error {
	if !assert.Contains(nil, string(s.Body), text) {
		return fmt.Errorf("expected body to contain '%s', got '%s'", text, string(s.Body))
	}
	return nil
}
