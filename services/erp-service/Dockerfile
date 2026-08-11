# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/erp-service ./cmd/erp-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/erp-service /usr/local/bin/erp-service
EXPOSE 8108 9108
ENTRYPOINT ["erp-service"]
