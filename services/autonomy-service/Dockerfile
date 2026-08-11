# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/autonomy-service ./cmd/autonomy-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/autonomy-service /usr/local/bin/autonomy-service
EXPOSE 8125 9125
ENTRYPOINT ["autonomy-service"]
