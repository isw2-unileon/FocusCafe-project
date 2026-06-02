.PHONY: install run-backend run-frontend build-backend build-frontend test lint e2e \
       db-up db-reset db-down docker-up docker-up-remote docker-down docker-build

ifeq ($(OS),Windows_NT)
    AIR_CONFIG := backend/.air.windows.toml
    AIR_BIN := $(shell go env GOPATH)/bin/air.exe
else
    AIR_CONFIG := backend/.air.toml
    AIR_BIN := $(shell go env GOPATH)/bin/air
endif

E2E_DATABASE_URL := postgresql://testuser:testpassword@localhost:5432/e2e_test_db?sslmode=disable

## --- Installation ---
install:
	go install github.com/air-verse/air@latest
ifeq ($(OS),Windows_NT)
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
else
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin
endif
	go mod download
	cd frontend && npm ci
	cd e2e && npm ci

## Install Playwright dependencies
install-e2e: install
	cd e2e && npx playwright install --with-deps chromium


## --- Local Development ---
run-backend:
	$(AIR_BIN) -c $(AIR_CONFIG)

run-frontend:
	cd frontend && npm run dev

build-backend:
	go build -o backend/bin/server ./backend/cmd/server

build-frontend:
	cd frontend && npm run build

## --- Testing & Linting ---
test:
	go test -v -race ./...
	cd frontend && npm run test

lint:
	$(shell go env GOPATH)/bin/golangci-lint run
	cd frontend && npm run lint

## Run E2E tests (requires backend + frontend running)
e2e:
	cd e2e && npx playwright test

## --- Database (Supabase Local) ---
## Start Supabase local, apply migrations + seed + test users (ready to use)
db-up:
	@echo "Starting Supabase local..."
	supabase start
	@echo "Resetting database with migrations and seed..."
	supabase db reset
	@echo "Applying test users..."
	powershell -ExecutionPolicy Bypass -File scripts/seed-local.ps1
	@echo ""
	@echo "Database is ready!"
	@echo "  Studio:    http://127.0.0.1:54323"
	@echo "  API:       http://127.0.0.1:54321"
	@echo "  Database:  postgresql://postgres:postgres@127.0.0.1:54322/postgres"
	@echo ""
	@echo "Now run: make run-backend && make run-frontend"

## Reset database (migrations + seed + test users)
db-reset:
	@echo "Resetting database..."
	supabase db reset
	@echo "Applying test users..."
	powershell -ExecutionPolicy Bypass -File scripts/seed-local.ps1
	@echo "Database reset complete!"

## Stop Supabase local
db-down:
	@echo "Stopping Supabase local..."
	supabase stop
	@echo "Supabase stopped."

## --- Docker (Containerized Stack) ---
## Build and run backend + frontend in Docker (requires db-up first)
docker-up:
	@echo "Building and starting Docker containers..."
	docker-compose --env-file docker/.env.docker up --build

## Stop Docker containers
docker-down:
	@echo "Stopping Docker containers..."
	docker-compose --env-file docker/.env.docker down

## Build Docker images without starting
docker-build:
	@echo "Building Docker images..."
	docker-compose --env-file docker/.env.docker build

## Build and run backend + frontend in Docker (connects to REMOTE Supabase)
docker-up-remote:
	@echo "Building and starting Docker containers (REMOTE mode)..."
	docker-compose --env-file docker/.env.docker.remote up --build
