# 001 - Extend pp For Multi-Service Convex Deploys

## Scope

Add the smallest generic `pp` features needed by self-hosted Convex projects:

- multiple local builds
- external image pulls
- build args from env
- route file templates
- service phases
- local hooks
- bind mounts

## Acceptance Criteria

- Existing single-build `deploy.yml` configs still validate, build, render, and deploy.
- A Convex-style config can render web, backend, dashboard, Postgres, and S3 services.
- Locally built images are saved and loaded as tar files.
- External images are pulled on the host before compose apply.
- Route templates can reference allocated service ports.
- Hooks can run locally after backend services are healthy.

