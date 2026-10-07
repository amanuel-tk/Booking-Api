include .env
export

DB_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/$(POSTGRES_DB)?sslmode=disable
MIGRATIONS_DIR=migrations

.PHONY: run db-up db-down db-psql migrate-create migrate-up migrate-down

run:
	air

db-up:
	docker compose up -d db

db-down:
	docker compose down

db-psql:
	docker compose exec db psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1