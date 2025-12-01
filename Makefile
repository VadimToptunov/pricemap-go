.PHONY: build run test clean docker-build docker-up docker-down migrate

# Build all binaries
build:
	go build -o bin/server ./cmd/server
	go build -o bin/scraper ./cmd/scraper
	go build -o bin/scheduler ./cmd/scheduler

# Run server
run:
	go run ./cmd/server

# Run scraper
scrape:
	go run ./cmd/scraper

# Run scheduler
schedule:
	go run ./cmd/scheduler

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker build
docker-build:
	docker-compose build

# Docker up
docker-up:
	docker-compose up -d

# Docker down
docker-down:
	docker-compose down

# Run migrations
migrate:
	go run ./cmd/server

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Run all linters
vet:
	go vet ./...

