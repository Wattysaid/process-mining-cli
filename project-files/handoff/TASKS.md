# Outstanding Tasks

**Document version:** R1.01 (2026-01-12)

## 1) Packaging and Offline Assets (Critical)
- Bundle Python assets into `resources/cli-tool-skills` during release builds.
- Update release workflow to verify packaged assets.
- Decide wheel strategy: ship offline wheels or embed assets.

## 2) Logging and Run Manifests (Critical)
~~Implement structured logging with redaction.~~
~~Emit per-run manifest (config snapshot, inputs, outputs, hashes, step status).~~

## 3) Non-Interactive Support (Critical)
~~Add flags for every prompt in init/connect/ingest/prepare/mine/report/review.~~
~~Ensure `--non-interactive` fails on missing inputs.~~

## 4) Connector Expansion (High)
~~Snowflake and BigQuery: capture config + validation stub.~~
~~Doctor should validate connector reachability.~~

## 5) Reporting Enhancements (High)
~~HTML and PDF output paths.~~
~~Standardized report bundle output.~~

## 6) Security and Policy Enforcement (High)
- Enforce policy in all commands, not just warnings.
- Ensure no secrets are logged or saved.

## 7) Config Validation and Migration (Medium)
~~Schema versioning for `pm-assist.yaml`.~~
- Migrations for old config/profile versions.

## 8) Test Coverage (Medium)
~~Unit tests for config/manifest/QA.~~
~~Smoke test with synthetic dataset (script added and wired into CI).~~

## 9) Signed Releases (Nice-to-have)
~~Signed releases + verification hooks (cosign).~~

## 10) UX/Platform Hardening (High)
- Windows venv path handling (use `.venv/Scripts` on Windows).
- Non-interactive TUI fallback (TTY detection; bypass Bubble Tea when not interactive).
- Directory picker should allow directory selection.
- Profile/business `set/show` should resolve sanitized filenames.
- Run-id reuse behavior when not explicitly provided.
