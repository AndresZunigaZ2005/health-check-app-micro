# Health Check App Microservice - Documentación de Implementación

## Descripción General

El microservicio Health Check App es una aplicación desarrollada en Go que monitorea el estado de salud de otros microservicios mediante verificaciones HTTP periódicas. Permite registrar servicios dinámicamente, consultar su estado y recibir notificaciones por correo cuando algún servicio falla.

## Arquitectura

### Componentes Principales

El microservicio está estructurado en los siguientes componentes:

1. **API Handlers**: Endpoints HTTP para registro y consulta de servicios
2. **Checker**: Verificador de salud que realiza peticiones HTTP periódicas
3. **Store**: Almacenamiento en memoria con persistencia en archivo JSON
4. **Notifier**: Enviador de notificaciones por correo electrónico
5. **Models**: Modelos de datos para servicios y estados
6. **Utils**: Utilidades de logging y helpers

### Tecnologías Utilizadas

- **Go**: Lenguaje de programación
- **Gin**: Framework web minimalista
- **HTTP Client**: Cliente HTTP estándar de Go
- **JSON**: Serialización/deserialización de datos
- **SMTP**: Envío de correos electrónicos
- **File I/O**: Persistencia en archivo JSON local

## Modelo de Datos

### Microservice

Modelo que representa un microservicio monitoreado.

