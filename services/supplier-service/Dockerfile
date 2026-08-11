# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/supplier-service ./cmd/supplier-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/supplier-service /usr/local/bin/supplier-service
EXPOSE 8117 9117
ENTRYPOINT ["supplier-service"]
