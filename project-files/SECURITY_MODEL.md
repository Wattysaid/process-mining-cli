# Security Model

**Document version:** R1.00 (2026-01-11)

## 1. Threat model (lightweight)
Assets:
- event data (often sensitive)
- derived artefacts (models, metrics, reports)
- credentials (OpenAI key, connector tokens)

Threats:
- accidental data exfiltration (logs, LLM calls)
- credential leakage (console output, saved config)
- unauthorised access to outputs
- tampering with run outputs

## 2. Security controls (MVP)
- Secrets redaction in logs
- Explicit opt-in for any external network calls
- Default read-only connectors
- Per-run artefact manifest (hashes optional post-MVP)
- Clear “what will be sent” prompt before OpenAI calls
- “Offline mode” always available

## 3. Enterprise controls (post-MVP)
- OS keychain integration for secrets
- Signed releases and signature verification
- Pluggable policy engine:
  - prohibit OpenAI usage
  - prohibit sample uploads
  - restrict connectors
- Centralised audit sink (optional)

## 4. Security acceptance tests
- No secrets appear in logs or config snapshots
- OpenAI calls blocked if disabled
- Output folder contains reproducibility artefacts (config snapshot, run manifest)

