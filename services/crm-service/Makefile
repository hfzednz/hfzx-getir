.PHONY: test run build migrate tidy docker-up docker-down

MODULE := github.com/nexora/crm-service

tidy:
	go mod tidy

test:
	go test ./...

build:
	go build -o bin/crm-service ./cmd/crm-service

run:
	HTTP_ADDR=:8102 go run ./cmd/crm-service

migrate:
	@echo "Applying SQL migrations to $$DATABASE_URL"
	@test -n "$$DATABASE_URL" || (echo "DATABASE_URL required" && exit 1)
	@for f in migrations/*.sql; do \
		echo ">> $$f"; \
		psql "$$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$$f"; \
	done

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v
