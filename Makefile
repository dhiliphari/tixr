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

## infra/up: start local infrastructure
.PHONY: infra/up
infra/up:
	docker compose up -d

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