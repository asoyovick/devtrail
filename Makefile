DATABASE_URL ?= postgres://devtrail:devtrail@localhost:5432/devtrail?sslmode=disable
MIGRATIONS_DIR ?= backend/migrations

.PHONY: fmt vet test backend-run cli-run migrate-up migrate-down

fmt:
	cd backend && gofmt -w .
	cd cli && gofmt -w .

vet:
	cd backend && go vet ./...
	cd cli && go vet ./...

test:
	cd backend && go test ./...
	cd cli && go test ./...

backend-run:
	cd backend && go run ./cmd/api

cli-run:
	cd cli && go run ./cmd/devtrail -- --help

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1
