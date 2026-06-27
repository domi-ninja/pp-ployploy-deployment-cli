# Deployment Config Schema Draft

## File

Default path: `deploy.yml`

Purpose: committed desired state for one deployable project. Secrets stay in `.env` or host-local secret files.

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
    agent_url: "https://127.0.0.1:7468"
    roles: [web]
  worker-01:
    ssh: deploy@worker-01.example.com
    agent_url: "https://127.0.0.1:7468"
    roles: [worker]

services:
  web:
    image: "quotes:${git_sha}"
    hosts: [app-01]
    command: ["node", "server.js"]
    env_file: ".env.prod"
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
    env_file: ".env.prod"
    health:
      command: ["node", "scripts/healthcheck-worker.js"]
      timeout_seconds: 60

volumes:
  app-data:
    driver: local

migrations:
  command: ["pnpm", "db:migrate:prod"]
  env_file: ".env.prod"
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
- `hosts`: map of host IDs to SSH and agent metadata.
- `services`: map of logical service IDs to image, host placement, runtime config, and health policy.

## Service Model

- One logical service may run on one or more hosts.
- CLI renders one compose file per host by selecting services whose `hosts` includes that host ID.
- `image` may use template variables: `${git_sha}`, `${project}`, `${environment}`, `${release}`.
- `env_file` references local files read during deploy and rendered into the host bundle.
- `ports`, `volumes`, `command`, `entrypoint`, `restart`, `labels`, and `depends_on` map closely to Docker Compose.
- Cross-host dependencies must use URLs or external service addresses; generated Docker networks are host-local only.

## Migration Model

- Prod DB migration is an explicit deploy step, not a side effect of app startup.
- Local development uses a separate dev DB.
- Migration command runs from the deploy machine with deploy-time env loaded, unless later changed to run inside a one-shot container.
- Failed migration aborts before service update.

## Host Agent Contract

- Agent is a persistent Go binary installed by Ansible.
- CLI talks to each host agent before and after apply.
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
- Required env files must exist locally before build/apply starts.
- Published ports must not conflict on the same host.
- At least one health check must exist per externally reachable service.
- `project.name`, host IDs, service IDs, and volume IDs must be DNS/slug-safe.
- Rendered compose files must be deterministic for the same config, env, and git SHA.

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

Bundle metadata records git SHA, image tags, config digest, host list, migration status, and smoke check results.
