---
id: ISSUE-005
version: R1.00
date_created: 2026-01-08
owner: Platform Tooling
type: runbook
title: "PEP 668: externally-managed-environment blocks system or user-site pip installs"
tags: ["linux", "ubuntu", "pep668", "pip", "packaging"]
error_signatures:
- "error: externally-managed-environment"
- "PEP 668"
- "Externally Managed Environment"
- "pip install --user"
---

## Symptoms

- `python3 -m pip install ...` fails with `error: externally-managed-environment`.
- `pip install --user ...` is blocked on Ubuntu 23.04+ (including 24.04).
- Installing Python packages outside a venv is refused unless `--break-system-packages` is used.

## Why this happens

- Debian/Ubuntu mark the system Python environment as externally managed to protect OS-managed packages.
- This is expected behaviour under PEP 668.

## Fast checks

Confirm you are in a venv:

```bash
python3 -c "import sys; print(sys.prefix); print(sys.base_prefix)"
# In a venv, sys.prefix != sys.base_prefix
```

## Remediation (safe default)

### Option A (recommended): use a project virtual environment

From repo root:

```bash
python3 -m venv .venv
. .venv/bin/activate
python -m pip install -U pip setuptools wheel
python -m pip install -r requirements.txt
```

### Option B: use apt for system packages (only when appropriate)

Example:

```bash
sudo apt update
sudo apt install -y python3-pandas
```

This works for common libraries but may not satisfy version needs for specialised tooling.

### Option C: override PEP 668 (not recommended)

Only if explicitly requested and risks are accepted:

```bash
python3 -m pip install --break-system-packages <package>
```

## Verification

```bash
. .venv/bin/activate
python -m pip -V
```

## Agent notes

- Always prefer venv-based installs for tooling and agents.
- Never propose `--break-system-packages` as a default remediation.
