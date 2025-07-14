.PHONY: build run test clean seed reset-db docker-build docker-run

# Build the application
build:
	go build -o bin/wallet-service .

# Run the application
run:
	go run .

# Seed the database with test data
seed:
	go run cmd/seed/main.go

# Reset database and seed
reset-db:
	docker compose down -v
	docker compose up -d postgres redis
	sleep 5
	go run cmd/seed/main.go

# Build Docker image
docker-build:
	docker compose build

# Run with Docker Compose
docker-run:
	docker compose up -d

# Stop Docker services
docker-stop:
	docker compose down

# View logs
docker-logs:
	docker compose logs -f