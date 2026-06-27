#!/usr/bin/env bash
# Build an AppImage of Frank (WebKitGTK-6.0 / GTK4).
#
# NOTE: Flatpak (packaging/flatpak/) is the more reliable self-contained route
# for a WebKit app — the GNOME runtime ships a correctly-wired WebKitGTK. AppImage
# is provided as a secondary option and has one real gotcha: WebKitGTK is
# multi-process and looks for its helper executables (WebKitNetworkProcess,
# WebKitWebProcess) + injected-bundle libs under a libexec path. linuxdeploy's
# GTK plugin does not always bundle those, so this script copies them explicitly
# and sets WEBKIT_EXEC_PATH in the AppRun wrapper. Test WebGL/network after build.
#
# Requires: go, plus internet on first run to fetch linuxdeploy tools.
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
root="$(cd "$here/../.." && pwd)"
work="$here/.build"
appdir="$work/Frank.AppDir"
tools="$work/tools"
mkdir -p "$tools"

# 1. build the binary (off RAM-backed /tmp)
export TMPDIR="${TMPDIR:-$HOME/.cache/gotmp}"; mkdir -p "$TMPDIR"
( cd "$root" && go build -ldflags "-s -w" -o "$work/frank" . )

# 2. AppDir skeleton
rm -rf "$appdir"; mkdir -p "$appdir/usr/bin" "$appdir/usr/share/applications" \
  "$appdir/usr/share/icons/hicolor/256x256/apps"
install -Dm755 "$work/frank" "$appdir/usr/bin/frank"
install -Dm644 "$root/packaging/flatpak/io.github._0magnet.frank.desktop" \
  "$appdir/usr/share/applications/io.github._0magnet.frank.desktop"
# A placeholder icon (replace with a real 256x256 PNG named frank.png).
: > "$appdir/usr/share/icons/hicolor/256x256/apps/io.github._0magnet.frank.png"

# 3. resolve linuxdeploy + appimagetool — prefer a system install (Arch AUR:
#    `linuxdeploy-appimage`, `appimagetool-bin`), else download as a fallback.
#    The GTK plugin is NOT packaged on Arch, so it is always fetched and made
#    discoverable via PATH (linuxdeploy auto-finds linuxdeploy-plugin-<name>).
arch="$(uname -m)"
fetch() { [ -f "$tools/$1" ] || { echo "fetching $1 (not installed)" >&2; curl -fL "$2" -o "$tools/$1"; chmod +x "$tools/$1"; }; echo "$tools/$1"; }
resolve() { if command -v "$1" >/dev/null 2>&1; then command -v "$1"; else fetch "$1" "$2"; fi; }
LINUXDEPLOY="$(resolve linuxdeploy "https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-${arch}.AppImage")"
APPIMAGETOOL="$(resolve appimagetool "https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-${arch}.AppImage")"
fetch linuxdeploy-plugin-gtk "https://raw.githubusercontent.com/linuxdeploy/linuxdeploy-plugin-gtk/master/linuxdeploy-plugin-gtk.sh" >/dev/null
export PATH="$tools:$PATH"   # so linuxdeploy finds the fetched gtk plugin

# 4. bundle deps with the GTK plugin (gathers gtk4 + webkit shared libs)
( cd "$work" && "$LINUXDEPLOY" --appdir "$appdir" \
    --plugin gtk \
    -d "$appdir/usr/share/applications/io.github._0magnet.frank.desktop" )

# 5. WebKit multi-process helpers + injected bundle (the gotcha)
wk_libexec="$(pkg-config --variable=libexecdir webkitgtk-6.0 2>/dev/null || echo /usr/libexec)"
wk_libdir="$(pkg-config --variable=libdir webkitgtk-6.0 2>/dev/null || echo /usr/lib)"
mkdir -p "$appdir/usr/libexec"
for h in WebKitNetworkProcess WebKitWebProcess WebKitGPUProcess; do
  [ -f "$wk_libexec/$h" ] && cp -v "$wk_libexec/$h" "$appdir/usr/libexec/" || true
done
# injected bundle lives under .../webkitgtk-6.0/injected-bundle/
[ -d "$wk_libdir/webkitgtk-6.0" ] && cp -rv "$wk_libdir/webkitgtk-6.0" "$appdir/usr/lib/" || true

# point WebKit at the bundled helpers from AppRun
cat >> "$appdir/AppRun" <<'EOF'

# --- Frank/WebKit additions ---
export WEBKIT_EXEC_PATH="$APPDIR/usr/libexec"
EOF

# 6. package
ARCH="$arch" "$APPIMAGETOOL" "$appdir" "$here/Frank-${arch}.AppImage"
echo "built: $here/Frank-${arch}.AppImage"
