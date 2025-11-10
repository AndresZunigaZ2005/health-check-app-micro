package features

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"health-check-app-micro/features/step_definitions"
)

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "pretty",
	Paths:  []string{"../features"},
}

func TestFeatures(t *testing.T) {
	status := godog.TestSuite{
		Name:                 "health-check-acceptance",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options:              &opts,
	}.Run()

	if status != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		// Setup antes de ejecutar todas las features
	})
	
	ctx.AfterSuite(func() {
		// Cleanup después de ejecutar todas las features
	})
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	stepdefinitions.InitializeScenario(ctx)
}

