.PHONY: setup init

setup:
	go install github.com/cosmtrek/air@latest
	cp .env.example .env
	make init

init:
	go install github.com/rubenv/sql-migrate/...@latest
	docker-compose up -d postgres
	@echo "Waiting for database connection..."
	@while ! docker exec kaivest-postgres pg_isready > /dev/null; do \
		sleep 1; \
	done
	make migrate-up

remove-infras:
	docker-compose down --remove-orphans

dev:
	go run ./cmd/*.go

air:
	air -c .air.toml
