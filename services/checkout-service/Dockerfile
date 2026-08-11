# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /checkout-service ./cmd/checkout-service

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /checkout-service /app/checkout-service
COPY migrations /app/migrations
ENV HTTP_ADDR=:8088
EXPOSE 8088
ENTRYPOINT ["/app/checkout-service"]
