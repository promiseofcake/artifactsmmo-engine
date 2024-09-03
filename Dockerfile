# build
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o artifactsmmo-engine ./cmd/engine

# runtime
FROM alpine:latest

ARG HOME_DIRECTORY

WORKDIR /app
COPY --from=builder /app/artifactsmmo-engine .
COPY .artifactsmmo-engine.yaml .

# Run the application
CMD ["./artifactsmmo-engine", "--config", "/app/.artifactsmmo-engine.yaml"]