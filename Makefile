.PHONY: install run-backend run-frontend build-backend build-frontend test lint e2e

ifeq ($(OS),Windows_NT)
    AIR_CONFIG := backend/.air.windows.toml
    AIR_BIN := $(shell go env GOPATH)/bin/air.exe
else
    AIR_CONFIG := backend/.air.toml
    AIR_BIN := $(shell go env GOPATH)/bin/air
endif


## Install all dependencies
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

## Run backend with hot reload
run-backend:
	$(AIR_BIN) -c $(AIR_CONFIG)
	
## Run frontend dev server
run-frontend:
	cd frontend && npm run dev

## Build backend binary
build-backend:
	go build -o backend/bin/server ./backend/cmd/server

## Build frontend for production
build-frontend:
	cd frontend && npm run build

## Run all tests
test:
	go test -v -race ./...
	cd frontend && npm run test

## Run linters
lint:
	$(shell go env GOPATH)/bin/golangci-lint run
	cd frontend && npm run lint

## Run E2E tests (requires backend + frontend running)
e2e:
	cd e2e && npx playwright test
