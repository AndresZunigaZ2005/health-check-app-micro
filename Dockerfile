# builder
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /monitor ./cmd/monitor

# final
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /monitor /monitor
EXPOSE 8080
ENTRYPOINT ["/monitor"]
