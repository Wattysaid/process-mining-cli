---
id: ISSUE-006
version: R1.00
date_created: 2026-01-08
owner: Platform Tooling
type: runbook
title: "_distutils_hack missing or setuptools incomplete inside venv (distutils-precedence.pth)"
tags: ["linux", "venv", "setuptools", "distutils", "pip"]
error_signatures:
- "No module named '_distutils_hack'"
- "distutils-precedence.pth"
- "_distutils_hack missing"
---

## Symptoms

- `ModuleNotFoundError: No module named '_distutils_hack'`
- Error references `distutils-precedence.pth`
- `python -m pip --version` fails inside the venv after attempted upgrades or repairs.

## Likely causes

- `setuptools` is partially installed or inconsistent with the venv state.
- A previous attempt used mixed installation methods (for example, `pip --target`, copying site-packages).
- venv bootstrap was interrupted and later patched inconsistently.

## Fast checks

```bash
. .venv/bin/activate
python -c "import setuptools; print(setuptools.__version__)" || echo "setuptools broken"
python -c "import _distutils_hack; print('distutils hack OK')" || echo "missing _distutils_hack"
python -m pip -V || echo "pip failing"
```

## Remediation (safe default)

### Option A (recommended): rebuild venv cleanly

```bash
deactivate 2>/dev/null || true
rm -rf .venv
python3 -m venv .venv
. .venv/bin/activate
python -m pip install -U pip setuptools wheel
python -m pip install -r requirements.txt
```

### Option B: force reinstall inside the venv (only if pip still runs)

```bash
. .venv/bin/activate
python -m pip install --force-reinstall --no-cache-dir -U setuptools wheel
python -m pip install --force-reinstall --no-cache-dir -U pip
```

If pip cannot run, Option B cannot proceed.

## Verification

```bash
. .venv/bin/activate
python -c "import _distutils_hack; print('OK')"
python -m pip -V
```

## Agent notes

- Treat `_distutils_hack` errors as a strong signal the venv is non-deterministic. Rebuild is usually faster overall.
- If the repo is on `/mnt/` under WSL, also apply ISSUE-002 to prevent recurrence.
