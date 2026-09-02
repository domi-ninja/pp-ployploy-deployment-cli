# Deployment examples

These are deployment-config snapshots from sibling repositories. They keep the original project names, hosts, and domains so each example stays concrete. Replace those values before using one for another project.

- `simple-website/` comes from `blog.domi.ninja` and shows one built web image behind a route.
- `convex-website/` comes from `humanist.design` and shows a self-hosted Convex stack with Postgres, object storage, route templates, hooks, and smoke checks.
- `stateful-go-website/` comes from `gotths-example` and shows a required secret, persistent volume, migrations, and a custom entrypoint.

Application source and dependency lockfiles are not duplicated here. No production environment file or secret is included.
