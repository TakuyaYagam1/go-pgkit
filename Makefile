FUZZTIME ?= 10s
ACTIONLINT_VERSION ?= v1.7.7

.PHONY: test test-race test-bench test-integration test-fuzz lint-actions fmt vet lint cover tidy

test:
	go test ./...

test-race:
	go test -race ./...

test-bench:
	go test -bench=. ./...

test-integration:
	go test -race -tags=integration -count=1 ./...

test-fuzz:
	go test ./postgres -run '^$$' -fuzz='^FuzzMaskURL$$' -fuzztime=$(FUZZTIME)
	go test ./pgutil -run '^$$' -fuzz='^FuzzPgErrorCode$$' -fuzztime=$(FUZZTIME)

lint-actions:
	go run github.com/rhysd/actionlint/cmd/actionlint@$(ACTIONLINT_VERSION)

fmt:
	gofmt -w .
	goimports -w .

vet:
	go vet ./...

lint:
	golangci-lint run --fix ./...

cover:
	go test -tags=integration -cover ./postgres ./pgutil ./migrator/goose ./migrator/migrate

tidy:
	go mod tidy
