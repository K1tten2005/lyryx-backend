ifneq (,$(wildcard ./.env))
    include .env
    export
endif

docs:
	swag init --parseDependency --parseInternal -g ./cmd/app/main.go

app-up:
	docker compose up -d 
	
migrate:
	goose -dir db/migrations postgres "host=localhost user=user password=password dbname=lyryx sslmode=disable" up
	PGPASSWORD=$(PG_PASSWORD) psql -h $(PG_HOST) -p $(PG_PORT) -U $(PG_USER) -d $(PG_DB) -f data/inserts.txt

setup-ai:
	docker exec -it lyryx_ollama ollama pull $(OLLAMA_MODEL_NAME)

clean-ai:
	docker exec lyryx_ollama sh -c "for m in \$$(ollama list | tail -n +2 | awk '{print \$$1}'); do ollama rm \$$m; done"