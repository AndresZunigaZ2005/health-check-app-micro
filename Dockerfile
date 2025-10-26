# Etapa de build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o health-check-app ./cmd/main.go

# Etapa final
FROM alpine:3.19
WORKDIR /root/
COPY --from=builder /app/health-check-app .
EXPOSE 8080
CMD ["./health-check-app"]
