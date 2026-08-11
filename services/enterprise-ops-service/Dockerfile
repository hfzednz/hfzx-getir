# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/enterprise-ops-service ./cmd/enterprise-ops-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/enterprise-ops-service /usr/local/bin/enterprise-ops-service
EXPOSE 8123 9123
ENTRYPOINT ["enterprise-ops-service"]
