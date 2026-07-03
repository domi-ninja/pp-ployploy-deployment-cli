# Deployment Pipeline Tickets

Numbered implementation tickets for the side-project deployment pipeline.

## Open

1. [011 - Implement SSH-Observed Status And Deployment Records](011-implement-status-and-deployment-records.md)
2. [012 - Implement Rollback](012-implement-rollback.md)
3. [013 - Wire End-To-End Example Project](013-wire-end-to-end-example-project.md)

## Done

1. [001 - Lock Deploy Config Schema](done/001-lock-deploy-config-schema.md)
2. [002 - Scaffold Monorepo Deploy CLI](done/002-scaffold-monorepo-deploy-cli.md)
3. [003 - Parse, Validate, And Render Config](done/003-parse-validate-render-config.md)
4. [004 - Render Per-Host Compose Bundles](done/004-render-per-host-compose-bundles.md)
5. [005 - Build And Tag Local Images](done/005-build-and-tag-local-images.md)
6. [006 - Transfer Images And Bundles Over SSH](done/006-transfer-images-and-bundles-over-ssh.md)
7. [009 - Apply Compose Releases On Hosts](done/009-apply-compose-releases-on-hosts.md)
8. [014 - Allocate Project Backend Ports](done/014-allocate-project-backend-ports.md)
9. [010 - Run Prod Migrations And Smoke Checks](done/010-run-prod-migrations-and-smoke-checks.md)

## Dropped

1. [007 - Build Go Host Agent MVP](dropped/007-build-go-host-agent-mvp.md)
2. [008 - Install Host Agent With Ansible](dropped/008-install-host-agent-with-ansible.md)

## Ticket Format

Each ticket defines scope, dependencies, deliverables, and acceptance criteria. Keep tickets small enough that one ticket can become one branch or PR.
