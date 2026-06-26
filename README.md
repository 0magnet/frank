# Frank

Frank is a native web browser written in Go that renders through an embedded
**WebKitGTK-6.0 (GTK4)** engine, so it displays modern JS / WebAssembly / WebGL /
CSS at full fidelity.

It is a fork of [thdwb](https://github.com/danfragoso/thdwb) (the "hotdog"
pure-Go browser). The original pure-Go rendering engine is preserved on the
**`pure-go`** branch; this `master` branch is the WebKit-GTK4 line. The pure-Go
engine packages (`mustard`/`bun`/`mayo`/`ketchup`/`gg`) remain in the module but
are no longer wired to the entry point.

## Status

A minimal but working browser: a GTK4 window with an address bar, back / forward
/ reload, and the WebKit engine. It is intended as the foundation for **Skywire
skynet integration** — browsing dmsg-addressed content over the Skywire network —
which is the next direction of work.

## Requirements

System libraries (resolved via `pkg-config`):

- `gtk4`
- `webkitgtk-6.0`

On Arch Linux: `sudo pacman -S gtk4 webkitgtk-6.0`

## Build & run

```sh
go build -o frank .            # build the browser binary
./frank                        # run (opens a welcome page)
./frank https://example.com    # open a URL
make run                       # or via make
```

`go build .` is self-contained — the entry point depends only on the
`gtk4` + `webkitgtk-6.0` system libraries (CGO).

### Old-GPU note

On GPUs whose driver cannot present WebKit's dmabuf zero-copy compositing path
(e.g. Intel Gen7 / GLES 3.0), WebGL renders one frame then freezes. Set
`FRANK_GL_COMPAT=1` to force the stable (slower, freeze-free) readback path:

```sh
FRANK_GL_COMPAT=1 ./frank
```

## Layout

- `main.go` — entry point + GPU compatibility profile.
- `browser.{h,c}` — GtkApplication and main window.
- `webview.c` — the WebKitGTK WebView, settings, and navigation.
- `chrome.c` — toolbar: back / forward / reload + address bar.
- `bun` `mayo` `ketchup` `mustard` `gg` `hotdog` `sauce` `pages` `assets`
  `profiler` — the retained pure-Go engine (not wired to this entry point).

## Branches

- `master` — WebKit-GTK4 browser (this).
- `pure-go` — the original pure-Go (GLFW/OpenGL) renderer, preserved.
