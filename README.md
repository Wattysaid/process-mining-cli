# PM Assist CLI

PM Assist CLI is an interactive, enterprise-ready process mining
assistant that guides analysts through end-to-end engagements from data
access to discovery, conformance, performance analysis, and reporting.
It is a local CLI tool with optional AI assistance (OpenAI, Anthropic,
Gemini, or local Ollama) that never makes decisions for the user and
never executes destructive actions without confirmation.

The tool is designed to be a trusted, auditable companion: it asks the
right questions, proposes next steps, generates reproducible code and
notebooks, and keeps an immutable trail of what was run.

------------------------------------------------------------------------

## Highlights

-   **User-led decisions**: prompts for approval at each step; no silent
    actions.\
-   **Profiles & business context**: stores user and business profiles
    in `.profiles/` and `.business/`.\
-   **Optional AI**: use AI for summaries, interpretation, and task
    translation; keep data processing deterministic.\
-   **Cross-platform**: Linux, macOS, Windows (WSL2).\
-   **Enterprise-safe**: read-only connectors, secrets via env vars,
    audit-friendly outputs.\
-   **Signed releases**: all release checksums are cryptographically
    signed with Cosign.

------------------------------------------------------------------------

## Quick Start

### Install (Linux / macOS / WSL2)

``` bash
curl -fsSL https://raw.githubusercontent.com/Wattysaid/process-mining-cli/main/scripts/install.sh | sh
```

> The installer downloads the latest release tarball and verifies
> checksums before installing.

### Verify installation

``` bash
pm-assist doctor
pm-assist version
```

### Create a new project

``` bash
pm-assist init
```

### Run a guided workflow

``` bash
pm-assist connect
pm-assist ingest
pm-assist map
pm-assist prepare
pm-assist mine
pm-assist report
```

### Run the full wizard

``` bash
pm-assist start
```

------------------------------------------------------------------------

## How It Works

PM Assist is a local CLI that orchestrates Python-based skills. It
creates a project scaffold, registers data sources, and builds a
notebook as you confirm each step. The CLI uses AI only where it adds
value (summaries, interpretation, and intent translation), and relies on
deterministic Python scripts for data processing.

### Core flow

1.  **Initialize**: create project layout and profiles.\
2.  **Connect**: register read-only data sources.\
3.  **Ingest**: load and validate data.\
4.  **Map**: select case/activity/timestamp columns.\
5.  **Prepare**: clean and transform event logs.\
6.  **Mine**: discovery, conformance, performance, variant analysis.\
7.  **Report**: narrative insights and artefacts.\
8.  **Review**: automated QA checks and audit summary.

------------------------------------------------------------------------

## Example Scenarios

### Scenario 1: Procurement bottlenecks

You are asked to identify bottlenecks in the procurement process.

1.  `pm-assist init` creates a project and asks about your role,
    experience level, and target business.\
2.  `pm-assist connect` asks how to access procurement data (ERP, CSV
    exports, database).\
3.  `pm-assist ingest` validates tables and schema.\
4.  `pm-assist map` captures the case, activity, and timestamp columns.\
5.  `pm-assist prepare` cleans timestamps, filters invalid rows, and
    normalises activity names.\
6.  `pm-assist mine` highlights the longest waiting steps and most
    frequent rework loops.\
7.  `pm-assist report` creates a summary report and appends all code to
    the notebook.

### Scenario 2: Duplicate events in a dataset

1.  `pm-assist connect` registers the CSV.\
2.  `pm-assist ingest` previews rows and verifies schema.\
3.  `pm-assist prepare` scans duplicates and proposes a fix.\
4.  You approve, and the notebook is updated with the exact code used.

### Scenario 3: Executive-ready report

1.  Run your mining steps.\
2.  `pm-assist report` generates a concise Markdown report.\
3.  If AI is enabled, it drafts a narrative aligned to your industry.

------------------------------------------------------------------------

## Configuration and Profiles

-   **User profiles**: stored in `.profiles/` as YAML.\
-   **Business profiles**: stored in `.business/` as YAML.\
-   **Project config**: `pm-assist.yaml` in the project root.

The CLI uses these profiles to tailor prompts, defaults, and
recommendations.

------------------------------------------------------------------------

## Security and Privacy

-   Credentials are never stored in plaintext; use environment
    variables.\
-   Connectors are read-only by default.\
-   AI use is optional and policy-controlled.\
-   All outputs are local, deterministic, and audit-friendly.\
-   Release artefacts are checksumed and cryptographically signed.

------------------------------------------------------------------------

## Verifying Downloads (Supply Chain Security)

Each release includes:

-   `checksums.txt`\
-   `checksums.txt.sig` (Cosign signature)

The public verification key is published in this repo:

    project-files/keys/cosign.pub

### Verify checksums file

After downloading release assets:

``` bash
cosign verify-blob   --key project-files/keys/cosign.pub   --signature checksums.txt.sig   checksums.txt
```

If verification passes, you can trust the checksum file.

### Verify a tarball against checksums

``` bash
sha256sum pm-assist_linux_x64.tar.gz
cat checksums.txt
```

