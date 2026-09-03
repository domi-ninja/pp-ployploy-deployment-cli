#!/usr/bin/env bash

set -euo pipefail

ENV_FILE="${ENV_FILE:-.env.local}"
SSH_TARGET="${PP_HOST:-deploy@example.com}"
PROJECT="${PP_PROJECT:-convex-website}"
SERVICE="${PP_SERVICE:-convex-backend}"
FORCE=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --prod)
      ENV_FILE=".env.prod"
      shift
      ;;
    --env-file)
      if [[ $# -lt 2 ]]; then
        echo "Error: --env-file requires a value" >&2
        exit 1
      fi
      ENV_FILE="$2"
      shift 2
      ;;
    --force)
      FORCE=true
      shift
      ;;
    *)
      echo "Error: unknown argument $1" >&2
      exit 1
      ;;
  esac
done

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Error: $ENV_FILE file not found. Run scripts/bootstrap-prod-env.sh first." >&2
  exit 1
fi

existing_key="$(sed -n 's/^CONVEX_SELF_HOSTED_ADMIN_KEY=//p' "$ENV_FILE" | tail -n 1)"
if [[ -n "$existing_key" && "$FORCE" != true ]]; then
  echo "CONVEX_SELF_HOSTED_ADMIN_KEY already exists in $ENV_FILE"
  exit 0
fi

container="$(
  ssh "$SSH_TARGET" \
    "docker ps --filter label=pp.project=$PROJECT --filter label=pp.service=$SERVICE --format '{{.Names}}' | head -n 1"
)"

if [[ -z "$container" ]]; then
  echo "Error: no running $SERVICE container found for pp.project=$PROJECT on $SSH_TARGET" >&2
  exit 1
fi

admin_key="$(
  ssh "$SSH_TARGET" "docker exec $(printf '%q' "$container") /convex/generate_admin_key.sh | tail -n 1"
)"

if [[ -z "$admin_key" || "$admin_key" != *"|"* ]]; then
  echo "Error: Convex container returned an invalid admin key" >&2
  exit 1
fi

tmp="$(mktemp)"
awk -v key="$admin_key" '
  BEGIN { done=0 }
  /^CONVEX_SELF_HOSTED_ADMIN_KEY=/ {
    if (!done) {
      print "CONVEX_SELF_HOSTED_ADMIN_KEY=" key
      done=1
    }
    next
  }
  { print }
  END {
    if (!done) print "CONVEX_SELF_HOSTED_ADMIN_KEY=" key
  }
' "$ENV_FILE" > "$tmp"
mv "$tmp" "$ENV_FILE"
chmod 600 "$ENV_FILE"

echo "Updated CONVEX_SELF_HOSTED_ADMIN_KEY in $ENV_FILE from $container"
