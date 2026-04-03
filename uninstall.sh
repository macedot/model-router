#!/bin/bash
set -e

SERVICE_NAME="model-router"
BINARY_PATH="${HOME}/.local/bin/model-router"
SERVICE_FILE="$HOME/.config/systemd/user/${SERVICE_NAME}.service"

echo "Uninstalling model-router..."

# Stop and disable service if running
if systemctl --user list-unit-files | grep -q "^${SERVICE_NAME}.service"; then
    echo "Stopping service..."
    systemctl --user stop "${SERVICE_NAME}" 2>/dev/null || true
    echo "Disabling service..."
    systemctl --user disable "${SERVICE_NAME}" 2>/dev/null || true
    systemctl --user daemon-reload
    rm -f "$SERVICE_FILE"
    echo "Removed systemd service file."
fi

# Remove binary
if [ -f "$BINARY_PATH" ]; then
    rm -f "$BINARY_PATH"
    echo "Removed binary: ${BINARY_PATH}"
fi

echo ""
echo "Config files preserved:"
echo "  ~/.model-router/config.json"
echo "  ~/.env (if present)"
echo ""
echo "Uninstallation complete."
