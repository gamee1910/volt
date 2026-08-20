DB_URL=postgres://admin:password@localhost:5432/volt?sslmode=disable
MIGRATIONS=./migrations

.PHONY: migrate-create migrate-up migrate-down migrate-down-one migrate-version migrate-force

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=create_users_table"; \
		exit 1; \
	fi
	migrate create -ext sql -dir $(MIGRATIONS) -seq $(name)

migrate-up:
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" down

migrate-down-one:
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" down 1

migrate-version:
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" version

migrate-force:
	@if [ -z "$(version)" ]; then \
		echo "Usage: make migrate-force version=3"; \
		exit 1; \
	fi
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" force $(version)

gen-docs:
	swag init -g cmd/api/main.go -o docs --parseInternal