# Load .env variables and export them

## help: print this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

## run/api: run the cmd/api application
run/api:
	@go run ./cmd/api -db-dsn="${DB_DSN}"

## db/psql: connect to the database using psql
db/psql:
	@psql "${DB_DSN}"

## db/migrations/up: apply all up database migrations
db/migrations/up: confirm
	@echo 'Running up migrations...'
	@migrate -path ./migrations -database "${DB_DSN}" up

## db/migrations/new name=$1: create a new database migration
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}


audit: vendor
	@echo 'Formatting code...'
	go fmt ./...

	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...

	@echo 'Running tests...'
	go test -race -vet=off ./...
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

current_time = $(shell date -u "+%Y-%m-%dT%H:%M:%SZ")
git_description = $(shell git describe --always --dirty --tags --long)
linker_flags ='-s -X main.buildTime=${current_time} -X main.version=${git_description}'

build/api:
	@echo 'Building cmd/api for local system...'
	go build -ldflags=${linker_flags} -o ./bin/api ./cmd/api

	@echo 'Building cmd/api for linux/amd64...'
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o ./bin/linux_amd64/api ./cmd/api

.PHONY: help confirm run/api db/psql db/migrations/up db/migrations/new audit build/api