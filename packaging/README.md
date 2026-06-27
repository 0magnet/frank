# Packaging Frank

Frank links the system **WebKitGTK-6.0** + **GTK4** dynamically, so the bare
binary is ~2.4 MB but needs those libraries present. For self-contained
distribution (no system deps), two routes — **Flatpak is recommended**.

## Why this is sandbox-safe (the skynet angle)

The self-contained build embeds the **wasm-visor**, not a host-native visor. The
wasm-visor is dial-only and uses **WebSocket/WebRTC** transports — browser APIs.
So it needs only `--share=network`; it never wants raw sockets, listening ports,
or broad filesystem access. That is exactly why Frank can ship sandboxed
(Flatpak) where a host-native visor could not, and why packaging Frank does
**not** drag skywire-the-daemon into a sandbox — Frank is self-sufficient.

## Flatpak (recommended)

The GNOME runtime ships GTK4 **and** WebKitGTK 6.0, correctly wired
(multi-process helpers, etc.), so we don't bundle or hand-fix the engine.

```sh
flatpak install -y flathub org.gnome.Platform//47 org.gnome.Sdk//47 \
  org.freedesktop.Sdk.Extension.golang//24.08
flatpak-builder --user --install --force-clean build-dir \
  packaging/flatpak/io.github._0magnet.frank.yml
flatpak run io.github._0magnet.frank
```

Notes:
- `go build .` (the browser entry) imports no external Go modules, so the
  sandboxed offline build works as-is. If a future entry adds Go deps, run
  `go mod vendor`, commit `vendor/`, and the offline build still works.
- Add a real 256×256 icon `io.github._0magnet.frank.png` and reference it.
- Bump `runtime-version` as new GNOME runtimes ship newer WebKit.

## AppImage (secondary)

```sh
packaging/appimage/build-appimage.sh   # produces Frank-x86_64.AppImage
```

Caveat: WebKitGTK is multi-process; its helper executables
(`WebKitNetworkProcess`, `WebKitWebProcess`, …) and injected-bundle libs must be
bundled and found at runtime. The script copies them and sets `WEBKIT_EXEC_PATH`,
but this is the fragile part of AppImage-ing any WebKit app — **test WebGL and
network after building**. Flatpak avoids this entirely.

## Size comparison

| Route | Size | Self-contained | System dep |
|---|---|---|---|
| bare binary | ~2.4 MB | no | gtk4 + webkitgtk-6.0 |
| **Flatpak** | ~50–80 MB app (GNOME runtime shared) | yes | none |
| AppImage | ~80–120 MB | yes | none (FUSE to mount) |
| (CEF alternative) | ~290 MB | yes | none |
