#!/bin/sh

SCRIPT_DIR=$(dirname "$0")
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
HOOKS_DIR="$REPO_ROOT/.git/hooks"

mkdir -p "$HOOKS_DIR"

cat > "$HOOKS_DIR/post-commit" << 'EOF'
#!/bin/sh

VERSION_FILE="VERSION"

if [ ! -f "$VERSION_FILE" ]; then
    exit 0
fi

VERSION=$(cat "$VERSION_FILE" | tr -d '[:space:]')
TAG="v$VERSION"

if git rev-parse "$TAG" >/dev/null 2>&1; then
    exit 0
fi

git tag "$TAG"
git push origin "$TAG"

echo "Created and pushed tag $TAG"
EOF

chmod +x "$HOOKS_DIR/post-commit"

echo "Installed post-commit hook"
