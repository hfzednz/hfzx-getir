# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /promotion-service ./cmd/promotion-service

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /promotion-service /app/promotion-service
COPY migrations /app/migrations
ENV HTTP_ADDR=:8094
EXPOSE 8094
ENTRYPOINT ["/app/promotion-service"]
