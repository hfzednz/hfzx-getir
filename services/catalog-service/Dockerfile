# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /catalog-service ./cmd/catalog-service

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /catalog-service /app/catalog-service
COPY migrations /app/migrations
ENV HTTP_ADDR=:8082
EXPOSE 8082
ENTRYPOINT ["/app/catalog-service"]
