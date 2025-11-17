Feature: Obtener estado de todos los microservicios
  Scenario: Consultar salud global
    Given hay microservicios registrados
    When hago GET a "/health"
    Then la respuesta debe tener código 200
    And la respuesta debe ser una lista de microservicios
