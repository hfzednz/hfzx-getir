# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /out/bff-courier ./cmd/bff-courier
FROM alpine:3.20
COPY --from=build /out/bff-courier /usr/local/bin/bff-courier
EXPOSE 8112
ENTRYPOINT ["bff-courier"]
