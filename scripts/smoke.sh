#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

cat > "$TMP_DIR/log.csv" <<'CSV'
case_id,activity,timestamp
1,A,2024-01-01 10:00:00
1,B,2024-01-01 11:00:00
CSV

cd "$ROOT_DIR"
go build -o "$TMP_DIR/pm-assist" ./cmd/pm-assist

"$TMP_DIR/pm-assist" review \
  --project "$TMP_DIR" \
  --run-id smoke \
  --input "$TMP_DIR/log.csv" \
  --case case_id \
  --activity activity \
  --timestamp timestamp \
  --missing-threshold 0.1 \
  --duplicate-threshold 0.1 \
  --order-threshold 0.1 \
  --parse-threshold 0.1 \
  --allow-blocking true \
  --non-interactive

echo "[SUCCESS] Smoke run completed: $TMP_DIR/outputs/smoke"
