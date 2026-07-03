# 013 - Humanist End-To-End Fixture And Runbook

## Goal

Keep `../humanist.design` as the real end-to-end fixture for the lightweight `pp` deployment pipeline.

The original generic example-project ticket is superseded. Humanist now exercises the actual deployment shape that matters: local image build, upstream images, persistent Postgres, S3-compatible storage, self-hosted Convex, Caddy route templating, deploy hooks, phased apply, no-arg deploy, container cleanup, and smoke checks.

## Depends On

- 010
- `tickets/humanist-convex/004`

`011` and `012` remain separate open tickets. This fixture should document their gaps, not pretend SSH-observed status or hardened rollback are complete.

## Scope

- Treat `../humanist.design/deploy.yml` as the canonical end-to-end fixture.
- Keep the Humanist deploy runnable with no args from the app repo.
- Keep `pp down` usable as the non-data-destructive cleanup path for the fixture.
- Document the exact operator workflow used for p3.
- Record known limitations for status, rollback, backups, and destructive data reset.
- Keep follow-up tickets linked instead of expanding this ticket into v2 operations work.

## Current Proven Fixture

- `pp plan` validates Humanist and shows `.env.prod` as the deploy env source.
- No-arg `pp` deploy applies Humanist to `p3.domi.ninja`.
- `pp down` removes Humanist project containers by `pp.project=humanist-design` without deleting persistent data.
- Deploy starts `infra` before `backend`, avoiding the Convex/Postgres cold-start race.
- Humanist deploy passes all configured smoke checks:
  - frontend
  - `www` redirect
  - Convex `/version`
  - dashboard
  - Convex HTTP action
  - Convex query
  - Playwright landing-page load

## Deliverables

- End-to-end runbook for Humanist.
- Fixture notes explaining `.env.local` vs `.env.prod`.
- Cleanup notes for `pp down`, including that it preserves Postgres and MinIO data.
- Known limitations list.
- Links to follow-up tickets:
  - `011` for SSH-observed status
  - `012` for hardened rollback
  - `tickets/humanist-convex/005` for Postgres backups
  - a future destructive data reset ticket if empty-db fixture resets are required

## Acceptance Criteria

- Humanist runbook documents:
  - bootstrap or restore `.env.prod`
  - `pp plan`
  - no-arg `pp`
  - `pp status`
  - `pp down`
  - MinIO SSH forwarding
  - smoke-check expectations
- A fresh operator can identify which commands are safe and which preserve production data.
- The fixture does not require public MinIO UI exposure.
- The fixture explicitly states that `pp down` does not empty the database.
- The fixture links status correctness to ticket `011`.
- The fixture links rollback correctness to ticket `012`.
