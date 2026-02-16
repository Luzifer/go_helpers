#!/usr/bin/env bash
set -euo pipefail

function log() {
  echo "[$(date +%H:%M:%S)] $@" >&2
}

for module in $(find . -name 'go.mod' | cut -d / -f 2 | sort); do
  log "Working in ${module}..."

  tag="${module}/v$(semver-from-commits -f "${module}")"
  log "  => Calculated tag as ${tag}"

  if git tag | grep -q "${tag}"; then
    log "  => Tag already exists, no need to tag"
    continue
  fi

  log "  => Issuing tag..."
  git tag "${tag}" HEAD
done
