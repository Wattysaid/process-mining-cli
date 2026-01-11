#!/usr/bin/env bash
set -euo pipefail

PM_ASSIST_VERSION="${PM_ASSIST_VERSION:-latest}"
PM_ASSIST_BASE_URL="${PM_ASSIST_BASE_URL:-https://github.com/pm-assist/pm-assist/releases}"
INSTALL_DIR="${PM_ASSIST_INSTALL_DIR:-$HOME/.local/bin}"
SELF_UPDATE="false"

usage() {
  cat <<USAGE
Usage: install.sh [--version <version>] [--self-update]

Environment variables:
  PM_ASSIST_VERSION   Version to install (default: latest)
  PM_ASSIST_BASE_URL  Base release URL (default: GitHub releases)
  PM_ASSIST_INSTALL_DIR Install directory (default: ~/.local/bin)
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      PM_ASSIST_VERSION="$2"
      shift 2
      ;;
    --self-update)
      SELF_UPDATE="true"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "[ERROR] Unknown option: $1"
      usage
      exit 1
      ;;
  esac
done

log_info() { echo "[INFO] $*"; }
log_warn() { echo "[WARN] $*"; }
log_error() { echo "[ERROR] $*"; }

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    log_error "Missing required command: $1"
    exit 1
  fi
}

require_cmd curl
require_cmd tar

if command -v sha256sum >/dev/null 2>&1; then
  SHA256="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  SHA256="shasum -a 256"
else
  log_error "Missing sha256sum or shasum"
  exit 1
fi

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="x64";;
  arm64|aarch64) ARCH="arm64";;
  *) log_error "Unsupported architecture: $ARCH"; exit 1;;
 esac

case "$OS" in
  linux|darwin) ;; 
  *) log_error "Unsupported OS: $OS"; exit 1;;
 esac

ASSET="pm-assist_${OS}_${ARCH}.tar.gz"
CHECKSUMS="checksums.txt"

if [[ "$PM_ASSIST_VERSION" == "latest" ]]; then
  BASE_URL="$PM_ASSIST_BASE_URL/latest/download"
else
  BASE_URL="$PM_ASSIST_BASE_URL/download/$PM_ASSIST_VERSION"
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

log_info "Downloading $ASSET"
curl -fsSL "$BASE_URL/$ASSET" -o "$TMP_DIR/$ASSET"
log_info "Downloading checksums"
curl -fsSL "$BASE_URL/$CHECKSUMS" -o "$TMP_DIR/$CHECKSUMS"

EXPECTED=$(grep " $ASSET" "$TMP_DIR/$CHECKSUMS" | awk '{print $1}')
if [[ -z "$EXPECTED" ]]; then
  log_error "Checksum not found for $ASSET"
  exit 1
fi

ACTUAL=$($SHA256 "$TMP_DIR/$ASSET" | awk '{print $1}')
if [[ "$EXPECTED" != "$ACTUAL" ]]; then
  log_error "Checksum verification failed"
  exit 1
fi
log_info "Checksum verified"

mkdir -p "$INSTALL_DIR"

tar -xzf "$TMP_DIR/$ASSET" -C "$TMP_DIR"
if [[ ! -f "$TMP_DIR/pm-assist" ]]; then
  log_error "Binary not found in archive"
  exit 1
fi

if [[ "$SELF_UPDATE" == "true" ]]; then
  TARGET="$(command -v pm-assist || true)"
  if [[ -z "$TARGET" ]]; then
    TARGET="$INSTALL_DIR/pm-assist"
  fi
else
  TARGET="$INSTALL_DIR/pm-assist"
fi

mv "$TMP_DIR/pm-assist" "$TARGET"
chmod +x "$TARGET"

if [[ "$SELF_UPDATE" != "true" ]]; then
  if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    SHELL_NAME="$(basename "$SHELL")"
    if [[ "$SHELL_NAME" == "zsh" ]]; then
      RC_FILE="$HOME/.zshrc"
    else
      RC_FILE="$HOME/.bashrc"
    fi
    log_warn "Adding $INSTALL_DIR to PATH in $RC_FILE"
    echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$RC_FILE"
  fi
fi

log_info "Installed pm-assist to $TARGET"
log_info "Next steps:"
log_info "  pm-assist doctor"
log_info "  pm-assist init"
