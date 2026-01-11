---
id: ISSUE-002
version: R1.00
date_created: 2026-01-08
owner: Platform Tooling
type: runbook
title: "venv creation or dependency install times out (exit 124) on WSL or slow filesystems"
tags: ["linux", "wsl", "venv", "pip", "performance", "timeouts"]
error_signatures:
- "exit code 124"
- "timeout while creating the venv"
- "command exceeded 10s"
- "timed out after"
---

## Symptoms

- `venv` creation or installs fail with exit code **124** or a timeout message.
- Partial environment state is left behind (for example, `.venv` exists but is incomplete).
- This is commonly observed when working from `/mnt/<drive>/...` in WSL.

## Likely causes

- The repo is on the Windows-mounted filesystem (`/mnt/d`, `/mnt/c`), which is slower for many small file operations (pip installs create thousands of files).
- The orchestration script has an aggressive timeout (for example, 10 seconds) which is too short for venv bootstrap and first install.
- Network latency or intermittent package index access.

## Fast checks

Identify where the repo lives:

```bash
pwd
# If it starts with /mnt/, you are on a Windows mount.
```

Measure basic filesystem throughput (quick signal, not a benchmark):

```bash
time python3 -c "import os; [open('tmp_'+str(i),'w').write('x') for i in range(2000)]; print('ok')"
rm -f tmp_*
```

## Remediation (safe default)

### Option A (recommended): move the repo to the WSL Linux filesystem

From WSL:

```bash
mkdir -p ~/repos
# Best: re-clone into Linux FS
cd ~/repos
git clone <repo-url>
```

If you must keep the original, copy it:

```bash
cp -a /mnt/d/GitHub/<repo> ~/repos/<repo>
cd ~/repos/<repo>
```

Then recreate the venv:

```bash
rm -rf .venv
python3 -m venv .venv
. .venv/bin/activate
python -m pip install -U pip setuptools wheel
python -m pip install -r requirements.txt
```

### Option B: increase the timeout in the calling script

- Increase the venv bootstrap timeout to **60 to 180 seconds** for first-run installs.
- Implement retry with backoff for network steps.

### Option C: reduce first-run work

- Pin and cache wheels where possible.
- Use a lockfile and a local wheelhouse if operating offline.

## Verification

- Re-run venv creation and a minimal pip action:

```bash
python3 -m venv .venv
. .venv/bin/activate
python -m pip -V
```

## Agent notes

- If the path starts with `/mnt/`, recommend **Option A** immediately.
- Treat timeouts as a leading indicator for later corruption (pip vendor tree, missing wrappers).
