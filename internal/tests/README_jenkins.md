# Integración y ejecución de pruebas en Jenkins

Resumen rápido

Este documento explica cómo automatizar las pruebas unitarias y de integración ligera del proyecto
en Jenkins. Las pruebas del proyecto usan el paquete estándar de Go (`testing`) y servidores HTTP de
prueba (`httptest`). El repositorio ya incluye tests aislados (almacenamiento en archivos temporales),
por lo que la mayor parte del trabajo se realiza en la configuración del job/pipeline de Jenkins.

En pocas palabras:
- En el agente Jenkins hace falta instalar Go y disponer de acceso al repo.
- Ejecutar `go mod download` y luego `go test` (o usar `gotestsum` para generar JUnit XML).
- Publicar resultados JUnit en Jenkins y opcionalmente archivar cobertura.

Requisitos previos en el agente Jenkins

- Go (compatible con la versión indicada en `go.mod`).
- Acceso al repositorio (checkout) y permisos para ejecutar comandos.
- (Opcional, recomendado) `gotestsum` para generar JUnit XML legible por Jenkins:
  - Instalar con: `go install gotest.tools/gotestsum@latest`.
  - Asegúrate de que `$(go env GOPATH)/bin` esté en el PATH del agente (para poder ejecutar `gotestsum`).
- (Opcional) `go-junit-report` es una alternativa para convertir `go test` -> JUnit XML.

Resumen de comandos útiles

- Ejecutar todas las pruebas del paquete central de tests:
```powershell
cd C:\path\to\health-check-app-micro
go test ./internal/tests -v
```

- Ejecutar todas las pruebas del módulo (todos los paquetes):
```powershell
go test ./... -v
```

- Ejecutar una prueba concreta por nombre (regexp):
```powershell
go test ./internal/tests -run "^TestAPI_RegisterAndBackgroundCheck$" -v
```

- Ejecutar con detector de race:
```powershell
go test ./internal/tests -race -v
```

- Generar cobertura:
```powershell
go test ./... -coverprofile=cover.out -v
go tool cover -html=cover.out -o cover.html
```

- Generar JUnit XML con `gotestsum` (recomendado para Jenkins):
```powershell
go install gotest.tools/gotestsum@latest
#$env:PATH = "$(go env GOPATH)\bin;" + $env:PATH  # asegurarse de que esté en PATH
gotestsum --junitfile junit-report.xml -- -v ./...
```

Pipeline recomendado (Jenkinsfile) — ejemplo para agentes Linux

Incluye un pipeline declarativo mínimo que descarga dependencias, ejecuta tests con `gotestsum`,
genera cobertura y publica resultados JUnit.

```groovy
pipeline {
  agent { label 'linux && go' }
  environment {
    GIN_MODE = 'release' // reduce el logging de Gin durante tests
    PATH = "${env.PATH}:${tool 'GOBIN' ?: ''}" // ajusta según tu instalación
  }
  stages {
    stage('Checkout') {
      steps { checkout scm }
    }
    stage('Setup') {
      steps {
        sh 'go version'
        sh 'go mod download'
        sh 'go install gotest.tools/gotestsum@latest'
        sh 'export PATH=$(go env GOPATH)/bin:$PATH'
      }
    }
    stage('Test') {
      steps {
        sh '''
          export GIN_MODE=release
          gotestsum --junitfile junit-report.xml -- -coverprofile=cover.out -v ./...
          go tool cover -html=cover.out -o cover.html || true
        '''
      }
      post {
        always {
          junit 'junit-report.xml'
          archiveArtifacts artifacts: 'cover.out,cover.html,junit-report.xml', fingerprint: true
        }
      }
    }
  }
}
```

Pipeline recomendado (Jenkinsfile) — ejemplo para agentes Windows (PowerShell)

```groovy
pipeline {
  agent { label 'windows && go' }
  stages {
    stage('Checkout') { steps { checkout scm } }
    stage('Setup') {
      steps {
        powershell 'go version'
        powershell 'go mod download'
        powershell 'go install gotest.tools/gotestsum@latest'
        powershell '$env:PATH = "$(go env GOPATH)\\bin;" + $env:PATH'
      }
    }
    stage('Test') {
      steps {
        powershell '''
          $env:GIN_MODE='release'
          $env:PATH = "$(go env GOPATH)\\bin;" + $env:PATH
          gotestsum --junitfile junit-report.xml -- -coverprofile=cover.out -v ./...
          go tool cover -html=cover.out -o cover.html
        '''
      }
      post {
        always {
          junit 'junit-report.xml'
          archiveArtifacts artifacts: 'cover.out,cover.html,junit-report.xml'
        }
      }
    }
  }
}
```

Notas y buenas prácticas para Jenkins

- Variables de entorno: Si tu aplicación utiliza variables (ej. SMTP para notificaciones) y quieres
  ejecutar tests que dependan de ellas, define esos secretos en Jenkins (Credentials) y pásalos como
  variables de entorno en el pipeline. Las pruebas actuales usan fallback cuando SMTP no está configurado,
  por lo que no requieren credenciales.
- Silenciar logs: establecer `GIN_MODE=release` reduce ruido en las salidas de tests.
- Timeout: establece timeouts de job/stage para evitar builds colgados si un test queda esperando.
- Parcheo de PATH: tras `go install`, añade `$(go env GOPATH)/bin` al PATH del proceso para usar herramientas instaladas.
- Reportes JUnit: `gotestsum` produce JUnit nativo; si usas `go test` directamente, puedes convertir la salida con `go-junit-report`.
- Cobertura: publica `cover.out` y `cover.html` como artefactos; puedes usar herramientas adicionales para analizar cobertura.
- Paralelismo: si planeas ejecutar tests en paralelo en varios agentes, asegúrate de que los tests estén totalmente aislados
  (nuestros tests usan `t.TempDir()` y `NewStoreWithPath(...)`, por lo que están preparados para esto).

Solución de problemas comunes

- Error: "command not found" para `gotestsum`: asegúrate de que `$(go env GOPATH)/bin` esté en PATH del agente.
- Tests que escriben en `services.json` del repo: si ves ese comportamiento, confirma que los tests usan `NewStoreWithPath`
  o que el job ejecuta en un workspace limpio.
- Tests con logs de Gin repetidos: setea `GIN_MODE=release` para evitar warnings de Gin durante la ejecución.

¿Quieres que añada estos artefactos al repo?

- Puedo crear un `Jenkinsfile` (Linux o Windows) en la raíz del repo con el ejemplo que prefieras.
- Puedo añadir scripts de conveniencia `run-tests.sh` y `run-tests.ps1` que encapsulen los comandos de CI (gotestsum + cobertura).

Si me confirmas si tu Jenkins corre en Linux o Windows y si quieres que añada el `Jenkinsfile` y/o los scripts,
los creo y los pruebo localmente antes de agregarlos al repo.
