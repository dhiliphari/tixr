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

## run/tixr: run the cmd/tixr application
.PHONY: run/tixr
run/tixr:
	go run ./cmd/tixr -db-dsn="${TIXR_DB_DSN}"

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
	@echo 'Vetting code...'
	go vet ./...
	go tool staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

# ==================================================================================== #
# BUILD
# ==================================================================================== #

git_description = $(shell git describe --always --dirty --tags --long 2>/dev/null || echo dev)
linker_flags = '-s -X main.version=${git_description}'

## build/tixr: build the cmd/tixr application
.PHONY: build/tixr
build/tixr:
	@echo 'Building cmd/tixr...'
	go build -ldflags=${linker_flags} -o=./bin/tixr ./cmd/tixr