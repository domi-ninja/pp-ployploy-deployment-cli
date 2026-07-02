# 003 - Wire p3 Convex Runtime Services

## Scope

Model the p3 runtime services for Humanist:

- Convex backend on ports `3210` and `3211`
- Convex dashboard on port `3211`
- Postgres with p3 persistent storage
- S3-compatible storage container with p3 persistent storage

## Acceptance Criteria

- Runtime data lives outside release directories.
- Convex backend receives Postgres and S3 env vars.
- The dashboard remains public and protected by Convex master-key auth.
- S3 admin exposure is explicit and not accidentally public.

