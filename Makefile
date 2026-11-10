-include .env

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/catalog: run the catalog gRPC service
.PHONY: run/catalog
run/catalog:
	go run ./cmd/catalog -db-dsn="${TIXR_DB_DSN}"

## run/booking: run the booking gRPC service
.PHONY: run/booking
run/booking:
	go run ./cmd/booking -db-dsn="${TIXR_DB_DSN}"

## run/gateway: run the HTTP gateway
.PHONY: run/gateway
run/gateway:
	go run ./cmd/gateway

## infra/up: start local infrastructure and wait until it's healthy
.PHONY: infra/up
infra/up:
	docker compose up -d --wait

## infra/down: stop local infrastructure (keeps data)
.PHONY: infra/down
infra/down:
	docker compose down

## infra/destroy: stop local infrastructure and DELETE all data
.PHONY: infra/destroy
infra/destroy: confirm
	docker compose down -v

## db/psql: open a psql shell inside the postgres container
.PHONY: db/psql
db/psql:
	docker compose exec postgres psql -U tixr -d tixr

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Create migration files for ${name}...'
	migrate create -seq -ext sql -dir migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Running up migrations...'
	migrate -path migrations -database="${TIXR_DB_DSN}" up

## db/seed: DELETE all data and load sample events and seats
.PHONY: db/seed
db/seed: confirm
	docker compose exec -T postgres psql -q -v ON_ERROR_STOP=1 -U tixr -d tixr < scripts/seed.sql

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format all .go files and tidy module dependencies
.PHONY: tidy
tidy:
	@echo 'Formatting .go files...'
	go fmt ./...
	@echo 'Tidying module dependencies...'
	go mod tidy

## audit: run quality control checks
.PHONY: audit
audit:
	@echo 'Checking module dependencies...'
	go mod tidy -diff
	go mod verify
	@echo 'Checking formatting...'
	@test -z "$$(gofmt -l .)" || (echo 'Files need formatting:'; gofmt -l .; exit 1)
	@echo 'Linting protobuf files...'
	buf lint
	@echo 'Vetting code...'
	go vet ./...
	go tool staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## test: run unit tests
.PHONY: test
test:
	go test -race ./...

## test/integration: run unit and integration tests (requires Docker)
.PHONY: test/integration
test/integration:
	go test -race -tags=integration ./...

# ==================================================================================== #
# BUILD
# ==================================================================================== #

git_description = $(shell git describe --always --dirty --tags --long 2>/dev/null || echo dev)
linker_flags = '-s -X main.version=${git_description}'

## build/catalog: build the catalog service
.PHONY: build/catalog
build/catalog:
	@echo 'Building cmd/catalog...'
	go build -ldflags=${linker_flags} -o=./bin/catalog ./cmd/catalog

## build/booking: build the booking service
.PHONY: build/booking
build/booking:
	@echo 'Building cmd/booking...'
	go build -ldflags=${linker_flags} -o=./bin/booking ./cmd/booking

## build/gateway: build the HTTP gateway
.PHONY: build/gateway
build/gateway:
	@echo 'Building cmd/gateway...'
	go build -ldflags=${linker_flags} -o=./bin/gateway ./cmd/gateway

## build/all: build every service
.PHONY: build/all
build/all: build/catalog build/booking build/gateway

# ==================================================================================== #
# PROTOBUF
# ==================================================================================== #

## proto/lint: lint protobuf files
.PHONY: proto/lint
proto/lint:
	buf lint

## proto/gen: generate Go code from protobuf files
.PHONY: proto/gen
proto/gen:
	@echo 'Generating code from proto/ into gen/...'
	buf generate

## proto/breaking: check protobuf files for breaking changes against main
.PHONY: proto/breaking
proto/breaking:
	buf breaking --against '.git#branch=main'