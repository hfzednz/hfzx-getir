# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/review-service ./cmd/review-service

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/review-service /usr/local/bin/review-service
EXPOSE 8103 9103
ENTRYPOINT ["review-service"]
