#!/usr/bin/env bash
set -euo pipefail

REVEALJS_VER="5.1.0"

VENDOR_DIR="web/static"
CHECKSUM_FILE="$VENDOR_DIR/CHECKSUMS.sha256"

download() {
    local url="$1" dest="$2"
    echo "Downloading $url -> $dest"
    curl -fsSL -o "$dest" "$url"
}

mkdir -p "$VENDOR_DIR/reveal"
mkdir -p "$VENDOR_DIR/fonts"

download "https://cdn.jsdelivr.net/npm/reveal.js@${REVEALJS_VER}/dist/reveal.js" \
         "$VENDOR_DIR/reveal/reveal.js"
download "https://cdn.jsdelivr.net/npm/reveal.js@${REVEALJS_VER}/dist/reveal.css" \
         "$VENDOR_DIR/reveal/reveal.css"

download "https://cdn.jsdelivr.net/fontsource/fonts/noto-sans-tc@latest/chinese-traditional-400-normal.woff2" \
         "$VENDOR_DIR/fonts/NotoSansTC-Regular.woff2"
download "https://cdn.jsdelivr.net/fontsource/fonts/noto-sans-tc@latest/chinese-traditional-700-normal.woff2" \
         "$VENDOR_DIR/fonts/NotoSansTC-Bold.woff2"

download "https://cdn.jsdelivr.net/fontsource/fonts/jetbrains-mono@latest/latin-400-normal.woff2" \
         "$VENDOR_DIR/fonts/JetBrainsMono-Regular.woff2"

if [ "${1:-}" = "--update-checksums" ]; then
    cd "$VENDOR_DIR"
    find . -type f ! -name CHECKSUMS.sha256 ! -name VERSIONS.md | sort | xargs sha256sum > CHECKSUMS.sha256
    echo "Checksums updated in $CHECKSUM_FILE"
else
    echo "Verifying checksums..."
    cd "$VENDOR_DIR"
    sha256sum -c CHECKSUMS.sha256
fi

echo "Done."
