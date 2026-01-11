# LLM Integration: Security, cost control, and agent behaviour

**Document version:** R1.00 (2026-01-11)

## 1. Principles
- LLM integration is **optional** and **explicitly opted-in** by the user.
- The default mode must be **offline** (no external calls) unless a provider is configured.
- Avoid sending raw data by default:
  - prefer aggregates, schema summaries, descriptive statistics
  - allow user to explicitly approve samples if needed

## 2. Credential handling
Supported providers (MVP):
- OpenAI (`OPENAI_API_KEY`)
- Anthropic (`ANTHROPIC_API_KEY`)
- Gemini (`GEMINI_API_KEY` or `GOOGLE_API_KEY`)
- Ollama (local, no key; configurable via `OLLAMA_HOST`)

Supported mechanisms (MVP):
- Environment variables (preferred)
- Project `.env` file (dev only; discourage for enterprise)

Config keys (stored in `pm-assist.yaml`):
- `llm.provider`, `llm.model`, `llm.enabled`, `llm.offline`
- `llm.endpoint` (for Ollama or custom gateways)

CLI behaviour:
- `pm-assist agent` and `pm-assist report --narrative` must:
  - check provider configuration and required env vars
  - if missing, show steps:
    - `export OPENAI_API_KEY="..."` (Linux/macOS/WSL)
    - `export ANTHROPIC_API_KEY="..."`
    - `export GEMINI_API_KEY="..."`
    - `export OLLAMA_HOST="http://localhost:11434"`
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
- Use official provider SDKs where available (OpenAI, Anthropic, Google) and a simple HTTP client for Ollama.
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
- Make it easy to disable LLM usage at build time and runtime.
