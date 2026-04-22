.PHONY: test test-cover test-cover-html test-cover-filtered clean

test:
	go test ./... -v


test-cover:
	go test ./internal/presentation/... -cover -coverprofile=coverage_full.out -coverpkg=./internal/presentation/...
	@cat coverage_full.out | grep -v "mocks/" > coverage.out
	@go tool cover -func=coverage.out | grep total
	@echo "\nCoverage report saved to coverage.out"

test-cover-html: test-cover
	go tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage report saved to coverage.html"

test-verbose:
	go test ./internal/presentation/... -v -cover

test-single:
	go test ./internal/presentation/... -v -run $(TEST)


clean:
	rm -f coverage.out coverage.html coverage_full.out