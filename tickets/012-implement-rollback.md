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

## Deliverables

- `deploy rollback` implementation.
- Previous-release selector.
- DB rollback executor.
- Rollback metadata format.
- Tests for missing previous release and successful rollback planning.

## Acceptance Criteria

- Rollback never rebuilds images.
- Rollback fails early if previous artifacts are missing.
- Rollback fails early if DB rollback is required but no rollback command or restore plan exists.
- Successful rollback updates host state.
- Failed rollback preserves logs and current observed state.
- Rollback is invoked automatically after failed host apply or failed smoke check when migration already succeeded.
