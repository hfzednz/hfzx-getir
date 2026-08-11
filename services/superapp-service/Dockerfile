# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/superapp-service ./cmd/superapp-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/superapp-service /usr/local/bin/superapp-service
EXPOSE 8121 9121
ENTRYPOINT ["superapp-service"]
