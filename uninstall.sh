#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Uninstalling gitea-sync...${NC}"

# Check system-wide location first
if [ -f "/usr/local/bin/gitea-sync" ]; then
    BIN_DIR="/usr/local/bin"
    MAN_DIR="/usr/local/share/man/man1"
    SYSTEM_WIDE=true
elif [ -f "$HOME/.local/bin/gitea-sync" ]; then
    BIN_DIR="$HOME/.local/bin"
    MAN_DIR="$HOME/.local/share/man/man1"
    SYSTEM_WIDE=false
else
    echo -e "${RED}gitea-sync not found in standard locations${NC}"
    exit 1
fi

# Remove binary
if [ -f "$BIN_DIR/gitea-sync" ]; then
    echo "Removing binary from $BIN_DIR..."
    rm -f "$BIN_DIR/gitea-sync"
fi

# Remove man page
if [ -f "$MAN_DIR/gitea-sync.1.gz" ]; then
    echo "Removing man page from $MAN_DIR..."
    rm -f "$MAN_DIR/gitea-sync.1.gz"

    # Update man database if possible
    if command -v mandb &> /dev/null; then
        if [ "$SYSTEM_WIDE" = true ]; then
            if [ "$EUID" -eq 0 ]; then
                mandb -q 2>/dev/null || true
            fi
        else
            mandb -u -q 2>/dev/null || true
        fi
    fi
fi

echo -e "${GREEN}Uninstallation complete!${NC}"
