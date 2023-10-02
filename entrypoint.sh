#!/bin/sh
set -e

# Create a non-root user
NON_ROOT_USER=appuser

if ! id "$NON_ROOT_USER" &>/dev/null; then
    echo "[entrypoint.sh] Creating non-root user..."
    adduser --disabled-password --gecos "" --uid 1000 "$NON_ROOT_USER"
fi

# Check for user-provided certificates and add them
if [ -d "$TRUSTED_CA_PATH" ]; then
    echo "[entrypoint.sh] Adding user-provided certificates..."
    for cert in $TRUSTED_CA_PATH/*; do
        cp "$cert" /usr/local/share/ca-certificates/
    done
    update-ca-certificates
fi

# Run user's initialization script if present
if [ -f "$INIT_SCRIPT_PATH" ]; then
    echo "[entrypoint.sh] Running initialization script..."
    "$INIT_SCRIPT_PATH"
fi

# Drop privileges by switching to the non-root user and run the main application
echo "[entrypoint.sh] Switching to non-root user and starting the application..."
exec su -s /bin/sh -c "$@" "$NON_ROOT_USER"