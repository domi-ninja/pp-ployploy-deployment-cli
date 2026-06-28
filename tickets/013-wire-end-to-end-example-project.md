# 013 - Wire End-To-End Example Project

## Goal

Prove the deployment pipeline against one representative side project shape.

## Depends On

- 001 through 012

## Scope

- Add an example `deploy.yml`.
- Include web service, worker service, volume, migration, DB rollback, and smoke check.
- Exercise local build, bundle render, SSH transfer, compose apply, SSH-observed status, and rollback in a non-critical environment.
- Document the exact operator workflow.

## Deliverables

- Example project config.
- End-to-end runbook.
- Known limitations list.
- Follow-up ticket list for v2.

## Acceptance Criteria

- `deploy plan` gives a readable complete preview.
- `deploy` applies the example release without manual args.
- `deploy status` accurately reports host state.
- `deploy rollback` restores the previous verified release and DB state.
