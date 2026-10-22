# Sloth - Root Makefile

.PHONY: help up down restart rebuild ps urls logs test lint lint-fix typecheck network

# Default target
help:
	@echo "Sloth Commands"
	@echo ""
	@echo "Development:"
	@echo "  make up                 - Start the stack (traefik + backend + frontend + mailpit)"
	@echo "  make down               - Stop the stack"
	@echo "  make restart            - Recreate all services (picks up .env changes)"
	@echo "  make rebuild            - Rebuild images, then start (after Dockerfile or go.mod/package.json changes)"
	@echo "  make ps                 - Show container status and local URLs"
	@echo "  make urls               - Show local URLs only"
	@echo "  make logs [s=backend]   - Follow docker logs (all services, or one)"
	@echo ""
	@echo "Quality:"
	@echo "  make test               - Backend tests (in the running backend container)"
	@echo "  make lint               - Frontend eslint (in the running frontend container)"
	@echo "  make lint-fix           - Frontend eslint --fix"
	@echo "  make typecheck          - Frontend nuxi typecheck (in the running frontend container)"
	@echo ""
	@echo "Migrations:"
	@echo "  make goose-static-create <type>_<name> - Create a migration (eg. create_table_my_table)"
	@echo "  make goose-static-up | goose-static-down-1 | ...  - See the goose-* targets below"

# =============================================================================
# Development
# =============================================================================

# The composes sloth generates for deployed projects join the external network
# "traefik" by name, so it must exist before anything starts.
network:
	@docker network inspect traefik >/dev/null 2>&1 || docker network create traefik >/dev/null

up: network
	@test -f .env || { echo ".env missing: cp .env.example .env and fill in the OAuth keys"; exit 1; }
	@docker compose up -d --wait
	@echo ""
	@echo "Stack started:"
	@$(MAKE) --no-print-directory urls

down:
	@docker compose down

# Recreate rather than restart: `docker compose restart` reuses the old
# container and does NOT re-read env_file changes.
restart: network
	@docker compose up -d --force-recreate --wait
	@$(MAKE) --no-print-directory urls

rebuild: network
	@docker compose up -d --build --wait
	@$(MAKE) --no-print-directory urls

ps:
	@docker compose ps --format "table {{.Name}}\t{{.Status}}"
	@echo ""
	@$(MAKE) --no-print-directory urls

# Single source of truth for the local URLs. Only Traefik holds a host port (80);
# *.localhost resolves to 127.0.0.1 without any /etc/hosts entry.
urls:
	@echo "  App:       http://localhost/_/"
	@echo "  Mailpit:   http://mail.sloth.localhost"
	@echo "  Traefik:   http://traefik.sloth.localhost"
	@echo "  Deployed:  http://<service-usn>.localhost"

# Optional single service: make logs s=backend
logs:
	@docker compose logs -f $(s)

# =============================================================================
# Quality
# =============================================================================

test:
	@docker compose exec backend go test ./...

lint:
	@docker compose exec frontend npm run lint .

lint-fix:
	@docker compose exec frontend npm run lint -- --fix .

typecheck:
	@docker compose exec frontend npx nuxi typecheck

# =============================================================================
# Migrations (goose, run on the host)
# =============================================================================

goose-static-create:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-static-create <type_of_creation>_<name_of_migration> (eg. create_table_my_table)"; \
	else \
		goose -s --dir database/migrations/static create $(word 2, $(MAKECMDGOALS)) sql; \
	fi
goose-static-up:
	goose --dir database/migrations/static sqlite3 database/database.sqlite up
goose-dynamic-up:
	goose -no-versioning --dir database/migrations/dynamic sqlite3 database/database.sqlite up
goose-static-up-1:
	goose --dir database/migrations/static sqlite3 database/database.sqlite up-by-one
goose-dynamic-up-1:
	goose -no-versioning --dir database/migrations/dynamic sqlite3 database/database.sqlite up-by-one
goose-static-down-1:
	goose --dir database/migrations/static sqlite3 database/database.sqlite down
goose-dynamic-down-1:
	goose -no-versioning --dir database/migrations/dynamic sqlite3 database/database.sqlite down
goose-static-down-to-0:
	goose --dir database/migrations/static sqlite3 database/database.sqlite down-to 0
goose-dynamic-down-to-0:
	goose -no-versioning --dir database/migrations/dynamic sqlite3 database/database.sqlite down-to 0
goose-static-down-to:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to <version_number>"; \
	else \
		goose --dir database/migrations/static sqlite3 database/database.sqlite down-to $(word 2, $(MAKECMDGOALS)); \
	fi
goose-dynamic-down-to:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to <version_number>"; \
	else \
		goose -no-versioning --dir database/migrations/dynamic sqlite3 database/database.sqlite down-to $(word 2, $(MAKECMDGOALS)); \
	fi
goose-static-down-to-and-up:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to-and-up <version_number>"; \
	else \
		goose --dir database/migrations/static sqlite3 database/database.sqlite down-to $(word 2, $(MAKECMDGOALS)); \
		goose --dir database/migrations/static sqlite3 database/database.sqlite up; \
	fi
goose-dynamic-down-to-and-up:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to-and-up <version_number>"; \
	else \
		goose -no-versioning --dir database/migrations/dynamic sqlite3 database/database.sqlite down-to $(word 2, $(MAKECMDGOALS)); \
		goose -no-versioning --dir database/migrations/dynamic sqlite3 database/database.sqlite up; \
	fi
# Swallows the extra word in `make goose-static-create <name>` so make does not
# treat it as a target.
%:
	@: