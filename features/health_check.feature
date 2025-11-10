# language: es
# =============================================================================
# ARCHIVO DE CARACTERÍSTICAS (FEATURES) - HEALTH CHECK APP
# =============================================================================

Característica: Monitoreo de Salud de Microservicios
  
  Antecedentes:
    Dado que el servicio de health check está disponible

  # ===== REGISTRAR SERVICIO =====
  Escenario: Registrar un microservicio para monitoreo
    Cuando registro un microservicio con datos válidos
    Entonces la respuesta debe tener estado 201
    Y el cuerpo debe contener los datos del servicio registrado

  # ===== CONSULTAR ESTADO GLOBAL =====
  Escenario: Consultar estado de todos los servicios
    Dado que existe al menos un servicio registrado
    Cuando consulto el estado de todos los servicios
    Entonces la respuesta debe tener estado 200
    Y el cuerpo debe contener información de servicios

  # ===== CONSULTAR ESTADO INDIVIDUAL =====
  Escenario: Consultar estado de un servicio específico
    Dado que existe al menos un servicio registrado
    Cuando consulto el estado de un servicio específico
    Entonces la respuesta debe tener estado 200
    Y el cuerpo debe contener información del servicio

  # ===== HEALTH CHECK =====
  Escenario: Verificar salud del servicio de health check
    Cuando consulto el endpoint de health check
    Entonces la respuesta debe tener estado 200

