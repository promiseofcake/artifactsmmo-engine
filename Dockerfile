# build
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd
COPY internal/ ./internal
RUN find /app

RUN go build -o artifactsmmo-engine ./cmd/engine

# runtime
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/artifactsmmo-engine .
COPY .artifactsmmo-engine.yaml .

# Run the application
CMD ["./artifactsmmo-engine", "--config", "/app/.artifactsmmo-engine.yaml"]