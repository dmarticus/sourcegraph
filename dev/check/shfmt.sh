#!/usr/bin/env bash

echo "--- shfmt (ensure shell-scripts are formatted consistently)"

set -e
cd "$(dirname "${BASH_SOURCE[0]}")"/../..

set +e

# Ignore bash scripts in git submodules
OUT=$(
  shfmt -d \
    $(shfmt -f . | grep -v docker-images/syntax-highlighter/crates/)
)
EXIT_CODE=$?
set -e
echo -e "$OUT"

if [ $EXIT_CODE -ne 0 ]; then
  echo -e "$OUT" | ./dev/ci/annotate.sh -s "shfmt"
  echo "^^^ +++"
fi

exit $EXIT_CODE
