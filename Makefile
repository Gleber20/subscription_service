POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DATABASE=subscriptions_db

DB_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DATABASE)?sslmode=disable

.PHONY: up down run migrate-up migrate-down migrate-reset migrate-create

up:
	docker compose up -d

down:
	docker compose down

run:
	go run ./cmd/api

migrate-up:
	migrate -path ./migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DB_URL)" down 1

migrate-reset:
	migrate -path ./migrations -database "$(DB_URL)" down -all

migrate-create:
	migrate create -ext sql -dir ./migrations -seq -digits 4 $(name)

swagger:
	swag init -g cmd/api/main.go -o docs