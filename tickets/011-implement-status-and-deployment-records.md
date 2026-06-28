# 011 - Implement Status And Deployment Records

## Goal

Implement `deploy status` and make deployment records useful enough for operations.

## Depends On

- 007
- 009
- 010

## Scope

- Query all configured host agents.
- Display desired vs observed state.
- Show current release, previous release, git SHA, dirty flag, service status, image tags, disk space, and last deploy result.
- Store local deployment metadata under `.deploy/releases/`.
- Store host-local deployment metadata for agent reporting.

## Deliverables

- `deploy status` implementation.
- Local metadata schema.
- Host metadata schema.
- Tests for status formatting and mismatch detection.

## Acceptance Criteria

- Status works without starting a deployment.
- Missing/unreachable agents are clearly shown per host.
- Desired/observed mismatches are called out.
- Previous release data is available for rollback.
- Migration and DB rollback state are visible in status output.
