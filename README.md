# creator-portfolio

![creator-portfolio](.github/screenshot.png)

Personal portfolio for a photographer & videographer: a public site plus an `/admin` panel for managing media. The backend is a Go/Gin API with SQLite storage; the frontend is a Nuxt 3 (Vue, SSR) app.

## Architecture

The monorepo holds two apps: `apps/api` (Go/Gin REST API, SQLite + uploaded media under `DATA_DIR`, JWT auth for the admin panel) and `apps/site` (Nuxt 3 SSR frontend). The site proxies `/api/**` and `/media/**` to the api via nitro routeRules, so browsers only ever talk to the site origin.

Request flow: browser → site (:3000) → `/api/*`, `/media/*` proxied → api (:8080) → SQLite + media dir.

## Requirements

- Go 1.25+
- Node.js 22
- Docker with Compose v2 (for the containerized stacks)
- [Task](https://taskfile.dev) (task runner)

## Quickstart

```sh
task setup                                                    # install go modules + npm packages
cp deploy/dev/docker/.env.example deploy/dev/docker/.env      # optional: dev defaults work as-is
task dev                                                      # hot-reload stack: air + nuxt dev server
```

Open http://localhost:3000 (admin panel at `/admin`).

For a production-like stack with built images:

```sh
cp deploy/local/.env.example deploy/local/.env   # then edit the secrets inside
task local
```

## Ports

| Service | Port                                      |
| ------- | ----------------------------------------- |
| site    | 3000                                      |
| api     | 8080 (`GET /api/health` for healthchecks) |

## Environment variables

| Variable               | Default                                                     | Where set                          | Purpose                                                                                    |
| ---------------------- | ----------------------------------------------------------- | ---------------------------------- | ------------------------------------------------------------------------------------------ |
| `PORT`                 | `8080`                                                      | compose `environment`              | api listen port                                                                            |
| `DATA_DIR`             | `./data` on host, `/data` in containers                     | compose `environment`              | SQLite db + uploaded media                                                                 |
| `GIN_MODE`             | `debug` (`release` in the local stack)                      | compose `environment`              | gin mode; `release` forbids default secrets                                                |
| `ADMIN_USER`           | `admin`                                                     | `deploy/*/…/.env`                  | admin login                                                                                |
| `ADMIN_PASSWORD`       | `admin` (dev default)                                       | `deploy/*/…/.env`                  | admin password                                                                             |
| `JWT_SECRET`           | `dev-secret-change-me` (dev default)                        | `deploy/*/…/.env`                  | HMAC secret for session tokens                                                             |
| `COOKIE_SECURE`        | `false`                                                     | `deploy/*/…/.env`                  | set `true` when serving over https                                                         |
| `NUXT_API_PROXY`       | `http://localhost:8080` (host), `http://api:8080` (compose) | compose `environment`, build `ARG` | proxy target for `/api` and `/media`; baked in at build time for the production site image |
| `NUXT_PUBLIC_SITE_URL` | `http://localhost:3000`                                     | environment                        | public origin of the site, used for canonical/og URLs                                      |

Dev-stack variables live in `deploy/dev/docker/.env` (optional, see `.env.example` next to it); local-stack variables in `deploy/local/.env`.

## Development

Host-side checks (no Docker needed):

```sh
task api:test          # go test ./...
task api:vet           # go vet ./...
task api:lint          # golangci-lint run (if installed)
task site:lint         # eslint
task site:test         # vitest
task site:typecheck    # vue-tsc
task site:build        # nuxt build
task test              # all of the above (vet + go test + eslint + vitest + typecheck + build)
```

You can also run the apps directly on the host: `task api:run` and `task site:dev`.

## Project structure

```
├── apps/
│   ├── api/            Go/Gin REST API (SQLite storage, JWT auth for /admin)
│   └── site/           Nuxt 3 frontend (SSR; proxies /api and /media to the api)
├── deploy/
│   ├── dev/            hot-reload stack (air + nuxt dev server)
│   │   ├── compose.yaml
│   │   ├── docker/     dockerfiles + optional .env (.env.example provided)
│   │   └── data/       SQLite db + uploaded media (gitignored)
│   └── local/          production-like stack (built images, non-root, healthchecks)
│       ├── compose.yaml
│       ├── docker/     dockerfiles + .env.example
│       └── data/       SQLite db + uploaded media (gitignored)
├── Taskfile.yaml       task entry points (task dev, task local, task test, …)
└── .github/
```

## Release notes

Before deploying with `GIN_MODE=release`:

- Set `ADMIN_PASSWORD` and `JWT_SECRET` to non-default values — the api refuses to start otherwise. `JWT_SECRET` must be at least 32 characters long; this is enforced in release mode.
- Set `COOKIE_SECURE=true` when serving over https.
