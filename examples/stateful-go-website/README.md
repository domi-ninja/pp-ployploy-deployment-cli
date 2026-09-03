# Stateful Go website

Use this example for a small server-rendered application that needs durable local data and database migrations. `pp` builds one Go image, mounts a named volume at `/data`, passes the required secret, checks `/health`, and publishes the service through Caddy.

The example contains:

- [`deploy.yml`](deploy.yml) for the image, secret, volume, health check, and route
- [`Dockerfile`](Dockerfile) for the Go build and runtime image
- [`docker-entrypoint.sh`](docker-entrypoint.sh) for copying packaged assets and applying SQLite migrations before startup
- [`app.prod.toml`](app.prod.toml) for runtime defaults
- [`.env.example`](.env.example) for the required `JWT_SECRET`

Copy the files into the application repository, replace `deploy@example.com` and `stateful.example.com`, then create `.env` with a real `JWT_SECRET`. The Dockerfile expects the Go modules, `cmd/server`, frontend assets, and Goose migrations from the application.

Run `pp plan` to catch missing config or environment values before `pp deploy`.
