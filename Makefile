DB_SERVICE_NAME=postgres

# Application Entry point
start: check-db
	go run main.go

# Internal target: Ensures DB is up and running before proceeding
check-db:
	@echo "Checking database status..."
	@if [ -z "$$(docker compose ps -q $(DB_SERVICE_NAME))" ] || [ -z "$$(docker compose ps --filter "status=running" -q $(DB_SERVICE_NAME))" ]; then \
		echo "Database is not running. Starting database..."; \
		docker compose up -d; \
		echo "Waiting for database to be ready..."; \
		sleep 5; \
	else \
		echo "Database is already running."; \
	fi

# Run all tests
test:
	go test -v ./...

# Run tests for determining test coverage
test-cover:
	go test -cover ./...

# Starts database instance
start-db:
	docker compose up -d

# Stops database  
stop-db:
	docker compose down

# Stops database and deletes it's data
clean-db:
	docker compose down -v