.PHONY: dev dev-all build test fmt fmt-check lint clean vet tidy frontend docker ci help test-coverage

dev:
	go run ./cmd/server/

dev-all:
	@echo "Starting backend on :8080 and frontend on :4200..."
	@(cd frontend && npm start) & go run ./cmd/server/ & wait

build:
	go build -o dartscheduler ./cmd/server/

test:
	go test -race -timeout 300s ./cmd/... ./domain/... ./usecase/... ./scheduler/... ./infra/...

test-coverage:
	go test -race -timeout 300s -coverprofile=coverage.out ./cmd/... ./domain/... ./usecase/... ./scheduler/... ./infra/...
	go tool cover -func=coverage.out

fmt:
	go fmt ./...
	cd frontend && npm run format || true

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Go files need formatting:" && gofmt -l . && exit 1)

lint:
	go vet ./...
	cd frontend && npm run lint || echo "Frontend lint not configured yet"

vet:
	go vet ./...

tidy:
	go mod tidy

frontend:
	cd frontend && npm run build -- --configuration=production

docker:
	docker compose up --build

clean:
	rm -f dartscheduler
	rm -rf frontend/.angular/cache
	rm -rf web/dist

ci: fmt-check lint test frontend
	@echo "CI checks passed"

help:
	@echo "Available targets:"
	@echo "  make dev        - Start Go backend (port 8080)"
	@echo "  make dev-all    - Start backend + frontend together"
	@echo "  make frontend   - Build Angular frontend"
	@echo "  make test       - Run Go tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make lint       - Run all linters"
	@echo "  make fmt        - Format all code"
	@echo "  make fmt-check  - Check code formatting without changes"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make docker     - Build and start via docker-compose"
	@echo "  make ci         - Run full CI pipeline"
