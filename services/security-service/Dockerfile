# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/security-service ./cmd/security-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/security-service /usr/local/bin/security-service
EXPOSE 8109 9109
ENTRYPOINT ["security-service"]