Ensure the hash matches the corresponding line in `checksums.txt`.

------------------------------------------------------------------------

## How Releases Work (For Maintainers)

Releases are fully automated via GitHub Actions.

### Release pipeline

1.  Push a version tag: `vX.Y.Z`\
2.  GitHub Actions builds binaries for all platforms\
3.  Artefacts are packaged into `.tar.gz`\
4.  `checksums.txt` is generated\
5.  `checksums.txt` is signed using Cosign\
6.  GitHub Release is created with all assets

### Triggering a new release

From the repo root:

``` bash
git tag v1.2.3
git push origin v1.2.3
```

That is all that is required. Do **not** manually create the GitHub
Release first.

------------------------------------------------------------------------

## Adding Release Notes (For Maintainers)

Release notes are controlled by GitHub Releases and can be added in two
ways:

### Option A: Edit release after automation

1.  Go to GitHub → Releases\
2.  Open the generated release\
3.  Click **Edit release**\
4.  Add:
    -   Summary
    -   Breaking changes
    -   Migration notes
    -   Known issues

This does not affect artefacts or signatures.

### Option B: Auto-generate release notes

You can let GitHub auto-generate notes from PR titles and commits:

-   Enable **Generate release notes** in repository settings, or
-   Modify the workflow to include:

``` yaml
with:
  generate_release_notes: true
```

in the `softprops/action-gh-release` step.

------------------------------------------------------------------------

## Adding Extra Files to the Release

If you want to ship extra artefacts (for example:

-   installer scripts
-   config templates
-   example datasets

Add them in the release job before the upload step and include them in
the file list:

``` yaml
with:
  files: |
    release/*.tar.gz
    release/checksums.txt
    release/checksums.txt.sig
    release/install.sh
```

Anything listed there will be attached to the GitHub Release.

------------------------------------------------------------------------

## Updating the Installer

If you host an installer script (for example `install.sh`):

1.  Ensure it downloads:
    -   the correct platform tarball
    -   `checksums.txt`
2.  Verify checksum before extracting
3.  Optionally verify `checksums.txt` using `cosign.pub`

Typical flow inside `install.sh`:

-   Download artefacts
-   Verify signature of `checksums.txt`
-   Verify tarball hash
-   Extract binary to `/usr/local/bin`

If the release asset names change, update the installer logic
accordingly.

------------------------------------------------------------------------

## Repository Layout

-   `cmd/` and `internal/`: Go CLI source code.\
-   `python/`: Python workflows and scripts.\
-   `.codex/skills/cli-tool-skills/`: minimum skills library.\
-   `project-files/`: product specs, architecture docs, and signing
    keys.\
-   `scripts/`: installer and packaging scripts.\
-   `.github/workflows/`: CI and release pipelines.

------------------------------------------------------------------------

## Development

### Go Version Requirement
PM Assist requires Go **1.23+**. Some cloud connector SDKs (BigQuery, Snowflake) and their transitive dependencies require Go 1.23 or newer, so earlier versions will fail `go mod tidy`.

Recommended toolchain behavior:
- `GOTOOLCHAIN=auto` (default) will download the required Go toolchain.
- `GOTOOLCHAIN=local` forces the locally installed Go version and will fail if it is < 1.23.

WSL2 install (example):
```bash
sudo apt-get update
sudo apt-get install -y golang-go
go version
```
If your distro packages an older Go, install via tarball or a version manager (asdf/gvm) to get 1.23+.

``` bash
go build ./cmd/pm-assist
./pm-assist version
```

When working under Windows, prefer developing inside the WSL filesystem
(`/home/...`) rather than `/mnt/c` or `/mnt/d` to avoid file permission
issues with shell scripts.

### Testing the CLI

Build + health checks:
```bash
go build ./cmd/pm-assist
./pm-assist version
./pm-assist doctor
```

Interactive flow (no data yet):
```bash
./pm-assist
./pm-assist profile init
./pm-assist business init
./pm-assist agent setup
```

Sample dataset smoke run:
1) Create a project
```bash
./pm-assist init
```

2) Register the bundled sample CSV
```bash
./pm-assist connect --type file --name sample --paths .codex/skills/cli-tool-skills/pm-02-ingest-profile/assets/sample_log.csv --format csv
```

3) Run the pipeline
```bash
./pm-assist ingest --connector sample
./pm-assist prepare
./pm-assist mine
./pm-assist report --html true --pdf false
```

4) Inspect outputs
- `outputs/<run-id>/analysis_notebook.ipynb`
- `outputs/<run-id>/stage_09_report/`
- `outputs/<run-id>/run_manifest.json`
- `outputs/<run-id>/config_snapshot.yaml`

### Troubleshooting Toolchains
If you see errors like “requires go >= 1.23.0”:
- Ensure `go version` reports 1.23+.
- If you cannot upgrade locally, run with `GOTOOLCHAIN=auto` so Go downloads the required toolchain:
```bash
GOTOOLCHAIN=auto go mod tidy
GOTOOLCHAIN=auto go test ./...
```

------------------------------------------------------------------------

## License

Proprietary. Do not distribute without permission.
