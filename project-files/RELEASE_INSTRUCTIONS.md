# Release Instructions (GitHub Releases)

**Document version:** R1.01 (2026-01-12)

This file explains how to produce release artifacts that the installer expects.

## 1) Asset names and structure
Installer expects:
- `pm-assist_<os>_<arch>.tar.gz`
- `checksums.txt`

Where:
- `<os>` is `linux` or `darwin`
- `<arch>` is `x64` or `arm64`

Each tarball must contain a single binary named `pm-assist` at the archive root:
```
pm-assist
```

Example assets:
- `pm-assist_linux_x64.tar.gz`
- `pm-assist_linux_arm64.tar.gz`
- `pm-assist_darwin_x64.tar.gz`
- `pm-assist_darwin_arm64.tar.gz`
- `checksums.txt`
- `checksums.txt.sig` (optional, for signature verification)

`checksums.txt` must contain SHA256 sums in the format:
```
<sha256>  pm-assist_linux_x64.tar.gz
<sha256>  pm-assist_linux_arm64.tar.gz
<sha256>  pm-assist_darwin_x64.tar.gz
<sha256>  pm-assist_darwin_arm64.tar.gz
```

## 2) Build prerequisites
- Go 1.23+
- Access to module downloads (GitHub Actions or local network)

Optional:
- `sha256sum` or `shasum -a 256`

## 3) Manual build (local)
From repo root:
```bash
VERSION=v0.1.0
OUTDIR=dist/$VERSION
mkdir -p "$OUTDIR"

build() {
  GOOS=$1 GOARCH=$2 \
    go build -buildvcs=false -trimpath -ldflags "-s -w" -o pm-assist ./cmd/pm-assist
  TAR_OS=$3
  TAR_ARCH=$4
  tar -czf "$OUTDIR/pm-assist_${TAR_OS}_${TAR_ARCH}.tar.gz" pm-assist
  rm -f pm-assist
}

build linux amd64 linux x64
build linux arm64 linux arm64
build darwin amd64 darwin x64
build darwin arm64 darwin arm64

(
  cd "$OUTDIR"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum pm-assist_*.tar.gz > checksums.txt
  else
    shasum -a 256 pm-assist_*.tar.gz > checksums.txt
  fi
)
```

## 4) GitHub release creation (manual)
1) Create a git tag:
   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```
2) Create a GitHub Release for the tag.
3) Upload all tarballs + `checksums.txt` to the release.

## 5) GitHub Actions (recommended)
Create `.github/workflows/release.yml` that:
1) Runs on tag `v*`.
2) Builds the four binaries.
3) Packages `cli-tool-skills` into `resources/cli-tool-skills` using `scripts/package_resources.sh`.
4) Produces the tarballs and `checksums.txt`.
5) Creates a release and uploads assets.

Minimal build matrix:
- `linux/amd64`, `linux/arm64`
- `darwin/amd64`, `darwin/arm64`

## 6) Installer compatibility
The installer script `scripts/install.sh` uses:
- `PM_ASSIST_BASE_URL` (default GitHub releases)
- `PM_ASSIST_VERSION` (`latest` or a tag like `v0.1.0`)

Expected download URLs:
```
https://github.com/pm-assist/pm-assist/releases/latest/download/pm-assist_linux_x64.tar.gz
https://github.com/pm-assist/pm-assist/releases/latest/download/checksums.txt
```

## 7) Packaging note (Python assets)
The CLI resolves skills in this order:
1) `PM_ASSIST_SKILLS_DIR` env var
2) `<project>/.codex/skills/cli-tool-skills`
3) `<binary>/resources/cli-tool-skills`

For releases, run:
```bash
scripts/package_resources.sh .codex/skills/cli-tool-skills resources/cli-tool-skills
scripts/package_wheels.sh .codex/skills/cli-tool-skills/pm-99-utils-and-standards/requirements.txt resources/wheels
```

Ensure the packaged `resources/` directory (including `resources/wheels`) is shipped alongside the binary in the tarball.

## 8) Signature verification (optional)
If you sign releases, upload `checksums.txt.sig` and distribute the public key.
Installers and `pm-assist self-update` can verify signatures when provided a cosign public key.
Release workflow supports signing when `COSIGN_PRIVATE_KEY` and `COSIGN_PASSWORD` secrets are set.
