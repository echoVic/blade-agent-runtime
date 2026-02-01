#!/bin/bash
set -e

if [ -z "$1" ]; then
    echo "Usage: ./scripts/release.sh <version>"
    echo "Example: ./scripts/release.sh 0.0.6"
    exit 1
fi

VERSION="$1"
TAG="v$VERSION"

echo "==> Updating version to $VERSION"
sed -i '' "s/currentVersion = \".*\"/currentVersion = \"$VERSION\"/" cmd/bar/update.go

echo "==> Building binaries..."
mkdir -p dist

for OS in darwin linux; do
    for ARCH in amd64 arm64; do
        echo "    Building $OS/$ARCH..."
        GOOS=$OS GOARCH=$ARCH go build -o dist/bar ./cmd/bar
        tar -czf "dist/bar_${OS}_${ARCH}.tar.gz" -C dist bar
        rm dist/bar
    done
done

echo "==> Committing version bump..."
git add -f cmd/bar/update.go
git commit -m "chore: bump version to $VERSION" || true

echo "==> Creating tag $TAG..."
git tag "$TAG"

echo "==> Pushing to origin..."
git push
git push origin "$TAG"

echo "==> Creating GitHub release..."
gh release create "$TAG" \
    dist/bar_darwin_amd64.tar.gz \
    dist/bar_darwin_arm64.tar.gz \
    dist/bar_linux_amd64.tar.gz \
    dist/bar_linux_arm64.tar.gz \
    --title "$TAG" \
    --generate-notes

echo ""
echo "✓ Released $TAG"
echo "  https://github.com/echoVic/blade-agent-runtime/releases/tag/$TAG"
