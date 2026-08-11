# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /out/realtime-gateway ./cmd/realtime-gateway
FROM alpine:3.20
COPY --from=build /out/realtime-gateway /usr/local/bin/realtime-gateway
EXPOSE 8115
ENTRYPOINT ["realtime-gateway"]
