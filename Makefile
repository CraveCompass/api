.PHONY: dev-up dev-down

dev-up:
	docker compose -f deploy/docker/docker-compose.dev.yml up -d

dev-down:
	docker compose -f deploy/docker/docker-compose.dev.yml down