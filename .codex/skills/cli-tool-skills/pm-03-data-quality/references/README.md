# pm-03-data-quality Issue Fixes (R1.00)

This pack contains runbooks designed for automation agents to:

- recognise common environment setup failures from log output
- select a safe remediation path
- validate the fix with deterministic checks

## How agents should use this pack

1. Parse logs and extract error lines.
2. Match against `error_signatures` in `index.json` (substring or regex match).
3. Load the referenced markdown runbook.
4. Execute the safe default remediation first.
5. Run the verification commands and re-run the failed step.

## Safety defaults

- Prefer project virtual environments (`.venv`) over system or user-site installs.
- Prefer rebuilding broken venvs rather than in-place repairs.
- Avoid `--break-system-packages` unless explicitly requested.

## Contents

- `index.json`: mapping of issue IDs to signatures and runbook files
- `issue-fixes/*.md`: individual runbooks
