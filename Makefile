.PHONY: fmt vet test backend-run cli-run

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
