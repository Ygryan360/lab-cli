#!/usr/bin/env bash
# install.sh — Build and install the lab CLI tool
set -e

BINARY_NAME="lab"
INSTALL_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.config/lab"

# Colors
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
RESET='\033[0m'

info()    { echo -e "${CYAN}→ $*${RESET}"; }
success() { echo -e "${GREEN}✓ $*${RESET}"; }
warn()    { echo -e "${YELLOW}⚠ $*${RESET}"; }
die()     { echo -e "${RED}✗ $*${RESET}"; exit 1; }

echo ""
echo -e "${CYAN}${RESET}  Installing lab CLI..."
echo ""

# Check Go is installed
if ! command -v go &>/dev/null; then
  die "Go is not installed. Install it from https://go.dev/dl/"
fi

GO_VERSION=$(go version | awk '{print $3}')
info "Go found: $GO_VERSION"

# Build
info "Building binary..."
go build -ldflags="-s -w" -o "$BINARY_NAME" ./cmd/lab/
success "Binary built ($(du -h $BINARY_NAME | cut -f1))"

# Install
mkdir -p "$INSTALL_DIR"
mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"
success "Installed to $INSTALL_DIR/$BINARY_NAME"

# Ensure ~/.local/bin is in PATH
SHELL_RC=""
CURRENT_SHELL=$(basename "$SHELL")
case "$CURRENT_SHELL" in
  bash) SHELL_RC="$HOME/.bashrc" ;;
  zsh)  SHELL_RC="$HOME/.zshrc" ;;
  fish) SHELL_RC="$HOME/.config/fish/config.fish" ;;
  *)    SHELL_RC="$HOME/.profile" ;;
esac

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  warn "$INSTALL_DIR is not in your PATH"
  echo ""
  echo "  Add this to your $SHELL_RC:"
  echo ""
  if [[ "$CURRENT_SHELL" == "fish" ]]; then
    echo -e "  ${YELLOW}fish_add_path \$HOME/.local/bin${RESET}"
  else
    echo -e "  ${YELLOW}export PATH=\"\$HOME/.local/bin:\$PATH\"${RESET}"
  fi
  echo ""
  echo "  Then run: source $SHELL_RC"
  echo ""
else
  success "PATH is already configured"
fi

# Initialize config if needed
if [[ ! -f "$CONFIG_DIR/config.json" ]]; then
  mkdir -p "$CONFIG_DIR"
  LAB_PATH="$HOME/lab"
  cat > "$CONFIG_DIR/config.json" << CONF
{
  "lab_path": "$LAB_PATH",
  "default_profile": "Default",
  "profiles": [
    {"name": "Default"}
  ],
  "category_defaults": []
}
CONF
  success "Config initialized at $CONFIG_DIR/config.json"
  info "Lab path set to: $LAB_PATH"
  warn "Run 'lab config set-path <your-path>' if your lab is elsewhere"
else
  info "Config already exists at $CONFIG_DIR/config.json (not overwritten)"
fi

echo ""
echo -e "${GREEN}✓ Installation complete!${RESET}"
echo ""
echo "  Quick start:"
echo "    lab                   → open TUI"
echo "    lab help              → see all commands"
echo "    lab profiles add 'Laravel/PHP'"
echo "    lab profiles set-category sites 'Laravel/PHP'"
echo ""
