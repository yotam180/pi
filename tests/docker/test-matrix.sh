#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

ENVIRONMENTS=(
  ubuntu-fresh
  ubuntu-node
  ubuntu-python
  alpine-fresh
)

PASS=()
FAIL=()
SKIP=()

for env in "${ENVIRONMENTS[@]}"; do
  dockerfile="$SCRIPT_DIR/$env/Dockerfile"
  if [[ ! -f "$dockerfile" ]]; then
    echo "⚠  $env — Dockerfile not found, skipping"
    SKIP+=("$env")
    continue
  fi

  image="pi-test-$env"
  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "  Building: $env"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  if ! docker build -t "$image" -f "$dockerfile" "$REPO_ROOT" 2>&1; then
    echo "✗  $env — build failed"
    FAIL+=("$env")
    continue
  fi

  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "  Testing: $env"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  if docker run --rm "$image" 2>&1; then
    echo "✓  $env — passed"
    PASS+=("$env")
  else
    echo "✗  $env — failed"
    FAIL+=("$env")
  fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [[ ${#PASS[@]} -gt 0 ]]; then
  for env in "${PASS[@]}"; do
    echo "  ✓  $env"
  done
fi

if [[ ${#FAIL[@]} -gt 0 ]]; then
  for env in "${FAIL[@]}"; do
    echo "  ✗  $env"
  done
fi

if [[ ${#SKIP[@]} -gt 0 ]]; then
  for env in "${SKIP[@]}"; do
    echo "  ⚠  $env (skipped)"
  done
fi

echo ""
echo "  ${#PASS[@]} passed, ${#FAIL[@]} failed, ${#SKIP[@]} skipped"
echo ""

if [[ ${#FAIL[@]} -gt 0 ]]; then
  exit 1
fi
