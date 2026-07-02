# 002 - Add Humanist Deployment Files

## Scope

Add deployment artifacts to `../humanist.design`:

- production Dockerfile
- nginx SPA fallback config
- `deploy.yml`
- Caddy route template
- env example files
- safer Convex env push script behavior

## Acceptance Criteria

- `pp plan` can parse the Humanist deploy config once the `pp` extensions exist.
- Frontend build uses public Vite build args only.
- Convex sync uses `npx convex dev --once --env-file .env.local`.
- `www.humanist.design` redirects to the apex hostname.

