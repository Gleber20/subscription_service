DC := docker compose

DB_USER := postgres
DB_PASS := Simuve39
DB_HOST := db
DB_PORT := 5432
DB_NAME := subscriptions_db
DB_URL  := postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: up down logs ps rebuild \
        migrate-up migrate-down migrate-reset migrate-version \
        db-psql db-tables

up:
	$(DC) up -d --build

down:
	$(DC) down

rebuild:
	$(DC) up -d --build --force-recreate

logs:
	$(DC) logs -f --tail=200

ps:
	$(DC) ps

# --- Migrations (always inside docker network) ---
migrate-up:
	$(DC) run --rm migrate up

migrate-down:
	$(DC) run --rm migrate down 1

migrate-reset:
	$(DC) run --rm migrate down -all
	$(DC) run --rm migrate up

migrate-version:
	$(DC) run --rm migrate version

# --- DB helpers ---
db-psql:
	$(DC) exec db psql -U $(DB_USER) -d $(DB_NAME)

db-tables:
	$(DC) exec db psql -U $(DB_USER) -d $(DB_NAME) -c "\dt"

swagger:
	swag init -g cmd/api/main.go -o docs