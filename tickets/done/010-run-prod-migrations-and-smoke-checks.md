# 010 - Run Prod Migrations And Smoke Checks

## Goal

Run explicit prod DB migrations and post-deploy smoke checks safely.

## Depends On

- 003
- 009

## Scope

- Run configured migration command inside the built app image before service update by default.
- Load migration env from deploy-time env provisioning.
- Validate that a rollback command or restore plan exists before running a prod migration.
- Abort deployment if migration fails.
- Run configured smoke checks after compose apply.
- Support HTTP status checks and command checks.

## Deliverables

- Migration runner.
- Migration rollback runner.
- Smoke check runner.
- Timeout handling.
- Metadata fields for migration and smoke check results.

## Acceptance Criteria

- Local development DB is not part of deploy config execution.
- Prod DB migration is explicit and logged.
- Prod DB migration runs inside the deployment image.
- Failed migration prevents service changes.
- Failed smoke check marks release failed and points to rollback.
- DB rollback metadata is recorded for `deploy rollback`.

## Completion Notes

Completed against `../humanist.design` on p3.

- `pp` supports HTTP smoke checks, command smoke checks, redirect assertions, check timeouts, and smoke-check metadata.
- Migration and hook timeout fields are enforced.
- Migration metadata records command, image, env source, timing, and error state.
- Humanist deploy uses `.env.prod` for production deploy input and `.env.local` for local dev.
- Humanist production deploy passed frontend, redirect, Convex version, dashboard, HTTP action, Convex query, and Playwright landing-page smoke checks.

## Humanist Test Plan

Use `../humanist.design` as the concrete acceptance test for this ticket.

The important environment rule is stricter than the original Humanist plan:

- Dev always happens locally.
- Dev env vars live in `../humanist.design/.env.local`.
- Prod deploy env vars live in `../humanist.design/.env.prod`.
- `.env.prod` is uncommitted and must be ignored by the Humanist repo.
- No production deploy command may read `.env.local`.
- No local dev command may mutate production Convex, production Postgres, or production S3.

This test intentionally separates all stateful services by environment:

- Dev Convex: selected by `.env.local`; expected to point at a local/self-hosted dev Convex instance.
- Prod Convex: selected by `.env.prod`; expected to point at `https://api.humanist.design`.
- Dev S3: selected by `.env.local`; expected to be local MinIO or another local S3-compatible endpoint.
- Prod S3: selected by `.env.prod`; expected to be the p3 MinIO service used only by production Convex.
- Dev Postgres: selected by `.env.local`; local only.
- Prod Postgres: selected by `.env.prod`; p3 only.

## Humanist Fixture Changes

Update `../humanist.design` so it proves the ticket behavior instead of relying on operator discipline.

1. Change `deploy.yml`.
   - Set top-level `env.source` to `.env.prod`.
   - Change every service `env.source` from `.env.local` to `.env.prod`.
   - Change every hook `env.source` from `.env.local` to `.env.prod`.
   - Make hook commands use the renamed Convex env script with the prod selector, for example `["bash", "scripts/push-env-to-convex.sh", "--prod"]`.
   - Keep frontend build args limited to public `VITE_*` values loaded from `.env.prod`.
   - Keep Convex backend, Postgres, and S3 secrets out of build args.

2. Update env bootstrapping.
   - Change `scripts/bootstrap-prod-env.sh` to default to `.env.prod`.
   - Keep support for an explicit path argument so tests can write to a temp env file.
   - Ensure generated prod values use production origins:
     - `VITE_SITE_URL=https://humanist.design`
     - `VITE_CONVEX_URL=https://api.humanist.design`
     - `VITE_CONVEX_SITE_URL=https://api.humanist.design`
     - `SITE_URL=https://humanist.design`
     - `CONVEX_SELF_HOSTED_URL=https://api.humanist.design`
     - `S3_ENDPOINT_URL=http://s3:9000` for the backend container
   - Do not generate or copy any dev Convex admin key into `.env.prod`.
   - If `CONVEX_SELF_HOSTED_ADMIN_KEY` is missing during prod deploy, refresh it from the running prod Convex backend, add it to `.env.prod`, and persist it locally.

3. Update Convex env sync.
   - Rename `scripts/push-.env.local-to-convex.sh` to `scripts/push-env-to-convex.sh`.
   - Default the script to dev behavior using `.env.local`.
   - Add a `--prod` flag that selects `.env.prod` and production Convex.
   - Keep an explicit `--env-file` override for tests and one-off maintenance.
   - Run Convex CLI with an explicit env file: `npx convex dev --once --env-file .env.prod`.
   - Validate that `.env.prod` contains `CONVEX_SELF_HOSTED_URL` and `CONVEX_SELF_HOSTED_ADMIN_KEY` before sync, after the refresh-if-missing step has run.
   - Do not prune production Convex env vars unless a dedicated `--prune` flag is present.