- **Name** (String): Nombre único del microservicio
- **Endpoint** (String): URL del endpoint de health check (http:// o https://)
- **Frequency** (Integer): Frecuencia de verificación en segundos (mínimo 10)
- **Status** (String): Estado actual (UP, DOWN, UNKNOWN)
- **LastCheck** (String): Fecha y hora de última verificación (RFC3339)
- **LastSuccess** (String): Fecha y hora de último éxito (RFC3339, opcional)
- **LastFailure** (String): Fecha y hora de último fallo (RFC3339, opcional)
- **FailureCount** (Integer): Contador de fallos consecutivos
- **Metadata** (Map): Metadatos adicionales del servicio

### Health Status

Estados posibles de un microservicio:

- **UP**: El servicio está funcionando correctamente
- **DOWN**: El servicio no está respondiendo
- **UNKNOWN**: Estado inicial antes de la primera verificación

## Endpoints de la API

### Registro de Microservicio

- **Endpoint**: `POST /register`
- **Descripción**: Registra un nuevo microservicio para monitoreo
- **Autenticación**: No requerida
- **Request Body**:
```json
{
  "name": "api-gateway",
  "endpoint": "http://api-gateway:8085/actuator/health",
  "frequency": 30,
  "metadata": {
    "description": "API Gateway Service",
    "team": "backend"
  }
}
```
- **Response**: 201 Created con datos del servicio registrado

### Consulta de Estado Global

- **Endpoint**: `GET /health`
- **Descripción**: Obtiene el estado de todos los microservicios registrados
- **Autenticación**: No requerida
- **Response**: Mapa de servicios con su estado actual
```json
{
  "api-gateway": {
    "name": "api-gateway",
    "endpoint": "http://api-gateway:8085/actuator/health",
    "status": "UP",
    "lastCheck": "2024-01-01T12:00:00Z",
    "frequency": 30
  },
  "gestion-perfil": {
    "name": "gestion-perfil",
    "endpoint": "http://gestion-perfil:8080/actuator/health",
    "status": "DOWN",
    "lastCheck": "2024-01-01T12:00:00Z",
    "frequency": 30
  }
}
```

### Consulta de Estado Individual

- **Endpoint**: `GET /health/{name}`
- **Descripción**: Obtiene el estado de un microservicio específico
- **Autenticación**: No requerida
- **Response**: Datos del servicio solicitado o 404 si no existe

## Componentes de Implementación

### API Handlers (api/handler.go)

Manejadores HTTP que procesan las solicitudes de la API.

#### RegisterHandler

Maneja el registro de nuevos microservicios.

**Funcionalidades**:
- Valida datos de entrada (nombre, endpoint, frecuencia)
- Verifica que el endpoint comience con http:// o https://
- Establece frecuencia mínima de 10 segundos (default 30)
- Inicializa estado como UNKNOWN
- Registra servicio en Store
- Inicia monitoreo del servicio con Checker
- Retorna respuesta 201 con datos del servicio

#### HealthAllHandler

Maneja la consulta de estado global.

**Funcionalidades**:
- Obtiene todos los servicios del Store
- Retorna mapa de servicios con su estado actual
- Formato JSON estructurado

#### HealthOneHandler

Maneja la consulta de estado individual.

**Funcionalidades**:
- Extrae nombre del servicio de la URL
- Busca servicio en Store
- Retorna datos del servicio o 404 si no existe

### Checker (checker/checker.go)

Verificador de salud que realiza peticiones HTTP periódicas.

#### StartHealthCheckLoop

Inicia el loop de verificación para todos los servicios registrados.

**Funcionalidades**:
- Obtiene todos los servicios del Store
- Inicia goroutine para cada servicio
- Cada goroutine ejecuta verificaciones periódicas

#### checkHealthWithFrequency

Ejecuta verificaciones periódicas para un servicio específico.

**Funcionalidades**:
- Crea ticker con frecuencia del servicio
- Ejecuta primera verificación inmediatamente
- Ejecuta verificaciones periódicas según frecuencia
- Maneja goroutine de forma independiente

#### checkHealth

Realiza una verificación individual de salud.

**Funcionalidades**:
- Realiza petición HTTP GET al endpoint del servicio
- Configura timeout de 5 segundos
- Verifica código de respuesta HTTP
- Actualiza estado en Store (UP o DOWN)
- Actualiza timestamps (LastCheck, LastSuccess, LastFailure)
- Incrementa FailureCount si falla
- Resetea FailureCount si tiene éxito
- Envía notificación por correo si detecta fallo

#### RegisterNewService

Registra un nuevo servicio para monitoreo después del inicio.

**Funcionalidades**:
- Inicia goroutine de verificación para nuevo servicio
- Permite registro dinámico sin reiniciar el servicio

### Store (store/store.go)

Almacenamiento en memoria con persistencia en archivo JSON.

#### NewStore

Crea una nueva instancia de Store.

**Funcionalidades**:
- Inicializa mapa en memoria
- Carga datos desde archivo `services.json` si existe
- Crea archivo si no existe
- Configura guardado automático periódico

#### RegisterService

Registra un nuevo servicio en el Store.

**Funcionalidades**:
- Almacena servicio en mapa en memoria
- Guarda datos en archivo JSON
- Thread-safe con mutex

#### Get

Obtiene un servicio específico por nombre.

**Funcionalidades**:
- Busca servicio en mapa
- Retorna puntero al servicio o nil si no existe
- Thread-safe con mutex

#### GetAll

Obtiene todos los servicios registrados.

**Funcionalidades**:
- Retorna mapa completo de servicios
- Thread-safe con mutex

#### UpdateStatus

Actualiza el estado de un servicio.

**Funcionalidades**:
- Actualiza campos de estado
- Actualiza timestamps
- Guarda cambios en archivo JSON
- Thread-safe con mutex

### Notifier (notifier/notifier.go)

Enviador de notificaciones por correo electrónico.

#### SendNotification

Envía notificación por correo cuando un servicio falla.

**Funcionalidades**:
- Construye mensaje de correo con detalles del fallo
- Envía correo vía SMTP
- Configuración desde variables de entorno
- Maneja errores de envío sin interrumpir flujo principal

**Configuración SMTP**:
- SMTP_HOST: Servidor SMTP
- SMTP_PORT: Puerto SMTP
- SMTP_USER: Usuario SMTP
- SMTP_PASSWORD: Contraseña SMTP
- SMTP_FROM: Email remitente
- SMTP_TO: Email destinatario

### Models

#### Microservice (models/microservice.go)

Modelo de datos para microservicio.

#### Health Status (models/health_status.go)

Constantes para estados de salud.

### Utils (pkg/utils/utils.go)

Utilidades de logging y helpers.

**Funcionalidades**:
- Inicialización de logger
- Funciones de logging (Info, Error, Warn)
- Formateo de mensajes

## Flujos de Procesamiento

### Flujo de Registro de Servicio

1. Cliente envía `POST /register` con datos del servicio
2. `RegisterHandler` recibe la solicitud
3. Valida datos de entrada:
   - Nombre requerido
   - Endpoint requerido y válido (http:// o https://)
   - Frecuencia mínima de 10 segundos
4. Crea instancia de `Microservice` con estado UNKNOWN
5. `Store.RegisterService()` almacena servicio
6. `Checker.RegisterNewService()` inicia monitoreo
7. Retorna respuesta 201 con datos del servicio

### Flujo de Verificación de Salud

1. `Checker.checkHealthWithFrequency()` crea ticker con frecuencia del servicio
2. Ejecuta primera verificación inmediatamente
3. `Checker.checkHealth()` realiza petición HTTP GET al endpoint
4. Configura timeout de 5 segundos
5. Verifica código de respuesta:
   - Si 200-299: Estado UP
   - Si otro código o error: Estado DOWN
6. `Store.UpdateStatus()` actualiza estado y timestamps
7. Si falla, incrementa `FailureCount`
8. Si falla y es primera vez, `Notifier.SendNotification()` envía correo
9. Espera hasta siguiente tick del ticker
10. Repite proceso periódicamente

### Flujo de Consulta de Estado

1. Cliente envía `GET /health` o `GET /health/{name}`
2. Handler correspondiente recibe la solicitud
3. `Store.Get()` o `Store.GetAll()` obtiene datos
4. Retorna respuesta JSON con estado actual

### Flujo de Notificación por Correo

1. `Checker.checkHealth()` detecta fallo en servicio
2. Verifica si es primera vez que falla (FailureCount == 1)
3. `Notifier.SendNotification()` se ejecuta:
   - Construye mensaje con detalles del fallo
   - Conecta a servidor SMTP
   - Envía correo con información del servicio
4. Si hay error en envío, se registra pero no interrumpe flujo

## Configuración

### Variables de Entorno

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

# Logging
LOG_LEVEL=info
```

### Persistencia

El servicio persiste datos en archivo `services.json`:

- **Ubicación**: Directorio de ejecución
- **Formato**: JSON con mapa de servicios
- **Guardado**: Automático después de cada actualización
- **Carga**: Al iniciar el servicio

## Testing

### Estructura de Tests

- **Unit Tests**: Pruebas de handlers, checker y store
- **Integration Tests**: Pruebas de endpoints con httptest
- **Mocking**: Uso de mocks para HTTP y SMTP

### Cobertura de Tests

El proyecto incluye tests para:
- API Handlers (15+ tests)
- Checker (verificación de salud)
- Store (almacenamiento)
- Casos exitosos y casos de error

## Despliegue

### Docker

El microservicio incluye un `Dockerfile` para contenedorización.

### Docker Compose

Configurado en `docker-compose.unified.yml`:
- Puerto: 8080
- Dependencias: Ninguna (solo requiere acceso HTTP a otros servicios)
- Health checks configurados

## Monitoreo y Logging

### Logging

- **Logger personalizado**: Utilidades de logging en `pkg/utils`
- **Niveles**: INFO, WARN, ERROR
- **Formato**: Texto estructurado con timestamps

### Health Checks

El servicio mismo no expone health check, pero puede ser monitoreado por otro Health Check App (recursivo).

## Consideraciones de Seguridad

1. **Validación de entrada**: Validación exhaustiva de endpoints
2. **Timeout**: Timeout de 5 segundos para evitar bloqueos
3. **SMTP**: Credenciales almacenadas en variables de entorno
4. **HTTPS**: Verificar endpoints HTTPS cuando sea posible
5. **Rate Limiting**: Considerar límites para registro de servicios

## Limitaciones y Consideraciones

1. **Persistencia local**: Datos almacenados en archivo JSON local
2. **Sin base de datos**: No hay base de datos, solo almacenamiento en memoria y archivo
3. **Monitoreo básico**: Solo verifica código HTTP, no valida contenido
4. **Notificaciones simples**: Solo notificaciones por correo, sin otros canales
5. **Sin historial**: No mantiene historial de verificaciones pasadas

## Mejoras Futuras

1. **Base de datos**: Migrar a PostgreSQL para persistencia robusta
2. **Historial**: Mantener historial de verificaciones
3. **Métricas**: Integración con Prometheus/Grafana
4. **Alertas múltiples**: Soporte para múltiples canales de notificación
5. **Validación de contenido**: Validar contenido de respuesta, no solo código HTTP
6. **Dashboard**: Interfaz web para visualización de estado
7. **Agrupación**: Agrupar servicios por categorías o equipos
8. **Dependencias**: Modelar dependencias entre servicios
9. **Circuit Breaker**: Implementar circuit breaker para servicios críticos
10. **Distributed Tracing**: Integración con sistemas de trazabilidad

