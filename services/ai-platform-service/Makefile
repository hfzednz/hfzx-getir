.PHONY: test build run ml
test:
	go test ./...
build:
	go build -o bin/ai-platform-service ./cmd/ai-platform-service
run:
	go run ./cmd/ai-platform-service
ml:
	cd ml && uvicorn app.main:app --host 0.0.0.0 --port 8206
