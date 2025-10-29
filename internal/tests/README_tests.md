# Pruebas unitarias agrupadas

Este directorio contiene `all_tests_test.go`, que agrupa pruebas unitarias centradas en verificar
el comportamiento de los endpoints y del loop de verificación de salud (checker).

Objetivo de la reestructuración
- Evitar que las pruebas dependan del archivo `services.json` en el directorio del repositorio.
- Ejecutar las pruebas basadas en endpoints HTTP simulados (httptest) y en comportamientos en memoria.

Descripción de las pruebas (en `all_tests_test.go`)
- TestAPI_RegisterAndBackgroundCheck: envía un POST a `/register` y comprueba que el checker en background
	consulta el endpoint simulado y actualiza el estado del servicio a `UP`.
- TestChecker_UpdatesStatus: registra un servicio y valida que el loop de verificación cambia su `Status` a `UP`
	cuando el endpoint responde `{"status":"UP"}`.
- TestStore_UpdateService: comprueba en memoria que `Store.UpdateService` actualiza `Status` y `LastCheck`.
- TestAPI_Register_InvalidCases: valida respuestas 400 para JSON inválido, campo nombre ausente y endpoint inválido.
- TestAPI_HealthOne_NotFound: verifica que GET `/health/:name` devuelve 404 si el servicio no existe.
- TestNotifier_ConsoleFallback: comprueba que si no hay configuración SMTP, el notificador escribe
	los mensajes en el log (fallback a consola).

Cómo se aislan las pruebas
- Se añadió en `internal/store` la función `NewStoreWithPath(path string) *Store`.
	Las pruebas crean un directorio temporal con `t.TempDir()` y construyen el store con
	`filepath.Join(tmp, "services.json")` para que cualquier persistencia quede en un archivo temporal
	fuera del repo.

Ejecutar las pruebas (PowerShell)

1) Ejecutar todas las pruebas del paquete de tests:
```powershell
cd C:\Users\juanm\Github\health-check-app-micro
go test ./internal/tests -v
```

2) Ejecutar todas las pruebas del repo:
```powershell
go test ./... -v
```

3) Ejecutar una prueba concreta (por nombre exacto usando regexp):
```powershell
go test ./internal/tests -run "^TestAPI_RegisterAndBackgroundCheck$" -v
```

4) Ejecutar con detector de race:
```powershell
go test ./internal/tests -race -v
```

Notas
- Las pruebas ahora se centran en endpoints y en la lógica en memoria; si quieres que además
	cubramos la persistencia en disco, podemos añadir tests específicos que usen `NewStoreWithPath`
	y verifiquen el contenido del archivo temporal.

Si quieres, añado un script `run-tests.ps1` que ejecute las opciones más comunes (all, race, cover, junit).
