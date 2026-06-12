#!/usr/bin/env bash

set -euo pipefail

mode="${1:-}"

if [[ -z "$mode" ]]; then
  echo "usage: $0 <smoke|quarantine|regression>" >&2
  exit 1
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
manifest_dir="$repo_root/frontend/tests/e2e-manifests"
smoke_manifest="$manifest_dir/smoke.txt"
quarantine_manifest="$manifest_dir/quarantine.txt"
e2e_dir="$repo_root/frontend/tests/e2e"

read_manifest() {
  local manifest="$1"
  if [[ ! -f "$manifest" ]]; then
    return 0
  fi
  while IFS= read -r spec; do
    spec="$(printf '%s' "$spec" | sed 's/\r$//')"
    [[ "$spec" =~ ^[[:space:]]*(#|$) ]] && continue
    if [[ -f "$repo_root/frontend/$spec" || -f "$repo_root/$spec" ]]; then
      printf '%s\n' "$spec"
    fi
  done < "$manifest"
}

case "$mode" in
  smoke)
    read_manifest "$smoke_manifest"
    ;;
  quarantine)
    read_manifest "$quarantine_manifest"
    ;;
  regression)
    tmp_exclude="$(mktemp)"
    trap 'rm -f "$tmp_exclude"' EXIT
    {
      read_manifest "$smoke_manifest"
      read_manifest "$quarantine_manifest"
    } | sort -u > "$tmp_exclude"

    while IFS= read -r file; do
      rel="tests/e2e/${file##"$e2e_dir"/}"
      if ! grep -Fqx -- "$rel" "$tmp_exclude"; then
        printf '%s\n' "$rel"
      fi
    done < <(find "$e2e_dir" -maxdepth 1 -name '*.spec.ts' -print | sort)
    ;;
  *)
    echo "unknown mode: $mode" >&2
    exit 1
    ;;
esac
