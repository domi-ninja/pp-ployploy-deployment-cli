# Deploy CLI

Go implementation of the side-project deployment system.

## Current Slice

- `go run ./cmd/deploy --help`
- `go run ./cmd/deploy plan`

`deploy plan` currently loads `deploy.yml`, validates schema references and required env values, reads git metadata, computes a timestamp-plus-SHA release ID, and prints host/service placement. Docker build, SSH transfer, host agent access, migrations, apply, status, and rollback are still ticketed work.
