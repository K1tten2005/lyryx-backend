docs:
	swag init --parseDependency --parseInternal -g cmd/service/main.go

database-up:
	docker compose up -d postgres

migrate:
	goose -dir db/migrations postgres "host=localhost user=user password=password dbname=lyryx sslmode=disable" up