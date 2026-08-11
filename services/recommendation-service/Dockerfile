# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/recommendation-service ./cmd/recommendation-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/recommendation-service /usr/local/bin/recommendation-service
EXPOSE 8105 9105
ENTRYPOINT ["recommendation-service"]
