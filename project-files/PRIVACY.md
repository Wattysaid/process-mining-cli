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

## 3. LLM usage
- Disabled by default.
- When enabled (OpenAI/Anthropic/Gemini/Ollama):
  - send summaries first (schema and aggregates)
  - require explicit approval for samples
  - record a ledger of calls and prompts (redacted) for audit

## 4. Retention
- Provide `pm-assist clean --run-id <id>` to remove run artefacts (post-MVP).
- Provide retention policies via config (post-MVP).

## 5. User profiles
- Profiles are stored in `.profiles/` as YAML.
- Store only user-provided metadata (name, role, aptitude, preferences).
- Avoid storing secrets or external credentials in profiles.

## 6. Business profiles
- Profiles are stored in `.business/` as YAML.
- Store only business metadata and system context, never credentials.
