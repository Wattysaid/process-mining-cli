---
id: ISSUE-004
version: R1.00
date_created: 2026-01-08
owner: Platform Tooling
type: runbook
title: "pip is corrupted inside the venv (pip._vendor.urllib3 ImportError)"
tags: ["linux", "venv", "pip", "corruption", "urllib3"]
error_signatures:
- "ImportError: cannot import name 'exceptions' from partially initialized module 'pip._vendor.urllib3'"
- "pip._vendor.urllib3"
- "partially initialized module"
---

## Symptoms

- `python -m pip ...` fails inside a venv with an ImportError referencing `pip._vendor.urllib3`.
- `pip._vendor.urllib3` appears incomplete (missing expected modules).
- Re-running pip commands produces inconsistent errors.

## Likely causes

- Partial install or interrupted upgrade of pip (common after timeouts or slow filesystem writes).
- Mixing installation methods (for example, `pip --target`, copying site-packages, or combining `ensurepip` with manual vendor overwrites).
- Using a venv located on `/mnt/<drive>` in WSL, increasing the chance of interrupted small-file operations.

## Fast checks

Activate venv and inspect:

```bash
. .venv/bin/activate
python -m pip -V || true
python -c "import pip; print(pip.__version__)" || true
python -c "import pip._vendor.urllib3 as u; print(u.__file__)" || true
```

## Remediation (safe default)

The safest fix is to recreate the venv cleanly. Do not attempt in-place patching unless you must.

### Option A (recommended): rebuild the venv

```bash
deactivate 2>/dev/null || true
rm -rf .venv
python3 -m venv .venv
. .venv/bin/activate
python -m pip install -U pip setuptools wheel
python -m pip install -r requirements.txt
```

If on WSL and the repo is on `/mnt/`, first apply ISSUE-002 (move to Linux FS) before rebuilding.

### Option B: in-place recovery (only if pip still runs)

```bash
. .venv/bin/activate
python -m ensurepip --upgrade
python -m pip install --force-reinstall --no-cache-dir -U pip setuptools wheel
```

If pip itself cannot run, Option B typically cannot proceed.

## Verification

```bash
. .venv/bin/activate
python -m pip -V
python -c "import pip._vendor.urllib3 as u; print('pip vendor urllib3 OK')"
```

## Agent notes

- Treat pip vendor ImportErrors as an indicator of an unrecoverable venv state.
- Prefer rebuild over repair to reduce time spent on non-deterministic fixes.
