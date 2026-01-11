---
id: ISSUE-007
version: R1.00
date_created: 2026-01-08
owner: Platform Tooling
type: runbook
title: "PEP 668: externally-managed-environment on Ubuntu 24.04+ prevents system pip installs"
tags: ["linux", "ubuntu", "python", "pip", "venv", "pep-668"]
error_signatures:
- "error: externally-managed-environment"
- "PEP 668"
- "Externally Managed Environment"
- "This environment is externally managed"
---

## Symptoms

- `pip install` commands outside a virtual environment fail with PEP 668 error.
- Ubuntu 24.04 (and newer) marks system Python as "externally managed".
- Error message: "error: externally-managed-environment".

## Likely causes (most common first)

- Ubuntu 24.04+ enforces PEP 668 by default, blocking system-wide pip installs.
- User attempting to install packages without activating a virtual environment.
- Missing `python3-venv` or `python3-full` package preventing venv creation.

## Fast checks

Check if PEP 668 is enforced:

```bash
python3 -m pip install --dry-run pandas 2>&1 | grep -i "externally-managed"
```

Check if venv module is available:

```bash
python3 -m venv --help >/dev/null 2>&1 && echo "venv available" || echo "venv missing"
```

## Remediation (safe default)

### Option A (recommended): create and use a project virtual environment

From your repo root:

```bash
cd /mnt/d/GitHub/processminingagenttest

# 1) Create a venv in the project
python3 -m venv .venv

# 2) Activate it (bash/zsh)
source .venv/bin/activate

# 3) Confirm you are now using the venv python/pip
which python
which pip
python -m pip -V

# 4) Upgrade packaging tools inside the venv
python -m pip install -U pip setuptools wheel

# 5) Install your skill requirements inside the venv
python -m pip install -r .codex/skills/process-mining-assistant/requirements.txt
```

Quick sanity check:

```bash
python -c "import pandas; print('\"'pandas OK'\"')"
```

To deactivate the venv:

```bash
deactivate
```

Common gotcha: if you forget to `source .venv/bin/activate`, you will keep using the system pip and keep seeing externally-managed-environment.

### Option B: install python3-full if venv creation fails

If `python3 -m venv .venv` fails, install the "full" Python bundle (covers venv and common tooling):

```bash
sudo apt-get update
sudo apt-get install -y python3-full
```

Then rerun the venv steps above.

### Option C: use apt for system packages (not recommended for development)

If you must install system-wide (requires sudo):

```bash
sudo apt install -y python3-pandas
```

**Note:** This approach is not recommended for development libraries like pm4py, as apt often doesn't have the versions you need. Use virtual environments instead.

## Verification

After creating and activating the venv:

```bash
source .venv/bin/activate
which python
which pip
python -m pip list
python -c "import sys; print(sys.prefix)"
```

The output should show paths inside `.venv/`, not `/usr/`.

## Agent notes

- Always recommend creating a virtual environment first on Ubuntu 24.04+.
- If venv creation fails, suggest installing `python3-full`.
- Emphasize the importance of activating the venv before installing packages.
