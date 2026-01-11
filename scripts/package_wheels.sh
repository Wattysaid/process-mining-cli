#!/usr/bin/env bash
set -euo pipefail

REQ_FILE="${1:-.codex/skills/cli-tool-skills/pm-99-utils-and-standards/requirements.txt}"
OUT_DIR="${2:-resources/wheels}"

if [[ ! -f "$REQ_FILE" ]]; then
  echo "[ERROR] Requirements file not found: $REQ_FILE"
  exit 1
fi

mkdir -p "$OUT_DIR"

python3 -m pip download \
  --only-binary=:all: \
  -r "$REQ_FILE" \
  -d "$OUT_DIR"

echo "[SUCCESS] Downloaded wheels to $OUT_DIR"
