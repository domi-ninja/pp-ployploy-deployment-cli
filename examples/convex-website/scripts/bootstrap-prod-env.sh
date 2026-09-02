#!/usr/bin/env bash

set -euo pipefail

ENV_FILE="${1:-.env.prod}"

touch "$ENV_FILE"
chmod 600 "$ENV_FILE"

random_hex() {
  openssl rand -hex 32
}

random_b64() {
  openssl rand -base64 36 | tr -d '\n'
}

get_env() {
  local key="$1"
  sed -n "s/^${key}=//p" "$ENV_FILE" | tail -n 1
}

set_env() {
  local key="$1"
  local value="$2"
  if grep -q "^${key}=" "$ENV_FILE"; then
    tmp="$(mktemp)"
    awk -v key="$key" -v value="$value" 'BEGIN { done=0 } $0 ~ "^" key "=" { if (!done) { print key "=" value; done=1 } next } { print } END { if (!done) print key "=" value }' "$ENV_FILE" > "$tmp"
    mv "$tmp" "$ENV_FILE"
  else
    printf '%s=%s\n' "$key" "$value" >> "$ENV_FILE"
  fi
}

set_if_missing() {
  local key="$1"
  local value="$2"
  if [[ -z "$(get_env "$key")" ]]; then
    set_env "$key" "$value"
  fi
}

set_if_missing VITE_SITE_URL "https://humanist.design"
set_if_missing VITE_CONVEX_URL "https://api.humanist.design"
set_if_missing VITE_CONVEX_SITE_URL "https://api.humanist.design"
set_if_missing SITE_URL "https://humanist.design"
set_if_missing CONVEX_SELF_HOSTED_URL "https://api.humanist.design"

set_if_missing INSTANCE_NAME "api.humanist.design"
set_if_missing INSTANCE_SECRET "$(random_hex)"

set_if_missing POSTGRES_DB "$(get_env INSTANCE_NAME)"
set_if_missing POSTGRES_USER "humanist"
set_if_missing POSTGRES_PASSWORD "$(random_b64)"
set_env POSTGRES_URL "postgres://$(get_env POSTGRES_USER):$(get_env POSTGRES_PASSWORD)@postgres:5432?sslmode=disable"
set_env DATABASE_URL "postgres://$(get_env POSTGRES_USER):$(get_env POSTGRES_PASSWORD)@postgres:5432/$(get_env POSTGRES_DB)?sslmode=disable"

set_if_missing MINIO_ROOT_USER "humanist-minio"
set_if_missing MINIO_ROOT_PASSWORD "$(random_b64)"
set_env AWS_ACCESS_KEY_ID "$(get_env MINIO_ROOT_USER)"
set_env AWS_SECRET_ACCESS_KEY "$(get_env MINIO_ROOT_PASSWORD)"

set_if_missing S3_STORAGE_EXPORTS_BUCKET "humanist-prod-convex-exports"
set_if_missing S3_STORAGE_SNAPSHOT_IMPORTS_BUCKET "humanist-prod-convex-snapshot-imports"
set_if_missing S3_STORAGE_MODULES_BUCKET "humanist-prod-convex-modules"
set_if_missing S3_STORAGE_FILES_BUCKET "humanist-prod-convex-files"
set_if_missing S3_STORAGE_SEARCH_BUCKET "humanist-prod-convex-search"
set_if_missing AWS_S3_DISABLE_SSE "true"

echo "Wrote production deploy values to $ENV_FILE"
echo "CONVEX_SELF_HOSTED_ADMIN_KEY is intentionally not generated here."
echo "It is generated from the running Convex backend by scripts/refresh-convex-admin-key.sh."
