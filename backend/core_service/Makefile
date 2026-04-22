.PHONY: test-integration test-integration-v test-integration-cover test-clean

test-integration:
	go test -v -tags=integration ./...

test-integration-v:
	go test -v -tags=integration -run $(TEST) ./...

test-integration-cover:
	go test -v -tags=integration -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-clean:
	go clean -testcache

test-all: test-clean test-integration