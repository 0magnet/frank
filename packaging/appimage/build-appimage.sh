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

# 3. fetch linuxdeploy + the GTK plugin (cached in tools/)
get() { [ -f "$tools/$1" ] || { echo "fetching $1"; curl -fL "$2" -o "$tools/$1"; chmod +x "$tools/$1"; }; }
arch="$(uname -m)"
get linuxdeploy "https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-${arch}.AppImage"
get linuxdeploy-plugin-gtk "https://raw.githubusercontent.com/linuxdeploy/linuxdeploy-plugin-gtk/master/linuxdeploy-plugin-gtk.sh"

# 4. bundle deps with the GTK plugin (gathers gtk4 + webkit shared libs)
( cd "$work" && "$tools/linuxdeploy" --appdir "$appdir" \
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
get appimagetool "https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-${arch}.AppImage"
ARCH="$arch" "$tools/appimagetool" "$appdir" "$here/Frank-${arch}.AppImage"
echo "built: $here/Frank-${arch}.AppImage"
