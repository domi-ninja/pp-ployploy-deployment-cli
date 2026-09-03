# Deployment examples

These examples started from sibling repositories, then replaced project-specific names and domains with consistent placeholders. Replace `deploy@example.com` and the `example.com` domains before using them.

- `simple-website/` shows one built web image at `simple.example.com`.
- `convex-website/` shows the more complex setup needed to run Convex in self-hosted mode. It lets a side project use Convex's database and backend functions while hosting the services on your own server.
- `stateful-go-website/` shows a required secret, persistent volume, migrations, and a custom entrypoint at `stateful.example.com`.

Application source and dependency lockfiles are not duplicated here. No production environment file or secret is included.
