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
- Use the existing local deployment metadata under `.deploy/releases/`.
- Read remote state from Docker, Caddy route files, image tags, disk usage, and pp allocation files.

## Existing Baseline

- `pp status` exists.
- Local `.deploy/state.json` records current and previous release IDs.
- Local `.deploy/releases/<release>/metadata.json` records release metadata, hosts, image bundles, migration state, apply state, rollback state, and smoke-check results.
- `pp plan` prints the deploy env source without secret values.

This ticket is not about adding another local-only status command. It is about making status trustworthy by comparing local desired state with observed host state over SSH.

## Deliverables

- SSH-observed `deploy status` implementation.
- Observed-state model layered on top of the existing local metadata schema.
- SSH observed-state reader.
- Tests for status formatting and mismatch detection.

## Acceptance Criteria

- Status works without starting a deployment.
- Missing/unreachable SSH hosts are clearly shown per host.
- Desired/observed mismatches are called out.
- Previous release data is available for rollback.
- Migration and DB rollback state are visible in status output.
- Smoke-check results are visible in status output.
- Status reports at least:
  - local current/previous release
  - remote containers by `pp.project`
  - container release labels
  - container health/status
  - image tags
  - Caddy route digest/presence
  - disk usage for release and persistent data paths
  - allocated backend ports from `/etc/pp/ports.tsv`
