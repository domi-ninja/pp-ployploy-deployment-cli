#!/usr/bin/env bash

set -euo pipefail

ENV_FILE=".env.local"
DEPLOYMENT_ARGS=()
PRUNE=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --prod)
      ENV_FILE=".env.prod"
      shift
      ;;
    prod)
      DEPLOYMENT_ARGS=(--prod)
      shift
      ;;
    --deployment-name)
      if [[ $# -lt 2 ]]; then
        echo "Error: --deployment-name requires a value" >&2
        exit 1
      fi
      DEPLOYMENT_ARGS=(--deployment "$2")
      shift 2
      ;;
    --env-file)
      if [[ $# -lt 2 ]]; then
        echo "Error: --env-file requires a value" >&2
        exit 1
      fi
      ENV_FILE="$2"
      shift 2
      ;;
    --prune)
      PRUNE=true
      shift
      ;;
    *)
      DEPLOYMENT_ARGS=(--deployment "$1")
      echo "Using deployment: $1"
      shift
      ;;
  esac
done

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Error: $ENV_FILE file not found" >&2
  exit 1
fi

if [[ "$ENV_FILE" == ".env.prod" ]]; then
  if ! grep -q '^CONVEX_SELF_HOSTED_URL=' "$ENV_FILE"; then
    echo "Error: CONVEX_SELF_HOSTED_URL is missing from $ENV_FILE" >&2
    exit 1
  fi
  if ! grep -q '^CONVEX_SELF_HOSTED_ADMIN_KEY=' "$ENV_FILE"; then
    echo "Error: CONVEX_SELF_HOSTED_ADMIN_KEY is missing from $ENV_FILE; run scripts/refresh-convex-admin-key.sh --prod first" >&2
    exit 1
  fi
fi

while IFS= read -r line || [[ -n "$line" ]]; do
  [[ "$line" =~ ^([A-Za-z_][A-Za-z0-9_]*)=(.*)$ ]] || continue
  var_name="${BASH_REMATCH[1]}"
  var_value="${BASH_REMATCH[2]}"
  if [[ "$var_value" =~ ^\"(.*)\"$ || "$var_value" =~ ^\'(.*)\'$ ]]; then
    var_value="${BASH_REMATCH[1]}"
  fi
  case "$var_name" in
    CONVEX_DEPLOYMENT|CONVEX_SELF_HOSTED_URL|CONVEX_SELF_HOSTED_ADMIN_KEY)
      export "$var_name=$var_value"
      ;;
  esac
done <"$ENV_FILE"

convex_bin="$(pwd)/node_modules/.bin/convex"
convex_cli_dir=""

if [[ -n "${CONVEX_SELF_HOSTED_URL:-}" ]]; then
  convex_cli_dir="$(mktemp -d)"
  cp package.json "$convex_cli_dir/package.json"
fi

run_convex() {
  if [[ -n "$convex_cli_dir" ]]; then
    (cd "$convex_cli_dir" && "$convex_bin" "$@")
  else
    pnpm exec convex "$@"
  fi
}

current_env_file="$(mktemp)"
cleanup() {
  rm -f "$current_env_file"
  if [[ -n "$convex_cli_dir" ]]; then
    rm -rf "$convex_cli_dir"
  fi
}
trap cleanup EXIT
run_convex env list "${DEPLOYMENT_ARGS[@]}" >"$current_env_file"

declare -A current_values=()
declare -A source_keys=()

while IFS= read -r line || [[ -n "$line" ]]; do
  [[ "$line" =~ ^([A-Za-z_][A-Za-z0-9_]*)=(.*)$ ]] || continue
  var_name="${BASH_REMATCH[1]}"
  var_value="${BASH_REMATCH[2]}"
  if [[ "$var_value" =~ ^\"(.*)\"$ || "$var_value" =~ ^\'(.*)\'$ ]]; then
    var_value="${BASH_REMATCH[1]}"
  fi
  current_values["$var_name"]="$var_value"
done <"$current_env_file"

updated=0

while IFS= read -r line || [[ -n "$line" ]]; do
  line="${line#"${line%%[![:space:]]*}"}"
  line="${line%"${line##*[![:space:]]}"}"
  [[ -z "$line" || "$line" == \#* ]] && continue

  if [[ "$line" =~ ^([A-Za-z_][A-Za-z0-9_]*)[[:space:]]*=[[:space:]]*(.*)$ ]]; then
    var_name="${BASH_REMATCH[1]}"
    var_value="${BASH_REMATCH[2]}"
    source_keys["$var_name"]=1

    if [[ "$var_value" =~ ^\"(.*)\"$ ]]; then
      var_value="${BASH_REMATCH[1]}"
    elif [[ "$var_value" =~ ^\'(.*)\'$ ]]; then
      var_value="${BASH_REMATCH[1]}"
    fi

    if [[ ${current_values[$var_name]+_} && "${current_values[$var_name]}" == "$var_value" ]]; then
      continue
    fi

    echo "Updating $var_name..."
    printf '%s' "$var_value" | run_convex env set "${DEPLOYMENT_ARGS[@]}" "$var_name"
    updated=$((updated + 1))
  else
    echo "Skipping unsupported env line: $line" >&2
  fi
done <"$ENV_FILE"

if [[ "$PRUNE" == true ]]; then
  for var_name in "${!current_values[@]}"; do
    if [[ ! ${source_keys[$var_name]+_} ]]; then
      echo "Removing $var_name..."
      run_convex env remove "${DEPLOYMENT_ARGS[@]}" "$var_name"
      updated=$((updated + 1))
    fi
  done
fi

if [[ "$updated" -eq 0 ]]; then
  echo "Convex environment is already up to date."
else
  echo "Updated $updated Convex environment variable(s)."
fi
