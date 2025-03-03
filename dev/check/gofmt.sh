#!/usr/bin/env bash

set -e

echo "--- gofmt"

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

# Check if all code is gofmt'd

DIFF=$(find . \( -path ./vendor -o -path ./vendored \) -prune -o -name '*.go' -exec gofmt -s -w -d {} +)
if [ -z "$DIFF" ]; then
  echo "Success: gofmt check passed."
  exit 0
else
  echo "ERROR: gofmt check failed:"
  echo -e "$DIFF" | ./enterprise/dev/ci/scripts/annotate.sh -s "gofmt"
  echo "^^^ +++"
  exit 1
fi
