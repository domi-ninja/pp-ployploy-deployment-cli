# 011 - Implement SSH-Observed Status And Deployment Records

## Goal

Implement `deploy status` with local deployment records plus SSH-observed host state.

## Depends On

- 009
- 010

## Scope

- Query configured hosts over SSH.
- Display desired vs observed state.
- Show current release, previous release, git SHA, dirty flag, service status, image tags, disk space, and last deploy result.
- Store local deployment metadata under `.deploy/releases/`.
- Read remote state from Docker, Caddy route files, image tags, disk usage, and pp allocation files.

## Deliverables

- `deploy status` implementation.
- Local metadata schema.
- SSH observed-state reader.
- Tests for status formatting and mismatch detection.

## Acceptance Criteria

- Status works without starting a deployment.
- Missing/unreachable SSH hosts are clearly shown per host.
- Desired/observed mismatches are called out.
- Previous release data is available for rollback.
- Migration and DB rollback state are visible in status output.
