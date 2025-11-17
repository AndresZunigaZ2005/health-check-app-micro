Feature: Obtener la salud de un microservicio
  Scenario: Consultar salud individual
    Given existe el microservicio "jwtmanual-taller1-micro"
    When hago GET a "/health/jwtmanual-taller1-micro"
    Then la respuesta debe tener código 200
    And el cuerpo debe contener el nombre "jwtmanual-taller1-micro"
