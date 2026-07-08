# Gitea + Reticulum

Gitea fork with [Reticulum](https://reticulum.network/) support: host git repositories over `rns://` via rngit, with admin tooling for permissions, mirroring, and NomadNet.

Based on upstream Gitea. Git push/pull/clone works without a browser, the web UI is built to work without JavaScript for core flows (browse repos, forms, star/watch, dashboard repo list).

## Clone over RNS

```text
git clone rns://<destination-hash>/<owner>/<repo>
```

Destination hash and rngit settings: **Admin → Config → Reticulum** (`/-/admin/reticulum`).

## Docker

```bash
cp .env.example .env   # edit SECRET_KEY before production use
docker compose up -d --build
```

Image is also published to `ghcr.io/quad4-software/gitea-reticulum`.

Manual release (binaries + GHCR image): **Actions → release-manual → Run workflow**.

## Build from source

```bash
make deps-frontend deps-backend
TAGS=bindata make build
./gitea web
```

See `docs/build-setup.md` and `docs/development.md` for a full dev environment.

## Configuration

`[reticulum]` in `app.ini` or environment variables (`GITEA__reticulum__ENABLED=true`, etc.). Enable the built-in rngit server for a single-container setup.

## License

MIT (same as upstream Gitea).
