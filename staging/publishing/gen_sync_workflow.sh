#!/usr/bin/env bash
# Regenerate .github/workflows/ci_sync_repo.yaml from
# staging/publishing/rules.yaml.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
OUT="${ROOT}/.github/workflows/ci_sync_repo.yaml"
RULES="${ROOT}/staging/publishing/rules.yaml"

{
    cat <<'HEADER'
name: CI / sync sentinez

permissions:
  contents: write

on:
  push:
    branches:
HEADER
    # top-level "branches:" list items (before "rules:")
    sed -n '/^branches:/,/^rules:/{/^  - /s/^  - /      - /p}' "$RULES"
    echo
    echo "jobs:"
    first=1
    for n in $(sed -n 's/^  - destination: *//p' "$RULES"); do
        [ "$first" = 1 ] || echo
        first=0
        cat <<JOB
  ${n}:
    name: sync sentinez/${n}
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: Sync sentinez/${n}
        env:
          SENZ_GITHUB_TOKEN: \${{ secrets.SENZ_GITHUB_TOKEN }}
        run: bash hack/ci/action_sync.sh ${n}
JOB
    done
} >"$OUT"
