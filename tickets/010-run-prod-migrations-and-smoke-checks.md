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