4. Update S3 bucket setup.
   - Make `scripts/ensure-minio-buckets.sh` read `.env.prod` during deploy.
   - Keep dev bucket setup separate, either by explicit `ENV_FILE=.env.local` or by a separate dev script.
   - Bucket names should use a systematic environment suffix or prefix, for example `humanist-dev-*` and `humanist-prod-*`, so accidental cross-env usage is obvious.
   - Bucket creation must run after the prod S3 service is healthy and before Convex function sync.
   - Do not expose the prod MinIO web UI through Caddy for this ticket.
   - Add a script for SSH forwarding the prod MinIO web UI when admin access is needed.

5. Update ignore and examples.
   - Add `.env.prod` to `../humanist.design/.gitignore`.
   - Keep `.env.local` ignored by the existing `*.local` rule.
   - Document that `.env.local` is local dev only and `.env.prod` is deploy input only.
   - Do not add `.env.prod.example` for this pass; the required key list in `deploy.yml` is enough.

## `pp` Implementation Details

Some primitives already exist in `pp`, but ticket 10 should finish and harden them.

1. Env source enforcement.
   - Validate every `env.source`, migration env source, and hook env source before deployment.
   - Include the env source path in plan output so a Humanist prod deploy visibly says `.env.prod`.
   - Do not add a generic `pp` policy that hard-rejects `.env.local`; environment naming policy stays with each project config.

2. Migration execution.
   - Keep migrations explicit under `migrations`.
   - Run migrations before service updates by default.
   - Run migrations inside the built app image for app-owned database migrations.
   - For Humanist, treat the pre-frontend backend setup sequence as deploy hooks rather than DB migrations:
     - start infra/backend phases
     - ensure prod S3 buckets
     - refresh prod Convex admin key if needed
     - push prod Convex env
     - sync prod Convex functions/schema
     - start frontend

3. Smoke checks.
   - Support HTTP checks with expected status.
   - Add command checks if still missing from implementation.
   - For Humanist prod, smoke checks should include:
     - `https://humanist.design/` returns `200`
     - `https://www.humanist.design/` redirects to `https://humanist.design`
     - `https://api.humanist.design/version` returns `200`
     - `https://dash.humanist.design/` reaches the expected dashboard response
     - an HTTP action smoke check
     - a Convex query smoke check
     - a tiny Playwright smoke check that loads the landing page and fails on obvious JavaScript console/page crashes

4. Metadata.
   - Record migration status, command, image, env source path, start time, end time, and error.
   - Record rollback command or restore plan.
   - Record smoke check group, each check name, target, expected result, actual result, duration, and error.
   - If a smoke check fails, mark the release failed and keep enough metadata for `deploy rollback` to explain what should happen next.

5. Failure behavior.
   - Failed migration aborts before services change.
   - Failed Humanist backend setup hook aborts before the frontend phase starts.
   - Failed smoke check after service update marks the release failed.
   - If there is a previous release, re-apply it and record rollback metadata.
   - If a prod DB migration already ran and later rollback is needed, run the configured rollback command or stop with the configured restore plan.

## Humanist Acceptance Criteria

- `../humanist.design/deploy.yml` uses `.env.prod` for prod deploy input.
- `../humanist.design/.env.prod` is ignored and absent from git.
- Humanist's prod deploy config and hooks do not reference `.env.local`.
- `pnpm dev` in Humanist continues to use `.env.local`.
- Convex function sync during deploy uses `.env.prod`.
- Prod S3 bucket creation uses `.env.prod`.
- Prod S3 buckets use systematic prod-specific names.
- Prod MinIO web UI is SSH-forwarded only, with no public Caddy route.
- A deploy failure in Convex env push or Convex sync stops before frontend traffic changes.
- Smoke check results are written to release metadata.
- The metadata makes it clear which env file was used without logging secret values.

## Resolved Decisions

- Humanist should run all three app-level smoke checks: HTTP action, Convex query, and a tiny Playwright landing-page load that catches obvious JavaScript crashes.
- `.env.prod` should persist `CONVEX_SELF_HOSTED_ADMIN_KEY`; deploy should add it when missing by refreshing it from the running prod backend.
- Rename the Convex env push script to `scripts/push-env-to-convex.sh`; default it to dev and add `--prod`.
- `pp` should not hard-code a global `.env.local` rejection policy.
- Prod MinIO web UI should stay SSH-only for now; add a helper script to forward it.
- Humanist S3 buckets should use a systematic environment suffix or prefix.
- Do not add `.env.prod.example` in this pass.
