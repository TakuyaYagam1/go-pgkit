.PHONY: test test-race test-bench test-integration fmt vet cover tidy

test:
	go test ./...

test-race:
	go test -race ./...

test-bench:
	go test -bench=. ./...

test-integration:
	go test -race -tags=integration -count=1 ./...

fmt:
	gofmt -w .
	goimports -w .

vet:
	go vet ./...

cover:
	go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

tidy:
	go mod tidy
