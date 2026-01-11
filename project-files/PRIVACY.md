# Privacy and Data Handling

**Document version:** R1.00 (2026-01-11)

## 1. Data minimisation
- Only ingest columns required for the selected analyses.
- Provide column exclusion rules in config.
- Prefer hashing or pseudonymisation for identifiers when producing shareable artefacts.

## 2. Default behaviours
- Do not transmit data externally.
- Do not auto-profile with third-party services.
- Do not persist raw extracts outside the run folder unless requested.

## 3. OpenAI usage
- Disabled by default.
- When enabled:
  - send summaries first (schema and aggregates)
  - require explicit approval for samples
  - record a ledger of calls and prompts (redacted) for audit

## 4. Retention
- Provide `pm-assist clean --run-id <id>` to remove run artefacts (post-MVP).
- Provide retention policies via config (post-MVP).

