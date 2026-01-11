#!/usr/bin/env bash
set -euo pipefail

SRC_DIR="${1:-.codex/skills/cli-tool-skills}"
OUT_DIR="${2:-resources/cli-tool-skills}"

if [[ ! -d "$SRC_DIR" ]]; then
  echo "[ERROR] Source skills directory not found: $SRC_DIR"
  exit 1
fi

mkdir -p "$OUT_DIR"

rsync -a --delete \
  --exclude '*.pyc' \
  --exclude '__pycache__/' \
  --exclude '.git/' \
  --exclude '*.ipynb' \
  "$SRC_DIR/" "$OUT_DIR/"

echo "[SUCCESS] Packaged skills to $OUT_DIR"
