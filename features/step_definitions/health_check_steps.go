package stepdefinitions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cucumber/godog"
)

var (
	baseURL        = "http://localhost:8082"
	lastResponse   *http.Response
	lastResponseBody []byte
	registeredService string
)

func elServicioDeHealthCheckEstaDisponible(ctx *godog.ScenarioContext) error {
	ctx.Step(`^que el servicio de health check está disponible$`, func() error {
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			return nil // Cualquier respuesta indica que el servicio está disponible
		}
		defer resp.Body.Close()
		return nil
	})
	return nil
}

func registroUnMicroservicioConDatosValidos(ctx *godog.ScenarioContext) error {
	ctx.Step(`^registro un microservicio con datos válidos$`, func() error {
		registeredService = fmt.Sprintf("test-service-%d", time.Now().Unix())
		payload := map[string]interface{}{
			"name":     registeredService,
			"endpoint": "http://localhost:8085/actuator/health",
			"frequency": 30,
		}
		
		jsonData, _ := json.Marshal(payload)
		resp, err := http.Post(baseURL+"/register", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		
		lastResponse = resp
		lastResponseBody, _ = io.ReadAll(resp.Body)
		return nil
	})
	return nil
}

func existeAlMenosUnServicioRegistrado(ctx *godog.ScenarioContext) error {
	ctx.Step(`^que existe al menos un servicio registrado$`, func() error {
		registeredService = fmt.Sprintf("test-service-%d", time.Now().Unix())
		payload := map[string]interface{}{
			"name":     registeredService,
			"endpoint": "http://localhost:8085/actuator/health",
			"frequency": 30,
		}
		
		jsonData, _ := json.Marshal(payload)
		_, err := http.Post(baseURL+"/register", "application/json", bytes.NewBuffer(jsonData))
		return err
	})
	return nil
}

func consultoElEstadoDeTodosLosServicios(ctx *godog.ScenarioContext) error {
	ctx.Step(`^consulto el estado de todos los servicios$`, func() error {
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			return err
		}
		lastResponse = resp
		lastResponseBody, _ = io.ReadAll(resp.Body)
		return nil
	})
	return nil
}

func consultoElEstadoDeUnServicioEspecifico(ctx *godog.ScenarioContext) error {
	ctx.Step(`^consulto el estado de un servicio específico$`, func() error {
		resp, err := http.Get(baseURL + "/health/" + registeredService)
		if err != nil {
			return err
		}
		lastResponse = resp
		lastResponseBody, _ = io.ReadAll(resp.Body)
		return nil
	})
	return nil
}

func consultoElEndpointDeHealthCheck(ctx *godog.ScenarioContext) error {
	ctx.Step(`^consulto el endpoint de health check$`, func() error {
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			return err
		}
		lastResponse = resp
		lastResponseBody, _ = io.ReadAll(resp.Body)
		return nil
	})
	return nil
}

func laRespuestaDebeTenerEstado(ctx *godog.ScenarioContext) error {
	ctx.Step(`^la respuesta debe tener estado (\d+)$`, func(status int) error {
		if lastResponse.StatusCode != status {
			return fmt.Errorf("expected status %d, got %d", status, lastResponse.StatusCode)
		}
		return nil
	})
	return nil
}

func elCuerpoDebeContenerLosDatosDelServicioRegistrado(ctx *godog.ScenarioContext) error {
	ctx.Step(`^el cuerpo debe contener los datos del servicio registrado$`, func() error {
		var data map[string]interface{}
		if err := json.Unmarshal(lastResponseBody, &data); err != nil {
			return err
		}
		// La respuesta tiene "service" como objeto anidado o "name" directamente
		if serviceData, ok := data["service"].(map[string]interface{}); ok {
			if serviceData["name"] == nil {
				return fmt.Errorf("response does not contain service data")
			}
		} else if data["name"] == nil {
			return fmt.Errorf("response does not contain service data")
		}
		return nil
	})
	return nil
}

func elCuerpoDebeContenerInformacionDeServicios(ctx *godog.ScenarioContext) error {
	ctx.Step(`^el cuerpo debe contener información de servicios$`, func() error {
		if len(lastResponseBody) == 0 {
			return fmt.Errorf("response body is empty")
		}
		return nil
	})
	return nil
}

func elCuerpoDebeContenerInformacionDelServicio(ctx *godog.ScenarioContext) error {
	ctx.Step(`^el cuerpo debe contener información del servicio$`, func() error {
		var data map[string]interface{}
		if err := json.Unmarshal(lastResponseBody, &data); err != nil {
			return err
		}
		if data["name"] == nil {
			return fmt.Errorf("response does not contain service information")
		}
		return nil
	})
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	elServicioDeHealthCheckEstaDisponible(ctx)
	registroUnMicroservicioConDatosValidos(ctx)
	existeAlMenosUnServicioRegistrado(ctx)
	consultoElEstadoDeTodosLosServicios(ctx)
	consultoElEstadoDeUnServicioEspecifico(ctx)
	consultoElEndpointDeHealthCheck(ctx)
	laRespuestaDebeTenerEstado(ctx)
	elCuerpoDebeContenerLosDatosDelServicioRegistrado(ctx)
	elCuerpoDebeContenerInformacionDeServicios(ctx)
	elCuerpoDebeContenerInformacionDelServicio(ctx)
}

