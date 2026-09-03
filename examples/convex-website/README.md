# Convex website

Use this example when you want to self-host a Convex application on one Docker server. It deploys the website, Convex backend and dashboard, Postgres, and MinIO. The hooks create the object-storage buckets, refresh the Convex admin key, push application environment variables, and sync the Convex functions.

The example contains:

- [`deploy.yml`](deploy.yml) for the services, build, volumes, hooks, and smoke checks
- [`Dockerfile`](Dockerfile) and [`nginx.conf`](nginx.conf) for the website
- [`deploy/caddy/convex-website.caddy.tmpl`](deploy/caddy/convex-website.caddy.tmpl) for the site, API, and dashboard routes
- [`scripts/`](scripts/) for production environment setup and post-start deployment work
- [`.env.local.example`](.env.local.example) as a list of the application values the hooks may need

Copy the files into the application repository, replace `deploy@example.com` and every `convex.example.com` hostname, then generate the base production environment file:

```sh
bash scripts/bootstrap-prod-env.sh .env.prod
```

Add the application-specific secrets and credentials listed in `.env.local.example`. Keep `.env.prod` out of Git. Run `pp plan` before the first `pp deploy`, since this example has more moving parts than most side projects deserve.
