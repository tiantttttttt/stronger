.PHONY: build test lint check seed-validate seed-sql env-init api-run db-up compose-up compose-down compose-logs swagger

COMPOSE = docker compose --env-file .env -f compose.yaml

env-init:
	@if [ ! -f .env ]; then \
		umask 077; \
		secret=$$(openssl rand -hex 32) || exit 1; \
		sed "s/^JWT_SECRET=.*/JWT_SECRET=$$secret/" .env.example > .env; \
		echo "Created .env with a random JWT secret"; \
	else echo "Keeping existing .env"; fi

build:
	GOCACHE=/tmp/stronger-go-build go build ./...

test:
	GOCACHE=/tmp/stronger-go-build go test ./...

lint:
	GOCACHE=/tmp/stronger-go-build go vet ./...

check: seed-validate test lint build

seed-validate:
	@node scripts/validate-phase1-catalog.mjs

seed-sql:
	@node scripts/render-phase1-catalog-seed.mjs

api-run: env-init
	@set -a; . ./.env; set +a; GOCACHE=/tmp/stronger-go-build go run .


db-up: env-init
	$(COMPOSE) up -d db

compose-up: env-init
	$(COMPOSE) up -d --build

compose-down:
	$(COMPOSE) down

compose-logs:
	$(COMPOSE) logs -f

SWAG ?= swag

swagger:
	$(SWAG) init -g main.go --parseInternal --outputTypes json,yaml -o docs
