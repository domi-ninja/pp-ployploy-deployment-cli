# Deployment Pipeline Plan

## Goal

Param-less deploy CLI for side projects: local development uses a separate dev DB, release builds happen locally, production migrations run explicitly against prod DB, and production is updated across multiple bare-metal Docker hosts with predictable state tracking.

## Core Shape

- `deploy`: single repo-local CLI command; no required args in the normal path.
- Inputs: committed deployment config, local `.env`, current git commit, reachable prod hosts.
- Output: built image pushed or shipped, rendered compose specs, applied host changes, deployment record.
- State source: repo config is desired state; server management client reports observed state.
- Provisioning: Ansible owns base server setup, Docker, volumes, users, firewall, and the management client.

## Artifact Strategy

- Build containers locally for now: deterministic local `docker buildx build` from repo root.
- Tag images by project, environment, git SHA, and optional monotonic release number.
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
- Rendered env files are transferred to hosts with strict permissions.
- Validate required secrets before build/apply; fail before touching prod if anything is missing.

## Server Management Client

- Persistent Go host agent delivered as a binary records: current release, compose project status, container health, image tags, disk space, last deploy log.
- CLI queries all target hosts before deployment and after rollout.
- SSH remains the transport for install/update/apply, but status is read from the agent.
- Agent exposes a narrow local or authenticated HTTP API and persists state on the host.

## Deploy Flow

1. Preflight: clean enough git state policy, config schema validation, `.env` validation, host reachability, Docker availability.
2. Build: build images locally, tag by git SHA, run local smoke/test commands.
3. Publish: transfer image tar to assigned hosts over SSH and load it into Docker.
4. Render: generate per-host compose and env files into a deploy bundle.
5. Apply: upload bundle, run `docker compose pull/load && docker compose up -d`.
6. Verify: wait for health checks, query management state, run optional HTTP smoke checks.
7. Record: write deployment metadata locally and on each host.
8. Rollback: re-apply previous known-good release bundle/image tag.

## CLI Contract

- `deploy`: full default deployment.
- `deploy plan`: show target hosts, services, image tags, and generated compose diff.
- `deploy status`: read observed state from hosts.
- `deploy rollback`: restore last known-good release.
- Config decides everything; flags are for diagnostics and emergency overrides only.

## Initial Milestones

1. Define deployment config schema and one example project.
2. Implement `deploy plan` with per-host compose rendering.
3. Implement local build and image tagging.
4. Implement Go host agent status API and SSH-based apply.
5. Add health checks, deployment records, and rollback.
6. Move repeated server setup into Ansible roles and wire hosts from Terraform-style vars later.

## Decisions To Lock

- SSH image transfer-first; registry support is optional later.
- Single monorepo deploy CLI package.
- Run migration commands explicitly on prod DB; local development uses a separate dev DB.
- Persistent Go host agent delivered as a binary.
