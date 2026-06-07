.PHONY: build run test migrate-up migrate-down docker-up docker-down clean

# Build the server binary
build:
	go build -o bin/server ./cmd/server

# Run the server locally
run:
	go run ./cmd/server --config config/config.yaml

# Run tests
test:
	go test ./... -v -count=1

# Docker
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-build:
	docker-compose build

# Database migrations (requires golang-migrate CLI)
migrate-up:
	migrate -path migrations -database "mysql://root:root123@tcp(127.0.0.1:3306)/hxcoupon" up

migrate-down:
	migrate -path migrations -database "mysql://root:root123@tcp(127.0.0.1:3306)/hxcoupon" down

# Tidy dependencies
tidy:
	go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/
