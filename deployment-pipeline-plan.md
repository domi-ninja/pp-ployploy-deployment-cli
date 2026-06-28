# Deployment Pipeline Plan

## Goal

Param-less Go deploy CLI for side projects: local development uses a separate dev DB, release builds happen locally, production migrations run explicitly against prod DB inside the built app container, and production is updated across multiple bare-metal Docker hosts with predictable state tracking.

## Core Shape

- `deploy`: single repo-local Go CLI command; no required args in the normal path.
- Inputs: committed deployment config, local `.env`, current git commit, reachable prod hosts.
- Output: built image tar, rendered compose specs, applied host changes, deployment record.
- State source: repo config is desired state; server management client reports observed state.
- Provisioning: Ansible owns base server setup, Docker, volumes, users, firewall, and the management client.
- Project fence: deployment system is Go; project code stays behind a language/tooling boundary.

## Artifact Strategy

- Build containers locally for now: deterministic local `docker buildx build` from repo root.
- Tag images by project, environment, git SHA, and timestamp-based release ID.
- Use SSH image transfer first: `docker save | ssh docker load` to each assigned host.
- Keep private registry support as a later optimization, not the default path.
- Never build on prod hosts; hosts only pull/load and run known image tags.

## Multi-Host Compose Strategy

- Treat each host as an independently applied compose project.
- Model topology in committed config: services, host placement, ports, volumes, networks, health checks.
- Generate one compose file per host from the same logical deployment config.
- Cross-host dependencies use explicit URLs, DNS, or reverse-proxy config; do not rely on Docker networks across hosts.
- Shared external services, especially prod DB, are config references, not managed deployment units unless explicitly owned.

## Config And Secrets

- Committed file: non-secret topology, image/build metadata, service definitions, host placement, health policy.
- `.env`: local secret values and environment-specific overrides used at deploy time.
- Runtime env vars are provisioned into containers by the deploy system; v1 can use strict-permission host env files, but the schema should model required container env rather than project code reading deployment internals.
- Validate required secrets before build/apply; fail before touching prod if anything is missing.

## Server Management Client

- Persistent Go host agent delivered as a binary records: current release, compose project status, container health, image tags, disk space, last deploy log.
- CLI queries all target hosts before deployment and after rollout.
- SSH is the only control path: agent access goes through SSH, image/bundle transfer goes through SSH, and compose apply runs over SSH.
- Agent binds locally on the host and persists state on the host.

## Deploy Flow

1. Preflight: config schema validation, `.env` validation, host reachability over SSH, Docker availability, dirty git state recorded.
2. Build: build images locally, tag by git SHA and timestamp release ID, run local smoke/test commands.
3. Publish: transfer image tar to assigned hosts over SSH and load it into Docker.
4. Render: generate per-host compose specs and runtime env artifacts into a deploy bundle.
5. Migrate: run prod DB migration inside the built app container before service update.
6. Apply: upload bundle, run `docker compose up -d` over SSH.
7. Verify: wait for health checks, query management state through SSH-reached agent, run optional HTTP smoke checks.
8. Record: write deployment metadata locally and on each host, including dirty worktree flag/digest.
9. Rollback: re-apply previous known-good release bundle/image tag and run DB rollback command/restore plan.

## CLI Contract

- `deploy`: full default deployment.
- `deploy plan`: show target hosts, services, image tags, and generated compose diff.
- `deploy status`: read observed state from hosts.
- `deploy rollback`: restore last known-good code and DB state.
- Config decides everything; flags are for diagnostics and emergency overrides only.
- Dirty worktrees are allowed; releases record git SHA, timestamp, dirty flag, and worktree diff digest.

## Initial Milestones

1. Define deployment config schema and one example project.
2. Implement `deploy plan` with per-host compose rendering.
3. Implement local build and image tagging.
4. Implement Go host agent status API and SSH-based apply.
5. Add health checks, deployment records, and rollback.
6. Move repeated server setup into Ansible roles and wire hosts from Terraform-style vars later.

## Failure Policy

- If migration fails, no service update runs.
- If migration succeeds but host apply or smoke checks fail, rollback runs for both code and DB.
- If only some hosts applied successfully, rollback re-applies the previous release to those hosts and runs the DB rollback/restore path.
- Rollback itself records full metadata and logs; failed rollback leaves observed state visible through `deploy status`.

## Decisions To Lock

- SSH image transfer-first; registry support is optional later.
- Single monorepo deploy CLI package, written in Go.
- Run migration commands explicitly on prod DB inside the built app container; local development uses a separate dev DB.
- Release IDs combine timestamp and git SHA.
- Dirty deploys are allowed and recorded in metadata.
- Runtime env variables are provisioned into containers by the deploy system.
- SSH is the control plane for agent access, image transfer, and compose apply.
- Persistent Go host agent delivered as a binary.
- Rollback covers both code and DB state.
