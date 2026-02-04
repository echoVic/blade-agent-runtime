#!/bin/bash
set -e

CHANGELOG_FILE="CHANGELOG.md"
VERSION_FILE="cmd/bar/update.go"

check_clean_working_tree() {
    if ! git diff --quiet HEAD; then
        echo "Error: Working tree has uncommitted changes"
        echo "Please commit or stash your changes before releasing"
        git status --short
        exit 1
    fi
}

show_help() {
    echo "Usage: ./scripts/release.sh <command> [options]"
    echo ""
    echo "Commands:"
    echo "  patch                Bump patch version (0.0.1 -> 0.0.2)"
    echo "  minor                Bump minor version (0.0.1 -> 0.1.0)"
    echo "  major                Bump major version (0.0.1 -> 1.0.0)"
    echo "  prepare <version>    Prepare release with specific version"
    echo "  publish              Build and publish the release"
    echo "  full <version>       Do both prepare and publish"
    echo ""
    echo "Examples:"
    echo "  ./scripts/release.sh patch"
    echo "  ./scripts/release.sh minor"
    echo "  ./scripts/release.sh major"
    echo "  ./scripts/release.sh prepare 0.0.15"
    echo "  ./scripts/release.sh publish"
    echo "  ./scripts/release.sh full 0.0.15"
}

get_current_version() {
    grep 'currentVersion = ' "$VERSION_FILE" | sed 's/.*"\(.*\)".*/\1/'
}

bump_version() {
    local current="$1"
    local type="$2"
    
    local major minor patch
    IFS='.' read -r major minor patch <<< "$current"
    
    case "$type" in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
    esac
    
    echo "$major.$minor.$patch"
}

update_version() {
    local version="$1"
    echo "==> Updating version to $version"
    sed -i '' "s/currentVersion = \".*\"/currentVersion = \"$version\"/" "$VERSION_FILE"
}

generate_changelog_content() {
    local last_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    local range=""
    
    if [ -n "$last_tag" ]; then
        range="$last_tag..HEAD"
    else
        range="HEAD"
    fi
    
    local added="" changed="" fixed="" removed=""
    
    while IFS= read -r line; do
        local type=$(echo "$line" | sed -n 's/^\([a-z]*\):.*/\1/p')
        local msg=$(echo "$line" | sed 's/^[a-z]*: //')
        
        case "$type" in
            feat)
                added="$added\n- $msg"
                ;;
            fix)
                fixed="$fixed\n- $msg"
                ;;
            refactor|perf|style)
                changed="$changed\n- $msg"
                ;;
            *)
                if [[ "$line" != chore:* && "$line" != docs:* && "$line" != test:* ]]; then
                    changed="$changed\n- $line"
                fi
                ;;
        esac
    done < <(git log --pretty=format:"%s" $range 2>/dev/null)
    
    local content=""
    if [ -n "$added" ]; then
        content="$content\n### Added$added\n"
    fi
    if [ -n "$changed" ]; then
        content="$content\n### Changed$changed\n"
    fi
    if [ -n "$fixed" ]; then
        content="$content\n### Fixed$fixed\n"
    fi
    if [ -n "$removed" ]; then
        content="$content\n### Removed$removed\n"
    fi
    
    echo -e "$content"
}

update_changelog() {
    local version="$1"
    local date=$(date +%Y-%m-%d)
    
    if grep -q "## \[$version\]" "$CHANGELOG_FILE"; then
        echo "==> Version $version already exists in CHANGELOG"
        return 0
    fi
    
    echo "==> Generating CHANGELOG for $version"
    
    local changelog_content=$(generate_changelog_content)
    
    if [ -z "$changelog_content" ] || [ "$changelog_content" = $'\n' ]; then
        echo "Warning: No commits found for changelog, using placeholder"
        changelog_content="\n### Changed\n- Version bump\n"
    fi
    
    local temp_file=$(mktemp)
    local header_done=0
    
    while IFS= read -r line; do
        echo "$line" >> "$temp_file"
        if [[ "$line" == "## [Unreleased]" ]]; then
            echo "" >> "$temp_file"
            echo "## [$version] - $date" >> "$temp_file"
            echo -e "$changelog_content" >> "$temp_file"
            header_done=1
        fi
    done < "$CHANGELOG_FILE"
    
    if [ "$header_done" -eq 0 ]; then
        echo "Warning: No [Unreleased] section found, prepending to file"
        local new_temp=$(mktemp)
        head -n 6 "$CHANGELOG_FILE" > "$new_temp"
        echo "" >> "$new_temp"
        echo "## [Unreleased]" >> "$new_temp"
        echo "" >> "$new_temp"
        echo "## [$version] - $date" >> "$new_temp"
        echo -e "$changelog_content" >> "$new_temp"
        tail -n +7 "$CHANGELOG_FILE" >> "$new_temp"
        mv "$new_temp" "$CHANGELOG_FILE"
        rm -f "$temp_file"
    else
        mv "$temp_file" "$CHANGELOG_FILE"
    fi
}

prepare_release() {
    local version="$1"
    
    if [ -z "$version" ]; then
        echo "Error: Version required"
        echo "Usage: ./scripts/release.sh prepare <version>"
        exit 1
    fi
    
    check_clean_working_tree
    
    local current=$(get_current_version)
    echo "Current version: $current"
    echo "New version: $version"
    echo ""
    
    update_version "$version"
    update_changelog "$version"
    
    echo ""
    echo "==> Committing changes..."
    git add "$VERSION_FILE" "$CHANGELOG_FILE"
    git commit -m "chore: release v$version"
    
    echo ""
    echo "✓ Release $version prepared"
    echo "  Run './scripts/release.sh publish' to build and publish"
}

publish_release() {
    check_clean_working_tree
    
    local version=$(get_current_version)
    local tag="v$version"
    
    echo "==> Publishing version $version"
    
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
    
    echo "==> Creating tag $tag..."
    git tag "$tag"
    
    echo "==> Pushing to origin..."
    git push
    git push origin "$tag"
    
    echo "==> Creating GitHub release..."
    gh release create "$tag" \
        dist/bar_darwin_amd64.tar.gz \
        dist/bar_darwin_arm64.tar.gz \
        dist/bar_linux_amd64.tar.gz \
        dist/bar_linux_arm64.tar.gz \
        --title "$tag" \
        --generate-notes
    
    echo ""
    echo "✓ Released $tag"
    echo "  https://github.com/echoVic/blade-agent-runtime/releases/tag/$tag"
}

full_release() {
    local version="$1"
    prepare_release "$version"
    publish_release
}

case "${1:-}" in
    patch|minor|major)
        current=$(get_current_version)
        new_version=$(bump_version "$current" "$1")
        full_release "$new_version"
        ;;
    prepare)
        prepare_release "$2"
        ;;
    publish)
        publish_release
        ;;
    full)
        full_release "$2"
        ;;
    -h|--help|help)
        show_help
        ;;
    *)
        show_help
        exit 1
        ;;
esac
