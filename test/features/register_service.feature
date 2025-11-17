Feature: Registrar microservicios
  Scenario: Registrar un microservicio correctamente
    Given tengo un microservicio llamado "jwtmanual-taller1-micro"
    And su endpoint "http://jwtmanual-taller1-micro:8081/health"
    And su frecuencia "10"
    And su email "unieventosuq5@gmail.com"
    When hago POST a "/register"
    Then la respuesta debe tener código 201
    And el cuerpo debe contener "Microservicio registrado exitosamente"
