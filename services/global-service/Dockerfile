# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/global-service ./cmd/global-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/global-service /usr/local/bin/global-service
EXPOSE 8119 9119
ENTRYPOINT ["global-service"]
