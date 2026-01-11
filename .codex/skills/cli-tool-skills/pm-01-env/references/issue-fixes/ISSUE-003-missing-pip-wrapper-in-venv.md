---
id: ISSUE-003
version: R1.00
date_created: 2026-01-08
owner: Platform Tooling
type: runbook
title: "FileNotFoundError: .venv/bin/pip missing after venv creation"
tags: ["linux", "venv", "pip", "bootstrap"]
error_signatures:
- "FileNotFoundError: .venv/bin/pip"
- ".venv/bin/pip: No such file or directory"
- "pip was never installed inside the venv"
---

## Symptoms

- `.venv` exists, but `.venv/bin/pip` does not.
- Setup scripts fail when they assume a pip wrapper at `.venv/bin/pip`.
- Running `python -m pip` inside the venv may also fail.

## Likely causes

- `python3-venv` is missing, so the venv cannot bootstrap pip via `ensurepip`.
- venv bootstrap was interrupted (timeout, slow filesystem, killed process).
- The environment was created with a stripped Python build that does not include `ensurepip`.

## Fast checks

```bash
python3 -c "import venv; print('venv ok')"
python3 -c "import ensurepip; print('ensurepip ok')" || echo "ensurepip missing"
dpkg -l | grep -E "python3-venv|python3-full" || true
```

Inspect venv contents:

```bash
ls -la .venv/bin | egrep "python|pip" || true
```

## Remediation (safe default)

1) Remove the broken venv:

```bash
rm -rf .venv
```

2) Install required system packages:

```bash
sudo apt update
sudo apt install -y python3-venv
```

If `ensurepip` is still missing or you want a fuller bundle:

```bash
sudo apt install -y python3-full
```

3) Recreate the venv and upgrade packaging tools:

```bash
python3 -m venv .venv
. .venv/bin/activate
python -m pip install -U pip setuptools wheel
```

4) Install requirements:

```bash
python -m pip install -r requirements.txt
```

## Verification

```bash
. .venv/bin/activate
python -m pip -V
which pip
pip -V
```

## Agent notes

- Prefer `python -m pip ...` rather than calling `.venv/bin/pip` directly in scripts.
- If this recurs and the repo is on `/mnt/`, also apply ISSUE-002.
