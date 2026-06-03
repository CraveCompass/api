.PHONY: dev-up dev-down dev-migrate

dev-up:
	docker compose --env-file .env -f deploy/docker/docker-compose.dev.yml up -d

dev-down:
	docker compose -f deploy/docker/docker-compose.dev.yml down

dev-migrate:
	docker exec -i docker-postgres-1 psql -U crave_compass_admin -d cravecompass_dev < deploy/database/001_init_schema.sql