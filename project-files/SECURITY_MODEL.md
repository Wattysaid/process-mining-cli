# Security Model

**Document version:** R1.01 (2026-01-12)

## 1. Threat model (lightweight)
Assets:
- event data (often sensitive)
- derived artefacts (models, metrics, reports)
- credentials (LLM keys, connector tokens)

Threats:
- accidental data exfiltration (logs, LLM calls)
- credential leakage (console output, saved config)
- unauthorised access to outputs
- tampering with run outputs
- reverse engineering or modification of tool code

## 2. Security controls (MVP)
- Secrets redaction in logs
- Explicit opt-in for any external network calls
- Default read-only connectors
- Per-run artefact manifest (hashes optional post-MVP)
- Clear “what will be sent” prompt before LLM calls
- “Offline mode” always available
- Policy controls enforced via config (`policy.llm_enabled`, `policy.offline_only`, connector allow/deny lists)
- Bundle Python assets and avoid editable source distributions
- Only allow user edits in `.profiles/` and project data/output directories
- Do not expose internal source files via CLI commands or logs
- CLI UI uses local rendering only; no terminal content is streamed externally

Status: Policy gating implemented; structured logging and run manifests implemented.

## 3. Enterprise controls (post-MVP)
- OS keychain integration for secrets
- Signed releases and signature verification
- Pluggable policy engine:
  - prohibit LLM usage
  - prohibit sample uploads
  - restrict connectors
- Centralised audit sink (optional)

## 4. Security acceptance tests
- No secrets appear in logs or config snapshots
- LLM calls blocked if disabled
- Output folder contains reproducibility artefacts (config snapshot, run manifest)
