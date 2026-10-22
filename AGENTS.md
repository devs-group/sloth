# Sloth

Self-hosted PaaS: users configure container services in a web UI, sloth writes a docker-compose file per project and
runs it on the host's Docker, routed by Traefik. OAuth login via GitHub and Google.

- Backend: Go 1.23 (gin, sqlx), SQLite (modernc), goose migrations. Entry point `main.go` in the repo root.
- Frontend: Nuxt 3 SPA (`ssr: false`), PrimeVue, Tailwind, zod schemas. Served under `/_/`.
- Runtime: the backend runs inside a `docker:dind` image and talks to the host Docker through the mounted
  `/var/run/docker.sock`.

## Layout

- `main.go`, `embed_dev.go`, `embed_prod.go`: server setup. Build tag `production` embeds the generated frontend;
  without it the backend proxies `/_/` to `FRONTEND_HOST` (the Nuxt dev server).
- `backend/handlers`: gin handlers, routes under `/v1` and `/v1/auth`.
- `backend/services`: business logic and SQL. `services/project.go` and `services/service.go` build the compose
  files for deployed projects, including the Traefik labels.
- `backend/pkg/compose`, `backend/pkg/docker`: compose types and docker calls.
- `backend/tests`: integration tests, settings in `backend/tests/.env.test`.
- `database/migrations/static`: goose migrations, applied automatically on every app start.
- `frontend/`: `pages`, `components`, `composables`, `api/index.ts` (API client), `schema/schema.ts` (zod),
  `service-templates/` (presets like Postgres, SeaweedFS S3; MinIO was dropped because it no longer ships free images).
- `deployment/Dockerfile`: production image. `backend/Dockerfile`, `frontend/Dockerfile`, `mailpit/Dockerfile`: local dev.

## Local development

Run everything through the Makefile, `make help` lists the targets.

1. `cp .env.example .env` and fill in the OAuth keys (see Configuration).
2. `make up` creates the external `traefik` network if missing, builds on first run and waits until all services
   are healthy. Backend (air) and frontend (nuxt dev) hot-reload from the bind-mounted sources.
3. `make down`, `make restart` (recreate, re-reads `.env`), `make rebuild` (after Dockerfile, `go.mod` or
   `package.json` changes), `make ps`, `make logs s=backend`.

Only Traefik publishes a host port (80). Everything else is routed by hostname, `*.localhost` resolves to 127.0.0.1
without any `/etc/hosts` entry:

- App: http://localhost/_/ (must be `localhost` for the Google OAuth redirect URL)
- Mailpit: http://mail.sloth.localhost
- Traefik dashboard: http://traefik.sloth.localhost
- Services deployed by sloth: `http://<service-usn>.localhost`

Rules for the dev stack:

- Do not add host `ports:` to backend, frontend or mailpit. Traefik reaches them over the Docker network.
- The `traefik` network stays external and named exactly `traefik`: the composes sloth generates for deployed
  projects join it by that name.
- Compose project name is `sloth-dev`, containers are `sloth-dev-<service>-1`.
- `/var/app` is the work directory in every container on purpose. Docker Desktop shares it as a virtual mount, so
  docker-in-docker bind mounts resolve on macOS too.

## Configuration

`.env` (template `.env.example`). OAuth apps:

- GitHub: https://github.com/settings/developers (or the organisation's settings), callback
  `http://localhost/_/auth/callback/github`
- Google: Google Cloud project, https://console.cloud.google.com/auth/clients, callback
  `http://localhost/_/auth/callback/google`

## Quality

- `make test`: backend tests in the running backend container. Without the stack: `go test ./...` from the root.
- `make typecheck`: `nuxi typecheck` in the running frontend container. Prefer it over running nuxi on the host.
  The container keeps its own `.nuxt` (anonymous volume), so a host run no longer breaks the dev server, but
  without that volume a host `nuxi` rewrites `.nuxt` with host paths and Tailwind finds no classes.
- `make lint`, `make lint-fix`: frontend eslint in the running frontend container. Without the stack:
  `npm --prefix frontend run lint .`
- CI (`.github/workflows/release.yml`) runs Nuxt lint and `go test ./...` on every PR. Both must pass.

## Migrations

- Requires goose on the host: `go install github.com/pressly/goose/v3/cmd/goose@v3.18.0`
- Create: `make goose-static-create create_table_my_table`
- Never edit a migration that already ran somewhere, add a new one. Goose skips applied versions, so an edited
  file never reaches existing databases (`00009` got `usn` that way, local DBs from before 2025-04-29 lack it).
  Local fix for such a stale DB: `make down`, delete `database/database.sqlite`, `make up`, log in again.
- Migrations run on every app start, and air restarts the app on changes. The `goose-*` targets in the Makefile
  roll back and reapply against `database/database.sqlite`.

## Deployment

- A push to `main` builds `ghcr.io/devs-group/sloth/app:latest` from `deployment/Dockerfile` and calls the deploy
  webhook. So merging to `main` deploys to production.
- Production: https://sloth.podseidon.io, fake webmail https://mail.sloth.podseidon.io. Server reachable as
  `ssh sloth`.
- On the server: Traefik v3 in `/root/docker-compose.yml` with config in `/root/.traefik/`, sloth itself in
  `/root/sloth/` (`docker-compose.yaml`, `.env`, `webhook.sh`). The webhook (`deploy.podseidon.io`) runs
  `webhook.sh`: pull, down, up.
- Docker bypasses ufw. Every published `ports:` entry on the server is reachable from the internet, so services stay
  behind Traefik without host ports.
- Local test of the prod image: `docker build -f ./deployment/Dockerfile -t sloth/app:latest .`, then
  `docker run -it --env-file .env -v /var/run/docker.sock:/var/run/docker.sock -p 9090:9090 --rm sloth/app:latest`.
  The Dockerfile runs `npm run generate:prod`, which points at production; switch it to `npm run generate`
  temporarily for a local test.
