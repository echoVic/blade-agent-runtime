#!/bin/bash
set -e

VERSION="${1:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

echo "Installing Blade Agent Runtime (BAR)..."

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case "$OS" in
    darwin|linux)
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

echo "Detected: $OS/$ARCH"

# Download URL (placeholder - update when releases are available)
if [ "$VERSION" = "latest" ]; then
    DOWNLOAD_URL="https://github.com/user/blade-agent-runtime/releases/latest/download/bar-${OS}-${ARCH}"
else
    DOWNLOAD_URL="https://github.com/user/blade-agent-runtime/releases/download/${VERSION}/bar-${OS}-${ARCH}"
fi

echo "Downloading from: $DOWNLOAD_URL"

# For now, build from source if binary not available
if ! curl -fsSL "$DOWNLOAD_URL" -o /tmp/bar 2>/dev/null; then
    echo "Binary not available, building from source..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        echo "Go is not installed. Please install Go 1.21+ first."
        echo "  brew install go  # macOS"
        echo "  apt install golang  # Ubuntu/Debian"
        exit 1
    fi
    
    # Build from source
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    echo "Cloning repository..."
    git clone --depth 1 https://github.com/user/blade-agent-runtime.git
    cd blade-agent-runtime
    
    echo "Building..."
    go build -o /tmp/bar ./cmd/bar
    
    cd /
    rm -rf "$TEMP_DIR"
fi

# Install binary
echo "Installing to $INSTALL_DIR/bar..."
sudo mv /tmp/bar "$INSTALL_DIR/bar"
sudo chmod +x "$INSTALL_DIR/bar"

echo ""
echo "✅ BAR installed successfully!"
echo ""
echo "Get started:"
echo "  cd your-project"
echo "  bar init"
echo "  bar task start my-task"
echo "  bar run -- your-agent-command"
echo ""
echo "For more information: bar --help"
