APP_NAME=ticketstream

.PHONY: tidy run-api run-worker test-unit test-integration test-e2e test-load

tidy:
	go mod tidy

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

test-unit:
	go test ./...

test-integration:
	go test -tags=integration ./...

test-e2e:
	go test -tags=e2e ./tests/e2e/...

test-load:
	k6 run ./tests/load/flash_sale.js
