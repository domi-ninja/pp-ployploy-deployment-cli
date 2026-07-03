# 012 - Implement Rollback

## Goal

Restore the previous known-good code and DB state with a single command.

## Depends On

- 011

## Scope

- Identify previous verified release from local and host metadata.
- Confirm required image tags and bundle files exist on hosts.
- Run configured DB rollback command or restore plan.
- Re-apply previous compose bundle.
- Handle partial deployments by rolling back only hosts that advanced.
- Run health checks after rollback.
- Update deployment records with rollback metadata.

## Existing Baseline

- `pp rollback` exists.
- Local state records current and previous release IDs.
- Rollback reuses previous bundle metadata and does not rebuild images.
- Failed deploy apply attempts to re-apply the previous release.
- If a configured migration ran and a later step fails, deploy attempts the migration rollback command.
- Basic rollback metadata is present in release metadata.

This ticket remains open because the current rollback path is operationally useful but not yet rigorous enough to be the trusted recovery mechanism.

## Deliverables

- Hardened `deploy rollback` implementation.
- Previous verified release selector.
- Remote artifact verifier.
- Partial-host rollback planner.
- DB rollback/restore-plan executor with clear failure behavior.
- Rollback metadata format that records attempted hosts, DB rollback/restore state, health checks, and errors.
- Tests for missing previous release, missing host artifacts, restore-plan-only migrations, partial-host rollback, and successful rollback planning.

## Acceptance Criteria

- Rollback never rebuilds images.
- Rollback fails early if previous artifacts are missing.
- Rollback fails early if DB rollback is required but no rollback command or restore plan exists.
- Successful rollback updates host state.
- Failed rollback preserves logs and current observed state.
- Rollback is invoked automatically after failed host apply or failed smoke check when migration already succeeded.
- Rollback uses SSH-observed state from ticket 011 when deciding what advanced and what needs rollback.
- Rollback health checks are recorded separately from deploy smoke checks.
