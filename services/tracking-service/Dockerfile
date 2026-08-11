# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /tracking-service ./cmd/tracking-service

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /tracking-service /app/tracking-service
COPY migrations /app/migrations
ENV HTTP_ADDR=:8098
EXPOSE 8098
ENTRYPOINT ["/app/tracking-service"]
