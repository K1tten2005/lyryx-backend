docs:
	swag init --parseDependency --parseInternal -g ./cmd/app/main.go

app-up:
	docker compose up -d 
	
migrate:
	goose -dir db/migrations postgres "host=localhost user=user password=password dbname=lyryx sslmode=disable" up
