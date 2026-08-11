# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /out/bff-admin ./cmd/bff-admin
FROM alpine:3.20
COPY --from=build /out/bff-admin /usr/local/bin/bff-admin
EXPOSE 8114
ENTRYPOINT ["bff-admin"]
