# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /out/bff-warehouse ./cmd/bff-warehouse
FROM alpine:3.20
COPY --from=build /out/bff-warehouse /usr/local/bin/bff-warehouse
EXPOSE 8113
ENTRYPOINT ["bff-warehouse"]
