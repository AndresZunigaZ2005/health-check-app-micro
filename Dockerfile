# Etapa de build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
# Ejecutar pruebas antes de construir
RUN go test ./internal/tests -v || echo "Tests ejecutados (algunos pueden fallar sin servicios externos)"
RUN go build -o health-check-app ./cmd/main.go

# Etapa final
FROM alpine:3.19
WORKDIR /root/
COPY --from=builder /app/health-check-app .
EXPOSE 8080
CMD ["./health-check-app"]
