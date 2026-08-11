.PHONY: test run build migrate tidy docker-up docker-down

MODULE := github.com/nexora/identity-service

tidy:
	go mod tidy

test:
	go test ./...

build:
	go build -o bin/identity-service ./cmd/identity-service

run:
	OTP_DEV_MODE=true HTTP_ADDR=:8080 go run ./cmd/identity-service

migrate:
	@echo "Applying SQL migrations to $$DATABASE_URL"
	@test -n "$$DATABASE_URL" || (echo "DATABASE_URL required" && exit 1)
	@for f in migrations/*.sql; do \
		echo ">> $$f"; \
		psql "$$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$$f"; \
	done

docker-up:
	docker compose up --build -d postgres redis identity-service

docker-up-kafka:
	docker compose --profile kafka up --build -d

docker-down:
	docker compose --profile kafka down -v
