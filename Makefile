.PHONY: build run test lint clean

# Build the application
build:
	CGO_ENABLED=1 go build -o bin/release-catalog ./cmd/server

# Run the application locally
run: build
	./bin/release-catalog

# Run tests
test:
	CGO_ENABLED=1 go test -v ./...

# Run tests with coverage
test-coverage:
	CGO_ENABLED=1 go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html *.db

# Docker
docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

# DT reference (separate)
dt-up:
	docker compose -f docker-compose.dt.yml up -d

dt-down:
	docker compose -f docker-compose.dt.yml down
