# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/data-platform-service ./cmd/data-platform-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/data-platform-service /usr/local/bin/data-platform-service
EXPOSE 8107 9107
ENTRYPOINT ["data-platform-service"]
