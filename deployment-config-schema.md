# Deployment Config Schema Draft

## File

Default path: `deploy.yml`

Purpose: committed desired state for one deployable project. Secrets stay in `.env` or a later secret provider, then are provisioned into runtime containers by the deploy system.

## Example

```yaml
version: 1

project:
  name: quotes
  environment: prod

build:
  context: .
  dockerfile: Dockerfile
  target: runtime
  platforms:
    - linux/amd64
  tags:
    - "quotes:${git_sha}"

hosts:
  app-01:
    ssh: deploy@app-01.example.com
    agent:
      local_url: "http://127.0.0.1:7468"
    roles: [web]
  worker-01:
    ssh: deploy@worker-01.example.com
    agent:
      local_url: "http://127.0.0.1:7468"
    roles: [worker]

services:
  web:
    image: "quotes:${git_sha}"
    hosts: [app-01]
    command: ["node", "server.js"]
    env:
      source: ".env.prod"
      required: ["DATABASE_URL", "SESSION_SECRET"]
    ports:
      - published: 443
        target: 3000
    volumes:
      - name: app-data
        target: /data
    health:
      http: "https://quotes.example.com/health"
      timeout_seconds: 60

  worker:
    image: "quotes:${git_sha}"
    hosts: [worker-01]
    command: ["node", "worker.js"]
    env:
      source: ".env.prod"
      required: ["DATABASE_URL", "QUEUE_URL"]
    health:
      command: ["node", "scripts/healthcheck-worker.js"]
      timeout_seconds: 60

volumes:
  app-data:
    driver: local

migrations:
  image: "quotes:${git_sha}"
  command: ["pnpm", "db:migrate:prod"]
  rollback_command: ["pnpm", "db:rollback:prod"]
  env:
    source: ".env.prod"
    required: ["DATABASE_URL"]
  run: before_services
  timeout_seconds: 300

checks:
  smoke:
    - name: public health
      url: "https://quotes.example.com/health"
      expect_status: 200
```

## Required Fields

- `version`: schema version; start at `1`.
- `project.name`: stable slug used for compose project names, image tags, bundle paths, and agent state.
- `project.environment`: deployment environment, usually `prod`.
- `build.context`: local Docker build context.
- `build.dockerfile`: Dockerfile path inside the context.
- `hosts`: map of host IDs to SSH and SSH-reached local agent metadata.
- `services`: map of logical service IDs to image, host placement, runtime config, and health policy.

## Service Model

- One logical service may run on one or more hosts.
- CLI renders one compose file per host by selecting services whose `hosts` includes that host ID.
- `image` may use template variables: `${git_sha}`, `${project}`, `${environment}`, `${release}`.
- `env.source` references local secret/env input read during deploy.
- `env.required` declares values that must be available before build/apply starts.
- Runtime environment is provisioned into containers by the deploy system; v1 may render strict-permission host env files, but services should only depend on container env vars.
- `ports`, `volumes`, `command`, `entrypoint`, `restart`, `labels`, and `depends_on` map closely to Docker Compose.
- Cross-host dependencies must use URLs or external service addresses; generated Docker networks are host-local only.

## Migration Model

- Prod DB migration is an explicit deploy step, not a side effect of app startup.
- Local development uses a separate dev DB.
- Migration command runs inside the built app image as a one-shot container with deploy-time env loaded.
- Failed migration aborts before service update.
- DB rollback is explicit. Each migration config must define `rollback_command` or an explicit restore plan before automatic rollback is allowed.

## Host Agent Contract

- Agent is a persistent Go binary installed by Ansible.
- CLI reaches each host agent through SSH before and after apply.
- Agent binds to a host-local interface; SSH is the control path.
- Minimum agent state:
  - host ID and agent version
  - current project release
  - compose project status
  - service/container health
  - loaded image tags
  - disk capacity and free space
  - last deploy ID, timestamp, status, and log path

## Validation Rules

- Host IDs referenced by services must exist in `hosts`.
- Volume names referenced by services must exist in `volumes`, unless marked external.
- Required env sources and required env values must exist locally before build/apply starts.
- Published ports must not conflict on the same host.
- At least one health check must exist per externally reachable service.
- `project.name`, host IDs, service IDs, and volume IDs must be DNS/slug-safe.
- Rendered compose files must be deterministic for the same config, env, and git SHA.
- Release IDs use timestamp plus git SHA, for example `20260628T142233Z-abc1234`.
- Dirty git worktrees are allowed, but metadata must record dirty state and a worktree diff digest.

## Generated Bundle

Per deployment, the CLI creates a local bundle:

```text
.deploy/releases/<release-id>/
  deploy.yml
  metadata.json
  images/
    quotes-linux-amd64.tar
  hosts/
    app-01/
      compose.yml
      env/
        web.env
    worker-01/
      compose.yml
      env/
        worker.env
```

Bundle metadata records release ID, git SHA, dirty flag, worktree diff digest, image tags, config digest, host list, migration status, rollback status, and smoke check results.

## Failure Policy

- If migration fails, service update is not attempted.
- If migration succeeds and a later host apply or smoke check fails, rollback must cover both code and DB.
- Partial host applies are treated as failed deployments; successfully updated hosts are rolled back to the previous release.
