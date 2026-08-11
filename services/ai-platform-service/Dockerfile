# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/ai-platform-service ./cmd/ai-platform-service
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/ai-platform-service /usr/local/bin/ai-platform-service
EXPOSE 8106 9106
ENTRYPOINT ["ai-platform-service"]
