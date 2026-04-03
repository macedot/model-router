#!/bin/bash
set -e

REPO="macedot/model-router"
INSTALL_DIR="${HOME}/.local/bin"
BINARY_PATH="${INSTALL_DIR}/model-router"
TEMP_DIR=$(mktemp -d)

# Detect OS and architecture
detect() {
    local os arch
    case "$(uname -s)" in
        Linux)  os="linux" ;;
        Darwin) os="darwin" ;;
        *)      echo "Unsupported OS: $(uname -s)"; exit 1 ;;
    esac
    case "$(uname -m)" in
        x86_64)  arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *)       echo "Unsupported arch: $(uname -m)"; exit 1 ;;
    esac
    echo "$os-$arch"
}

echo "Installing model-router from GitHub releases..."

# Download latest release
PLATFORM=$(detect)
ASSET="model-router-${PLATFORM}.tar.gz"
TMP_ASSET="${TEMP_DIR}/${ASSET}"
TMP_EXTRACT="${TEMP_DIR}/model-router"

URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

echo "Downloading ${ASSET}..."
curl -sL "$URL" -o "$TMP_ASSET"

echo "Extracting to ${INSTALL_DIR}..."
mkdir -p "$INSTALL_DIR"
tar -xzf "$TMP_ASSET" -C "$TEMP_DIR"
mv "$TMP_EXTRACT" "$BINARY_PATH"
chmod +x "$BINARY_PATH"

# Clean up
rm -rf "$TEMP_DIR"

echo ""
echo "Installed: ${BINARY_PATH}"
echo ""
echo "Ensure ${INSTALL_DIR} is in your PATH."
echo ""
echo "Commands:"
echo "  model-router              # Run directly"
echo "  systemctl --user start model-router   # As systemd service"
echo ""
echo "To uninstall:"
echo "  curl -sL https://raw.githubusercontent.com/macedot/model-router/master/uninstall.sh | bash"
