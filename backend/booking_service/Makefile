# Makefile в корне проекта

.PHONY: test
test:
	go test ./... -v

.PHONY: test-unit
test-unit:
	go test ./internal/... -v

.PHONY: test-integration
test-integration:
	go test ./tests/integration/... -v -tags=integration

.PHONY: test-coverage
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

.PHONY: test-race
test-race:
	go test ./... -race -v