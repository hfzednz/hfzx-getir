# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/hyperscale-cert-service ./cmd/hyperscale-cert-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/hyperscale-cert-service /usr/local/bin/hyperscale-cert-service
EXPOSE 8124 9124
ENTRYPOINT ["hyperscale-cert-service"]
