container_runtime := $(shell which podman 2>/dev/null || which docker 2>/dev/null)

ifneq ("$(wildcard .env)","")
	include .env
	export
endif

APP_NAME=builder
SUFFIX_SERVICE=builder
BIN_DIR=bin
MAIN_PATH=$(SUFFIX_SERVICE)/cmd/builder/main.go

.PHONY: build
build:
	go build -C $(SUFFIX_SERVICE) -o ../$(BIN_DIR)/$(APP_NAME) cmd/builder/main.go

.PHONY: run
run:
	go run -C $(SUFFIX_SERVICE) cmd/builder/main.go

.PHONY: unit-test
unit-test:
	make -C $(SUFFIX_SERVICE) unit-test

.PHONY: lint
lint:
	make -C builder lint
	make -C tests lint

.PHONY: db-gen
db-gen:
	make -C $(SUFFIX_SERVICE) db-gen

.PHONY: compose-up
compose-up:
	$(container_runtime) compose up -d db

.PHONY: compose-down
compose-down:
	$(container_runtime) compose down

.PHONY: compose-down-with_volumes
compose-down-with-volumes:
	$(container_runtime) compose down -v

.PHONY: install-tools
install-tools:
	@echo "Installing dependencies..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.12.2
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/k1LoW/tbls@latest
	sudo apt-get update && sudo apt-get install -y graphviz
	make install-tools-ci
	@echo "Tools installed successfully!"

.PHONY: gen-docs-database
gen-docs-database:
	@echo "Generating documentation for database..."
	tbls doc

.PHONY: db-migrate-create
db-migrate-create:
	migrate create -ext sql -dir $(SUFFIX_SERVICE)/internal/adapters/postgres/migrations -seq $(name)

.PHONY: run-tests
run-tests:
	${container_runtime} run --rm --network=host tests:latest