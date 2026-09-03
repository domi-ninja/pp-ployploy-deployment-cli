#!/usr/bin/env bash

set -euo pipefail

ENV_FILE="${ENV_FILE:-.env.local}"
SSH_TARGET="${PP_HOST:-deploy@example.com}"
PROJECT="${PP_PROJECT:-convex-website}"

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
    *)
      echo "Error: unknown argument $1" >&2
      exit 1
      ;;
  esac
done

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Error: $ENV_FILE file not found" >&2
  exit 1
fi

required=(
  MINIO_ROOT_USER
  MINIO_ROOT_PASSWORD
  S3_STORAGE_EXPORTS_BUCKET
  S3_STORAGE_SNAPSHOT_IMPORTS_BUCKET
  S3_STORAGE_MODULES_BUCKET
  S3_STORAGE_FILES_BUCKET
  S3_STORAGE_SEARCH_BUCKET
)

get_env() {
  local key="$1"
  local line value
  line="$(grep -E "^${key}=" "$ENV_FILE" | tail -n 1 || true)"
  value="${line#*=}"
  if [[ "$value" =~ ^\"(.*)\"$ ]]; then
    value="${BASH_REMATCH[1]}"
  elif [[ "$value" =~ ^\'(.*)\'$ ]]; then
    value="${BASH_REMATCH[1]}"
  fi
  printf '%s' "$value"
}

declare -A env_values=()
for name in "${required[@]}"; do
  env_values["$name"]="$(get_env "$name")"
  if [[ -z "${env_values[$name]}" ]]; then
    echo "Error: $name is missing from $ENV_FILE" >&2
    exit 1
  fi
done

remote_script="$(cat <<'REMOTE'
set -euo pipefail

project="$1"
access_key="$2"
secret_key="$3"
shift 3

s3_container="$(docker ps --filter "label=pp.project=$project" --filter "label=pp.service=s3" --format '{{.Names}}' | head -n 1)"
if [ -z "$s3_container" ]; then
  echo "Error: no running s3 container found for pp.project=$project" >&2
  exit 1
fi

network="$(docker inspect "$s3_container" --format '{{range $name, $_ := .NetworkSettings.Networks}}{{println $name}}{{end}}' | head -n 1)"
if [ -z "$network" ]; then
  echo "Error: could not find Docker network for $s3_container" >&2
  exit 1
fi

docker run --rm --network "$network" --entrypoint /bin/sh minio/mc:latest -eu -c '
  alias_value="$1"
  access_key="$2"
  secret_key="$3"
  shift 3
  mc alias set local "$alias_value" "$access_key" "$secret_key" >/dev/null
  for bucket in "$@"; do
    mc mb --ignore-existing "local/$bucket" >/dev/null
  done
' sh http://s3:9000 "$access_key" "$secret_key" "$@"
REMOTE
)"

ssh "$SSH_TARGET" \
  "bash -s -- $(printf '%q' "$PROJECT") $(printf '%q' "${env_values[MINIO_ROOT_USER]}") $(printf '%q' "${env_values[MINIO_ROOT_PASSWORD]}") $(printf '%q' "${env_values[S3_STORAGE_EXPORTS_BUCKET]}") $(printf '%q' "${env_values[S3_STORAGE_SNAPSHOT_IMPORTS_BUCKET]}") $(printf '%q' "${env_values[S3_STORAGE_MODULES_BUCKET]}") $(printf '%q' "${env_values[S3_STORAGE_FILES_BUCKET]}") $(printf '%q' "${env_values[S3_STORAGE_SEARCH_BUCKET]}")" \
  <<< "$remote_script"

echo "Ensured MinIO buckets exist on $SSH_TARGET"
