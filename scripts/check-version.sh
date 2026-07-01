#!/usr/bin/env bash
# Verify the Version constant in backend/internal/version/version.go matches the
# expected release version, so a tagged release can never ship a stale version
# string (which makes the in-app "Update available" banner stick forever).
#
# Usage:
#   scripts/check-version.sh [expected-version]
#     - With an argument: compares against it (a leading "v" is stripped).
#     - Without:          compares against the exact git tag on HEAD, else the
#                         most recent tag reachable from HEAD.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
file="$root/backend/internal/version/version.go"

const="$(sed -n 's/.*Version[[:space:]]*=[[:space:]]*"\(.*\)".*/\1/p' "$file")"
if [ -z "$const" ]; then
  echo "check-version: could not read Version from $file" >&2
  exit 2
fi

expected="${1:-}"
if [ -z "$expected" ]; then
  expected="$(git describe --tags --exact-match 2>/dev/null \
           || git describe --tags --abbrev=0 2>/dev/null || true)"
fi
expected="${expected#v}"

if [ -z "$expected" ]; then
  echo "check-version: no expected version given and no git tag found; version.go=$const" >&2
  exit 0
fi

if [ "$const" != "$expected" ]; then
  echo "check-version: MISMATCH — version.go=$const but expected=$expected" >&2
  echo "Bump backend/internal/version/version.go to \"$expected\" before tagging/releasing." >&2
  exit 1
fi

echo "check-version: OK — version.go=$const matches $expected"
