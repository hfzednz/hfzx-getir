# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /chat-assistant-service ./cmd/chat-assistant-service

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /chat-assistant-service /app/chat-assistant-service
ENV HTTP_ADDR=:8126
EXPOSE 8126
ENTRYPOINT ["/app/chat-assistant-service"]
