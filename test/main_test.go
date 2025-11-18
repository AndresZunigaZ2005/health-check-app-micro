package test

import (
	"fmt"
	"health-check-app-micro/test/steps"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/cucumber/godog"
)

var srv *exec.Cmd

func TestMain(m *testing.M) {
	// 1. Configurar entorno para pruebas (deshabilitar auto-registro de servicios)
	os.Setenv("SERVICES_CONFIG_PATH", "")
	
	// 2. Levantar servidor en background
	srv = exec.Command("go", "run", "../cmd/main.go")
	srv.Stdout = os.Stdout
	srv.Stderr = os.Stderr
	if err := srv.Start(); err != nil {
		panic("No se pudo iniciar el servidor: " + err.Error())
	}

	// 3. Esperar a que el servidor arranque y verificar que esté listo
	maxRetries := 15
	for i := 0; i < maxRetries; i++ {
		time.Sleep(500 * time.Millisecond)
		resp, err := http.Get("http://localhost:8082/health")
		if err == nil && resp != nil {
			resp.Body.Close()
			// Aceptar cualquier código de respuesta (200, 404, etc.) como indicador de que el servidor está respondiendo
			break
		}
		if i == maxRetries-1 {
			panic("El servidor no está respondiendo después de varios intentos")
		}
	}

	// 4. Ejecutar pruebas
	code := m.Run()

	// 5. Matar servidor
	if srv.Process != nil {
		_ = srv.Process.Kill()
		// Esperar un poco para que el proceso termine
		time.Sleep(500 * time.Millisecond)
	}

	os.Exit(code)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "health-check-suite",
		ScenarioInitializer: InitializeScenarios,
		Options: &godog.Options{
			Format: "pretty",
			Paths:  []string{"features"},
		},
	}

	if suite.Run() != 0 {
		t.Fatal("features failed")
	}
}

func InitializeScenarios(ctx *godog.ScenarioContext) {
	reg := &steps.RegisterState{}
	all := &steps.HealthAllState{}
	one := &steps.HealthOneState{}

	// registrar steps RegisterState
	ctx.Step(`^tengo un microservicio llamado "([^"]*)"$`, reg.TengoMicroservicio)
	ctx.Step(`^su endpoint "([^"]*)"$`, reg.SuEndpoint)
	ctx.Step(`^su frecuencia "(\d+)"$`, reg.SuFrecuencia)
	ctx.Step(`^su frecuencia (\d+)$`, reg.SuFrecuencia)
	ctx.Step(`^su email "([^"]*)"$`, reg.SuEmail)
	ctx.Step(`^hago POST a "([^"]*)"$`, reg.HagoPOSTA)
	ctx.Step(`^la respuesta debe tener código (\d+)$`, func(expected int) error {
		// Determinar qué estado usar basándose en qué estado tiene una respuesta
		if reg.Res != nil {
			return reg.ResponseCodeShouldBe(expected)
		} else if all.Res != nil {
			return all.ResponseCodeShouldBe(expected)
		} else if one.Res != nil {
			return one.ResponseCodeShouldBe(expected)
		}
		return fmt.Errorf("ningún estado tiene una respuesta para validar")
	})
	ctx.Step(`^el cuerpo debe contener "([^"]*)"$`, reg.BodyShouldContain)

	// health all
	ctx.Step(`^hay microservicios registrados$`, all.HayServiciosRegistrados)
	ctx.Step(`^hago GET a "([^"]*)"$`, func(path string) error {
		// Determinar qué estado usar basándose en el path
		if path == "/health" {
			return all.HagoGETA(path)
		}
		// Para paths que contienen un nombre de servicio, usar HealthOneState
		return one.HagoGETOne(path)
	})
	ctx.Step(`^la respuesta debe ser una lista de microservicios$`, all.LaRespuestaDebeSerLista)

	// health one
	ctx.Step(`^existe el microservicio "([^"]*)"$`, one.ExisteMicroservicio)
	ctx.Step(`^el cuerpo debe contener el nombre "([^"]*)"$`, one.BodyContainsName)
}
