# Contributing Guide (for internal and partner engineering teams)

**Document version:** R1.00 (2026-01-11)

## Branching
- `main`: stable
- `dev`: integration
- feature branches: `feat/<name>`
- fix branches: `fix/<name>`

## Pull requests
A PR must include:
- tests (or justification if not feasible)
- updates to docs if behaviour changes
- screenshots / snippets for CLI output changes

## Definition of Done
- Command help text updated
- Unit tests cover new logic
- Smoke test passes on Linux
- No secrets in logs
- Outputs remain deterministic

