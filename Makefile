.PHONY: dev build test fmt vet tidy frontend docker

dev:
	go run ./cmd/server/

build:
	go build -o dartscheduler ./cmd/server/

test:
	go test $(shell go list ./... | grep -v /frontend/)

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

frontend:
	cd frontend && npm run build -- --configuration=production && \
	cp -r dist/dart-scheduler ../web/dist/

docker:
	docker compose up --build
