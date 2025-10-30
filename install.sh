#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Installing gitea-sync...${NC}"

# Check if we should install system-wide or user-wide
if [ "$EUID" -eq 0 ] || [ -w /usr/local/bin ]; then
    # System-wide installation
    BIN_DIR="/usr/local/bin"
    MAN_DIR="/usr/local/share/man/man1"
    echo -e "${YELLOW}Installing system-wide to $BIN_DIR${NC}"
else
    # User installation
    BIN_DIR="$HOME/.local/bin"
    MAN_DIR="$HOME/.local/share/man/man1"
    echo -e "${YELLOW}Installing for current user to $BIN_DIR${NC}"
fi

# Build the binary
echo "Building gitea-sync..."
go build -o gitea-sync .

# Create directories if they don't exist
mkdir -p "$BIN_DIR"
mkdir -p "$MAN_DIR"

# Install binary
echo "Installing binary to $BIN_DIR..."
cp gitea-sync "$BIN_DIR/"
chmod +x "$BIN_DIR/gitea-sync"

# Install man page
if [ -f "gitea-sync.1" ]; then
    echo "Installing man page to $MAN_DIR..."
    cp gitea-sync.1 "$MAN_DIR/"
    gzip -f "$MAN_DIR/gitea-sync.1"

    # Update man database if possible
    if command -v mandb &> /dev/null; then
        if [ "$EUID" -eq 0 ]; then
            mandb -q 2>/dev/null || true
        else
            mandb -u -q 2>/dev/null || true
        fi
    fi
fi

# Check if BIN_DIR is in PATH
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    echo -e "${YELLOW}Warning: $BIN_DIR is not in your PATH${NC}"
    echo "Add the following line to your ~/.bashrc or ~/.zshrc:"
    echo -e "${GREEN}export PATH=\"$BIN_DIR:\$PATH\"${NC}"
    echo ""
    echo "Then run: source ~/.bashrc  (or ~/.zshrc)"
fi

echo -e "${GREEN}Installation complete!${NC}"
echo ""
echo "You can now run: gitea-sync"
echo "For help, run: gitea-sync --help"
echo "View manual with: man gitea-sync"
