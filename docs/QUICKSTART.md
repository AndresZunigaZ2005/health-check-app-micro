# Guía de Inicio Rápido - Health Check App Microservice

Esta guía te permitirá poner en marcha el microservicio de monitoreo de salud en tu entorno local y ejecutar pruebas básicas.

## Requisitos Previos

- Go 1.22 o superior
- Git

## Instalación Rápida

### 1. Clonar y Compilar

```bash
cd health-check-app-micro
go mod download
go build -o health-check-app-micro ./cmd
```

### 2. Configurar Variables de Entorno (Opcional)

Crear archivo `.env` en la raíz del proyecto:

```env
# Server
PORT=8080

# SMTP (Opcional, para notificaciones)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=app-password
SMTP_FROM=your-email@gmail.com
SMTP_TO=admin@example.com
```

### 3. Ejecutar Localmente

```bash
./health-check-app-micro
```

O sin compilar:

```bash
go run ./cmd
```

El servidor estará disponible en `http://localhost:8080`

## Verificación Inicial

### Verificar que el Servidor Está Corriendo

```bash
curl http://localhost:8080/health
```

Respuesta esperada: HTTP 200 con objeto vacío `{}` (no hay servicios registrados aún)

## Pruebas Básicas

### 1. Registrar un Microservicio para Monitoreo

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "api-gateway",
    "endpoint": "http://localhost:8085/actuator/health",
    "frequency": 30,
    "metadata": {
      "description": "API Gateway Service",
      "team": "backend"
    }
  }'
```

Respuesta esperada: HTTP 201 con datos del servicio registrado

### 2. Consultar Estado Global

```bash
curl http://localhost:8080/health
```

Respuesta esperada: HTTP 200 con mapa de servicios y su estado:
```json
{
  "api-gateway": {
    "name": "api-gateway",
    "endpoint": "http://localhost:8085/actuator/health",
    "status": "UP",
    "lastCheck": "2024-01-01T12:00:00Z",
    "frequency": 30
  }
}
```

### 3. Consultar Estado Individual

```bash
curl http://localhost:8080/health/api-gateway
```

Respuesta esperada: HTTP 200 con datos del servicio específico

### 4. Registrar Múltiples Servicios

```bash
# Registrar otro servicio
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "gestion-perfil",
    "endpoint": "http://localhost:8080/actuator/health",
    "frequency": 30
  }'

# Ver todos los servicios
curl http://localhost:8080/health
```

## Verificar Monitoreo Automático

Después de registrar un servicio, el checker comenzará a verificar su salud automáticamente según la frecuencia configurada. Verás logs como:

```
Verificando salud de: api-gateway
Estado: UP
```

### Verificar Archivo de Persistencia

Los servicios se guardan automáticamente en `services.json`:

```bash
cat services.json
```

## Ejecutar Tests

### Tests Unitarios

```bash
go test ./...
```

### Tests con Cobertura

```bash
go test ./... -cover
```

### Tests con Reporte Detallado

```bash
go test ./... -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Simular Servicio que Falla

### 1. Registrar un Servicio con Endpoint Inválido

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "servicio-falso",
    "endpoint": "http://localhost:9999/health",
    "frequency": 10
  }'
```

### 2. Esperar Verificación

Después de 10 segundos, el servicio debería mostrar estado `DOWN`:

```bash
curl http://localhost:8080/health/servicio-falso
```

### 3. Verificar Notificación por Correo

Si SMTP está configurado, deberías recibir un correo cuando el servicio pase a `DOWN` por primera vez.

## Verificar Logs

El servicio genera logs estructurados. Revisa la salida de consola para ver:
- Registro de servicios
- Verificaciones de salud
- Errores de conexión
- Notificaciones enviadas

## Troubleshooting

### Error: Puerto 8080 ya en uso

Cambiar el puerto usando variable de entorno:
```bash
PORT=8081 ./health-check-app-micro
```

### Error: Endpoint no responde

Verificar que el servicio monitoreado esté corriendo:
```bash
curl http://localhost:8085/actuator/health
```

### Error: services.json no se crea

Verificar permisos de escritura en el directorio:
```bash
ls -la services.json
chmod 644 services.json
```

### Error: SMTP no funciona

Las notificaciones por correo son opcionales. El servicio funcionará sin SMTP, solo no enviará correos cuando un servicio falle.

Para pruebas, puedes omitir la configuración de SMTP.

### Verificar que el Checker Está Funcionando

Revisar logs de consola. Deberías ver mensajes periódicos como:
```
Verificando salud de: api-gateway
Estado: UP
```

Si no ves estos mensajes, verificar que el servicio esté registrado y que la frecuencia no sea demasiado larga.

## Próximos Pasos

- Revisar `docs/IMPLEMENTATION.md` para detalles de arquitectura
- Configurar SMTP para recibir notificaciones por correo
- Registrar todos los microservicios de la plataforma
- Integrar con sistemas de alertas externos
- Revisar el archivo `services.json` para persistencia de datos

