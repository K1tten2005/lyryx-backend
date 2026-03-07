docs:
	swag init --parseDependency --parseInternal -g cmd/app/main.go

database-up:
	docker compose up -d postgres

migrate:
	goose -dir db/migrations postgres "host=localhost user=user password=password dbname=lyryx sslmode=disable" up

database-down:
	docker compose down -v