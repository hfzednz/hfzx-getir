# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /inventory-service ./cmd/inventory-service

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /inventory-service /app/inventory-service
COPY migrations /app/migrations
ENV HTTP_ADDR=:8083
EXPOSE 8083
ENTRYPOINT ["/app/inventory-service"]
