# OpenAI Integration: Security, cost control, and agent behaviour

**Document version:** R1.00 (2026-01-11)

## 1. Principles
- OpenAI integration is **optional** and **explicitly opted-in** by the user.
- The default mode must be **offline** (no external calls).
- Avoid sending raw data by default:
  - prefer aggregates, schema summaries, descriptive statistics
  - allow user to explicitly approve samples if needed

## 2. Credential handling
Supported mechanisms (MVP):
- Environment variable: `OPENAI_API_KEY`
- Project `.env` file (dev only; discourage for enterprise)

CLI behaviour:
- `pm-assist agent` and `pm-assist report --narrative` must:
  - check `OPENAI_API_KEY`
  - if missing, show steps:
    - `export OPENAI_API_KEY="..."` (Linux/macOS/WSL)
  - refuse to run if not set (unless user chooses offline narrative templates)

Never:
- print the key
- store it in config snapshots
- log it

## 3. Cost control
Must include:
- Per-run token budget (hard stop)
- Per-call max tokens
- Rate limiting / backoff
- A “dry run” that estimates the prompt size
- A local ledger file per run:
  - request count
  - estimated tokens in/out
  - model name

Recommended approach:
- Use a small number of “agent moves”:
  - interpret user goal
  - recommend next CLI command
  - draft narrative sections from already generated local artefacts
- Avoid repeated “chatty” loops.

## 4. Agent scope (what the agent is allowed to do)
Allowed:
- Explain options and trade-offs
- Suggest which pipeline steps to run
- Draft report text using local summaries and computed metrics
- Generate checklists and improvement backlogs

Not allowed (by default):
- Upload full raw event logs
- Upload sensitive identifiers (unless user explicitly approves)
- Execute destructive actions or modify source data
- Make final analytical decisions without user confirmation

## 5. Data minimisation strategy
When the agent needs context, provide in order:
1. Schema summary (column names, types, missingness)
2. Dataset shape and time range
3. Activity and variant frequency tables (top N)
4. Aggregate performance stats (median, p95, throughput)
5. Only if explicitly approved: anonymised samples

## 6. Implementation notes for Codex
- Use the official OpenAI SDK.
- Keep prompt templates in versioned files, not inline strings.
- Build an abstraction layer:
  - `LLMClient` with:
    - `enabled` flag
    - `budget` manager
    - `redaction` utilities
- Ensure “offline mode” produces reasonable outputs.

## 7. Compliance notes
- Provide a `SECURITY.md` and `PRIVACY.md` suitable for enterprise review.
- Make external endpoints explicit and auditable.
- Make it easy to disable OpenAI at build time and runtime.

