.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: db-gen
db-gen:
	sqlc generate

.PHONY: db-gen
unit-test:
	go test -v -race ./...