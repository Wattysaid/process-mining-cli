# Installation and Release Engineering

**Document version:** R1.00 (2026-01-11)

## 1. Distribution strategy (recommended)
- CLI is a single binary (Go recommended).
- Release artefacts published per OS/arch.
- Installer script downloads the correct artefact and installs it to `~/.local/bin`.
- First run creates a project-local `.venv` unless the user chooses a shared env.

## 2. Installer requirements (curl | sh)
The installer must:
- Detect platform (`uname -s`, `uname -m`)
- Validate prerequisites (`curl`, `unzip` or `tar`, `ca-certificates`)
- Fetch a version manifest (or use GitHub Releases “latest”)
- Download and verify checksum
- Install binary to `~/.local/bin/pm-assist`
- Ensure PATH includes `~/.local/bin` by updating shell rc file
- Print “next steps” including LLM setup (OpenAI/Anthropic/Gemini/Ollama)

Installer output conventions:
- `[INFO]`, `[SUCCESS]`, `[WARN]`, `[ERROR]`

## 3. GitHub Releases build matrix
MVP artefacts:
- linux-x64
- linux-arm64
- darwin-x64
- darwin-arm64
- windows-x64 (optional for native; WSL2 is acceptable for MVP)

## 4. Release pipeline (GitHub Actions)
- On tag `vX.Y.Z`:
  - build binaries
  - compute SHA256 for each artefact
  - publish release with assets:
    - `pm-assist_<os>_<arch>.zip`
    - `checksums.txt`
- On main branch:
  - run unit tests and linting
  - run minimal “smoke” CLI tests (doctor/version)

## 5. Update mechanism
Provide:
- `pm-assist self-update`
  - checks latest version
  - downloads and replaces the binary
  - re-verifies checksum

## 6. Python dependency management
Two viable approaches (Codex to choose, with rationale):
A) CLI manages a venv and installs Python package via pip (network required)
B) CLI ships Python wheels in release assets and installs offline (preferred for enterprise)

MVP can start with (A) and add (B) post-MVP. For IP protection, prefer shipping bundled wheels or embedded assets and avoid editable source distributions.

## 7. Graphviz dependency
pm4py visualisations may require OS-level Graphviz.
- Provide `pm-assist doctor` checks.
- Provide clear OS-specific install guidance.
