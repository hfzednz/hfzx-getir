# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/bff-customer ./cmd/bff-customer
FROM alpine:3.20
COPY --from=build /out/bff-customer /usr/local/bin/bff-customer
EXPOSE 8111
ENTRYPOINT ["bff-customer"]
