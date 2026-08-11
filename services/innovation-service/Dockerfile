# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/innovation-service ./cmd/innovation-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/innovation-service /usr/local/bin/innovation-service
EXPOSE 8122 9122
ENTRYPOINT ["innovation-service"]
