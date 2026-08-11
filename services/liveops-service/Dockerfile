# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/liveops-service ./cmd/liveops-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/liveops-service /usr/local/bin/liveops-service
EXPOSE 8116 9116
ENTRYPOINT ["liveops-service"]
