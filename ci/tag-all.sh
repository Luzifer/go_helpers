#!/usr/bin/env bash
set -euo pipefail

function log() {
  echo "[$(date +%H:%M:%S)] $@" >&2
}

declare -a modules=()
declare -a tags=()

command -v semver-from-commits >/dev/null

while IFS= read -r -d '' modfile; do
  module="$(dirname "${modfile}")"
  module="${module#./}"

  if [[ ${module} == "." ]]; then
    continue
  fi

  modules+=("${module}")
done < <(find . -name 'go.mod' -print0 | sort -z)

for module in "${modules[@]}"; do
  log "Working in ${module}..."

  version="$(semver-from-commits -f "${module}")"
  if [[ ! ${version} =~ ^[0-9]+\.[0-9]+\.[0-9]+([+-][0-9A-Za-z.-]+)?$ ]]; then
    log "  => Refusing invalid semantic version: ${version}"
    exit 1
  fi

  tag="${module}/v${version}"
  git check-ref-format "refs/tags/${tag}"
  log "  => Calculated tag as ${tag}"

  if git show-ref --verify --quiet "refs/tags/${tag}"; then
    log "  => Tag already exists, no need to tag"
    continue
  fi

  tags+=("${tag}")
done

for tag in "${tags[@]}"; do
  log "  => Issuing tag..."
  git tag "${tag}" HEAD
done
