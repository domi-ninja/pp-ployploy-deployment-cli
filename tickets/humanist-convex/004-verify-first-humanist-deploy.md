# 004 - Verify First Humanist Deploy

## Status

Done.

## Scope

Run the first deployment to p3 and verify the public endpoints.

## Acceptance Criteria

- `https://humanist.design/` returns 200.
- `https://www.humanist.design/` redirects to `https://humanist.design`.
- `https://api.humanist.design/version` returns 200.
- `https://dash.humanist.design/` reaches the Convex dashboard auth flow.
- A cheap Humanist app flow works against the production Convex deployment.

## Verification Notes

Verified against p3 with release `20260703T152527Z-b139a1c`.

- `pp down` removed the Humanist project containers without deleting persistent Postgres or MinIO data.
- No-arg `pp` redeployed Humanist successfully.
- p3 containers were healthy after deploy:
  - `humanist-design-web-1`
  - `humanist-design-convex-backend-1`
  - `humanist-design-convex-dashboard-1`
  - `humanist-design-postgres-1`
  - `humanist-design-s3-1`
- Release metadata recorded all smoke checks as `ok`:
  - frontend
  - `www` redirect
  - Convex version
  - dashboard
  - Convex HTTP action
  - Convex query
  - Playwright landing page
