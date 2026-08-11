#!/bin/sh
# install.sh — download and install the spec CLI binary for the detected platform.
#
# Supported platforms: darwin/arm64, darwin/amd64, linux/arm64, linux/amd64
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/agent-fox-dev/spec-format/main/install.sh | sh
#
# Environment variables:
#   INSTALL_DIR  — override install directory (default: /usr/local/bin)
#   VERSION      — override version to install (default: latest)
#   BASE_URL     — override download base URL

set -eu

BINARY_NAME="spec"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-latest}"
BASE_URL="${BASE_URL:-https://github.com/agent-fox-dev/spec-format/releases/download}"

# Detect operating system.
detect_os() {
  os="$(uname -s)"
  case "$os" in
    Darwin) echo "darwin" ;;
    Linux)  echo "linux" ;;
    *)      echo "unsupported" ;;
  esac
}

# Detect architecture.
detect_arch() {
  arch="$(uname -m)"
  case "$arch" in
    arm64|aarch64) echo "arm64" ;;
    x86_64|amd64)  echo "amd64" ;;
    *)             echo "unsupported" ;;
  esac
}

# Main install logic.
main() {
  os="$(detect_os)"
  arch="$(detect_arch)"

  if [ "$os" = "unsupported" ] || [ "$arch" = "unsupported" ]; then
    echo "Error: unsupported platform: $(uname -s)/$(uname -m)" >&2
    echo "Supported platforms: darwin/arm64, darwin/amd64, linux/arm64, linux/amd64" >&2
    exit 1
  fi

  platform="${os}/${arch}"
  binary_suffix="${os}-${arch}"
  download_url="${BASE_URL}/v${VERSION}/${BINARY_NAME}-${binary_suffix}"

  echo "Installing spec CLI for ${platform}..."

  # Create install directory if needed.
  if [ ! -d "$INSTALL_DIR" ]; then
    mkdir -p "$INSTALL_DIR" 2>/dev/null || {
      echo "Error: cannot create install directory: ${INSTALL_DIR}" >&2
      echo "Try running with sudo or set INSTALL_DIR to a writable path." >&2
      exit 1
    }
  fi

  target="${INSTALL_DIR}/${BINARY_NAME}"
  tmp_target="${target}.tmp.$$"

  # Download binary to a temporary file to avoid leaving partial files on failure.
  if command -v curl >/dev/null 2>&1; then
    if ! curl -fSL --progress-bar -o "$tmp_target" "$download_url" 2>&1; then
      rm -f "$tmp_target"
      echo "Error: failed to download ${BINARY_NAME} from ${download_url}" >&2
      exit 1
    fi
  elif command -v wget >/dev/null 2>&1; then
    if ! wget -q -O "$tmp_target" "$download_url" 2>&1; then
      rm -f "$tmp_target"
      echo "Error: failed to download ${BINARY_NAME} from ${download_url}" >&2
      exit 1
    fi
  else
    echo "Error: curl or wget is required to download the binary" >&2
    exit 1
  fi

  # Move temporary file to final location and make executable.
  chmod +x "$tmp_target"
  mv "$tmp_target" "$target"

  # Verify installation and print version.
  if [ -x "$target" ]; then
    echo "Successfully installed ${BINARY_NAME} to ${target}"
    "$target" --version 2>/dev/null || true
  else
    echo "Error: installation failed — binary is not executable" >&2
    exit 1
  fi
}

main
