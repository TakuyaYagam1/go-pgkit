.PHONY: test test-race fmt vet

test:
	go test ./...

test-race:
	go test -race ./...

fmt:
	gofmt -w .
	goimports -w .

vet:
	go vet ./...
