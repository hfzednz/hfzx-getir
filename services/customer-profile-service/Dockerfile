FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/customer-profile-service ./cmd/customer-profile-service

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/customer-profile-service /app/customer-profile-service
COPY migrations /app/migrations
COPY api /app/api
ENV HTTP_ADDR=:8081
EXPOSE 8081
USER nobody
ENTRYPOINT ["/app/customer-profile-service"]
