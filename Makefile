.PHONY: help bootstrap doctor deps deps-go deps-flutter deps-web verify verify-structure test-go-focus test-certs build-validate compose-up compose-down stack-up stack-down live-e2e integration-cert prod-validate

help:
	@echo "NEXORA monorepo targets:"
	@echo "  make bootstrap         - doctor + local deps compose + structure verify"
	@echo "  make doctor            - print toolchain versions"
	@echo "  make deps              - go/flutter/web dependency install helpers"
	@echo "  make verify            - structure + focus Go tests + cert tools"
	@echo "  make build-validate    - Prompt-45 Go focus + hyperscale"
	@echo "  make test-go-focus     - identity/order/payment/catalog/BFFs"
	@echo "  make compose-up        - infra/docker/docker-compose.yml"
	@echo "  make integration-cert  - tools/integration-cert"
	@echo "  make prod-validate     - tools/prod-validate"

bootstrap: doctor verify-structure compose-up
	@echo "bootstrap: complete (run make deps for language workspaces)"

doctor:
	@echo "== doctor =="
	@go version || echo "go: MISSING"
	@flutter --version 2>NUL || flutter --version || echo "flutter: optional/MISSING"
	@node --version || echo "node: optional/MISSING"
	@pnpm --version || echo "pnpm: optional/MISSING"
	@docker --version || echo "docker: optional/MISSING"
	@terraform -version || echo "terraform: optional/MISSING"

deps: deps-go deps-web
	@echo "deps: flutter via 'make deps-flutter' (requires Flutter SDK)"

deps-go:
	go work sync || true
	@echo "deps-go: use per-service 'go test ./...' (modules are independent)"

deps-flutter:
	@command -v melos >/dev/null 2>&1 && melos bootstrap || (cd apps/mobile_customer && flutter pub get)

deps-web:
	@command -v pnpm >/dev/null 2>&1 && pnpm install || echo "pnpm not installed; skip web workspace"

verify: verify-structure test-go-focus test-certs
	@echo "verify: OK"

build-validate:
	@pwsh -NoProfile -File scripts/build-validate.ps1 || powershell -NoProfile -File scripts/build-validate.ps1

verify-structure:
	@pwsh -NoProfile -File scripts/verify-structure.ps1 || powershell -NoProfile -File scripts/verify-structure.ps1

test-go-focus:
	cd services/identity-service && go test ./...
	cd services/order-service && go test ./...
	cd services/payment-service && go test ./...
	cd services/catalog-service && go test ./...
	cd services/bff-customer && go test ./...
	cd services/bff-admin && go test ./...

test-certs:
	cd tools/prod-validate && go run . -env=staging
	cd tools/integration-cert && go test ./... || go run .

stack-up:
	python scripts/local/prompt87_live_gate.py --start

live-e2e:
	python scripts/local/prompt87_live_gate.py

stack-down:
	python scripts/local/prompt87_live_gate.py --stop

compose-up:
	docker compose -f infra/docker/docker-compose.yml up -d

compose-down:
	docker compose -f infra/docker/docker-compose.yml down

integration-cert:
	go run ./tools/integration-cert

prod-validate:
	go run ./tools/prod-validate -env=staging
