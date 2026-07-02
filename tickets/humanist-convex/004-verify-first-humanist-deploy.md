# 004 - Verify First Humanist Deploy

## Scope

Run the first deployment to p3 and verify the public endpoints.

## Acceptance Criteria

- `https://humanist.design/` returns 200.
- `https://www.humanist.design/` redirects to `https://humanist.design`.
- `https://api.humanist.design/version` returns 200.
- `https://dash.humanist.design/` reaches the Convex dashboard auth flow.
- A cheap Humanist app flow works against the production Convex deployment.

