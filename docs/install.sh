#!/bin/sh
set -e

REPO="echoVic/blade-agent-runtime"
BINARY="bar"
INSTALL_DIR="${BAR_INSTALL_DIR:-$HOME/.local/bin}"

get_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        *)       echo "unsupported" ;;
    esac
}

get_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *)            echo "unsupported" ;;
    esac
}

main() {
    OS=$(get_os)
    ARCH=$(get_arch)

    if [ "$OS" = "unsupported" ] || [ "$ARCH" = "unsupported" ]; then
        echo "Error: Unsupported OS or architecture"
        echo "Please install manually: go install github.com/$REPO/cmd/bar@latest"
        exit 1
    fi

    VERSION="${BAR_VERSION:-latest}"
    if [ "$VERSION" = "latest" ]; then
        VERSION=$(curl -sL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    fi

    if [ -z "$VERSION" ]; then
        echo "Error: Could not determine version"
        exit 1
    fi

    FILENAME="${BINARY}_${OS}_${ARCH}.tar.gz"
    URL="https://github.com/$REPO/releases/download/$VERSION/$FILENAME"

    echo "Installing $BINARY $VERSION ($OS/$ARCH)..."

    TMPDIR=$(mktemp -d)
    trap "rm -rf $TMPDIR" EXIT

    echo "Downloading $URL..."
    if ! curl -sL "$URL" -o "$TMPDIR/$FILENAME"; then
        echo "Error: Download failed"
        echo "Please install manually: go install github.com/$REPO/cmd/bar@latest"
        exit 1
    fi

    echo "Extracting..."
    tar -xzf "$TMPDIR/$FILENAME" -C "$TMPDIR"

    echo "Installing to $INSTALL_DIR..."
    mkdir -p "$INSTALL_DIR"
    mv "$TMPDIR/$BINARY" "$INSTALL_DIR/$BINARY"
    chmod +x "$INSTALL_DIR/$BINARY"

    if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
        echo ""
        echo "Add this to your shell profile:"
        echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    fi

    echo ""
    echo "✓ $BINARY $VERSION installed successfully!"
    echo "  Run 'bar --help' to get started"
}

main
