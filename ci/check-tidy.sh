#!/usr/bin/env bash
set -euo pipefail

# Tidy every module independently so the workspace cannot mask missing dependencies.
find . -name 'go.mod' -execdir env GOWORK=off go mod tidy \;

# A second pass ensures the resulting module metadata is stable.
find . -name 'go.mod' -execdir env GOWORK=off go mod tidy \;

# Use status instead of diff so a newly created go.sum also fails the check.
changes="$(
  git status \
    --porcelain \
    --untracked-files=all \
    -- ':(glob)**/go.mod' ':(glob)**/go.sum'
)"
if [[ -z ${changes} ]]; then
  exit 0
fi

printf '%s\n' "${changes}"
git diff -- ':(glob)**/go.mod' ':(glob)**/go.sum'
exit 1
