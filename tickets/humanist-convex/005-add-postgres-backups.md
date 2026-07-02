# 005 - Add Postgres Backups

## Scope

Add a backup and restore process for Humanist's p3 Postgres database.

This is deferred from the first deployment pass.

## Acceptance Criteria

- A manual backup can be triggered without stopping production.
- Restore commands are documented and tested against a non-production database.
- Backup output is stored outside release directories.
- Retention is documented.
