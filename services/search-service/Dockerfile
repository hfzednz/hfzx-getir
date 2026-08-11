# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/search-service ./cmd/search-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/search-service /usr/local/bin/search-service
EXPOSE 8104 9104
ENTRYPOINT ["search-service"]
