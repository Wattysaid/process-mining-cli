---
id: ISSUE-001
version: R1.00
date_created: 2026-01-08
owner: Platform Tooling
type: runbook
title: "python: command not found (exit 127) when scripts call `python`"
tags: ["linux", "wsl", "python", "env-setup"]
error_signatures:
- "python: command not found"
- "exit code 127"
- "/bin/bash: line 1: python: command not found"
---

## Symptoms

- Setup or validation scripts fail with exit code **127**.
- Error text contains `python: command not found`.
- `python3` exists, but `python` does not.

## Likely causes (most common first)

- Ubuntu/Debian does not ship a `python` alias by default; only `python3` is installed.
- A script hardcodes `python` rather than using `python3` or a virtual environment interpreter.
- `$PATH` is modified and the interpreter is not resolvable.

## Fast checks

Run from the repo root:

```bash
command -v python || echo "python missing"
command -v python3 || echo "python3 missing"
python3 --version
```

If you are using a virtual environment:

```bash
ls -la .venv/bin/python .venv/bin/python3 2>/dev/null || true
```

## Remediation (safe default)

### Option A (recommended): call the venv interpreter explicitly

If the project uses `.venv`:

```bash
. .venv/bin/activate
python -V
python -m pip -V
python path/to/validator.py
```

If activation is undesirable, invoke directly:

```bash
.venv/bin/python path/to/validator.py
```

### Option B: update scripts to be portable

Preferred patterns:

- Use `python3` explicitly, or
- Use a shebang that targets venvs:

```bash
#!/usr/bin/env python3
```

If you control the script, avoid hardcoding `python`.

### Option C: install a `python` alias (system-level change)

On Ubuntu:

```bash
sudo apt update
sudo apt install -y python-is-python3
```

This makes `python` resolve to `python3`. Use this only if it aligns with your organisation’s standard.

## Verification

```bash
python --version || true
python3 --version
```

If venv-based:

```bash
. .venv/bin/activate
which python
python -c "import sys; print(sys.executable)"
```

## Agent notes

- Prefer **Option A**. It is the least invasive and respects per-project isolation.
- Only propose **Option C** if the user wants `python` available system-wide.
