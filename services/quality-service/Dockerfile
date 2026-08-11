# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/quality-service ./cmd/quality-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/quality-service /usr/local/bin/quality-service
EXPOSE 8118 9118
ENTRYPOINT ["quality-service"]
