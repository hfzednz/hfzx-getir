# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/platform-ops-service ./cmd/platform-ops-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/platform-ops-service /usr/local/bin/platform-ops-service
EXPOSE 8110 9110
ENTRYPOINT ["platform-ops-service"]
