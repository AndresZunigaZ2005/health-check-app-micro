package test

import (
	"health-check-app-micro/test/steps"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/cucumber/godog"
)

var srv *exec.Cmd

func TestMain(m *testing.M) {
	// 1. Levantar servidor en background
	srv = exec.Command("go", "run", "../cmd/main.go")
	srv.Stdout = os.Stdout
	srv.Stderr = os.Stderr
	if err := srv.Start(); err != nil {
		panic("No se pudo iniciar el servidor: " + err.Error())
	}

	// 2. Esperar a que el servidor arranque
	time.Sleep(2 * time.Second)

	// 3. Ejecutar pruebas
	code := m.Run()

	// 4. Matar servidor
	_ = srv.Process.Kill()

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
	ctx.Step(`^su frecuencia (\d+)$`, reg.SuFrecuencia)
	ctx.Step(`^su email "([^"]*)"$`, reg.SuEmail)
	ctx.Step(`^hago POST a "([^"]*)"$`, reg.HagoPOSTA)
	ctx.Step(`^la respuesta debe tener código (\d+)$`, reg.ResponseCodeShouldBe)
	ctx.Step(`^el cuerpo debe contener "([^"]*)"$`, reg.BodyShouldContain)

	// health all
	ctx.Step(`^hay microservicios registrados$`, all.HayServiciosRegistrados)
	ctx.Step(`^hago GET a "([^"]*)"$`, all.HagoGETA)
	ctx.Step(`^la respuesta debe ser una lista de microservicios$`, all.LaRespuestaDebeSerLista)

	// health one
	ctx.Step(`^existe el microservicio "([^"]*)"$`, one.ExisteMicroservicio)
	ctx.Step(`^hago GET a "([^"]*)"$`, one.HagoGETOne)
	ctx.Step(`^el cuerpo debe contener el nombre "([^"]*)"$`, one.BodyContainsName)
}
