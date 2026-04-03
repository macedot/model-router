#!/bin/bash
set -e

SERVICE_NAME="model-router"
SERVICE_FILE="$HOME/.config/systemd/user/${SERVICE_NAME}.service"
BINARY_PATH="$(pwd)/model-router"

echo "Installing ${SERVICE_NAME} as user service..."

# Build the binary
echo "Building..."
make build

# Create systemd user service directory if it doesn't exist
mkdir -p "$HOME/.config/systemd/user"

# Create the service file
cat > "$SERVICE_FILE" << EOF
[Unit]
Description=Model Router - AI Model Proxy
After=network.target

[Service]
ExecStart=${BINARY_PATH}
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
EOF

echo "Created service file: ${SERVICE_FILE}"

# Reload systemd and enable the service
systemctl --user daemon-reload
systemctl --user enable "${SERVICE_NAME}.service"

echo ""
echo "Installation complete!"
echo ""
echo "Commands:"
echo "  systemctl --user start ${SERVICE_NAME}   # Start the service"
echo "  systemctl --user stop ${SERVICE_NAME}    # Stop the service"
echo "  systemctl --user status ${SERVICE_NAME}  # Check status"
echo "  journalctl --user -u ${SERVICE_NAME}     # View logs"
echo ""
echo "To uninstall:"
echo "  systemctl --user stop ${SERVICE_NAME}"
echo "  systemctl --user disable ${SERVICE_NAME}"
echo "  rm ${SERVICE_FILE}"
