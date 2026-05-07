.PHONY: install run-backend run-frontend build-backend build-frontend test lint e2e \
       local-up local-down local-seed local-reset local-status

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

## --- Docker & Database ---
db-up:
	@echo "Starting database container..."
	docker compose up -d db

db-down:
	@echo "Stopping database container..."
	docker compose down -v

db-seed:
	@echo "Waiting for database to be ready..."
	@powershell -ExecutionPolicy Bypass -Command "$$ready = $$false; while (-not $$ready) { Write-Host -NoNewline '.'; Start-Sleep -Seconds 1; $$result = (docker exec focuscafe-project-db-1 pg_isready -U testuser -d e2e_test_db); if ($$result -match 'accepting connections') { $$ready = $$true } }"
	@echo.
	@echo "Database ready. Running migrations and seeding..."
	@set DATABASE_URL=$(E2E_DATABASE_URL)&& go run ./backend/cmd/seed/main.go

## --- E2E Orchestration ---
e2e: db-up db-seed
	@echo "Starting backend for E2E tests..."
	@powershell -ExecutionPolicy Bypass -Command "$$proc = Start-Process -FilePath 'go' -ArgumentList 'run', './backend/cmd/server/main.go' -NoNewWindow -PassThru -WorkingDirectory '.' -Environment @{ DATABASE_URL = '$(E2E_DATABASE_URL)'; PORT = '8080'; GIN_MODE = 'release' }; $$proc.Id | Out-File -FilePath 'backend.pid' -Encoding utf8"
	
	@echo "Starting frontend for E2E tests..."
	@powershell -ExecutionPolicy Bypass -Command "$$proc = Start-Process -FilePath 'npm' -ArgumentList 'run', 'dev' -NoNewWindow -PassThru -WorkingDirectory 'frontend'; $$proc.Id | Out-File -FilePath 'frontend.pid' -Encoding utf8"

	@echo "Waiting for backend (port 8081)..."
	@powershell -ExecutionPolicy Bypass -Command "$$ready = $$false; while (-not $$ready) { Write-Host -NoNewline '.'; Start-Sleep -Seconds 1; try { $$resp = Invoke-WebRequest -Uri http://localhost:8080/health -UseBasicParsing -ErrorAction SilentlyContinue; if ($$resp.StatusCode -eq 200) { $$ready = $$true } } catch {} }"
	
	@echo "Waiting for frontend (port 5173)..."
	@powershell -ExecutionPolicy Bypass -Command "$$ready = $$false; while (-not $$ready) { Write-Host -NoNewline '.'; Start-Sleep -Seconds 1; try { $$resp = Invoke-WebRequest -Uri http://localhost:5173 -UseBasicParsing -ErrorAction SilentlyContinue; if ($$resp.StatusCode -eq 200) { $$ready = $$true } } catch {} }"

	@echo "Running Playwright tests..."
	cd e2e && npx playwright test

# ========================================
# Local development environment (Supabase)
# ========================================

## Start local Supabase stack
local-up:
	supabase start
	@echo ""
	@echo "Local Supabase is running!" 
	@echo "  Studio:    http://127.0.0.1:54323"
	@echo "  API:       http://127.0.0.1:54321"
	@echo "  Database:  postgresql://postgres:postgres@127.0.0.1:54322/postgres"
	@echo ""
	@echo "Run 'make local-seed' to create test users."

## Stop local Supabase stack
local-down:
	supabase stop

## Reset local database (re-apply migrations + seed.sql) and create test users
local-reset:
	supabase db reset
	@echo "Applying test users..."
	powershell -ExecutionPolicy Bypass -File scripts/seed-local.ps1

## Create test users (requires backend NOT running, only Supabase)
local-seed:
	powershell -ExecutionPolicy Bypass -File scripts/seed-local.ps1

## Show Supabase local status and credentials
local-status:
	supabase status
