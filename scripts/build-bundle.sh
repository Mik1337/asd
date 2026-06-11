#!/usr/bin/env bash
# Build cross-platform release archives + checksums into ./dist.
# Usage: scripts/build-bundle.sh [version]
set -euo pipefail

cd "$(dirname "$0")/.."

VERSION="${1:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
BINARY="asd"
DIST="dist"
PLATFORMS="darwin/amd64 darwin/arm64 linux/amd64 linux/arm64"

rm -rf "$DIST"
mkdir -p "$DIST"

echo "Building asd $VERSION"
for p in $PLATFORMS; do
	os="${p%/*}"
	arch="${p#*/}"
	stage="$DIST/${BINARY}_${os}_${arch}"
	mkdir -p "$stage"

	echo "  -> $os/$arch"
	CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
		go build -ldflags "-s -w -X main.version=$VERSION" -o "$stage/$BINARY" .

	for extra in README.md DESIGN.md LICENSE; do
		[ -f "$extra" ] && cp "$extra" "$stage"/
	done

	tar -czf "$DIST/${BINARY}_${os}_${arch}.tar.gz" -C "$stage" .
	rm -rf "$stage"
done

( cd "$DIST" && shasum -a 256 ./*.tar.gz > checksums.txt )

echo "---"
ls -lh "$DIST"
echo "Bundle ready in ./$DIST (version $VERSION)"
