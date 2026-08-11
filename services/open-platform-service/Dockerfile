# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/open-platform-service ./cmd/open-platform-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/open-platform-service /usr/local/bin/open-platform-service
EXPOSE 8120 9120
ENTRYPOINT ["open-platform-service"]
