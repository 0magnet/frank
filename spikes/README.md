# Engine spikes

Standalone experiments that led to Frank's engine choice. Each is its own Go
module, so they are not built by the parent `./...` and can keep their own
system-library requirements. They are kept for the findings, not as live code.

The question all three answer: **which embedded browser engine renders
JS + WebAssembly + WebGL correctly in a native Go program?** That is the gating
requirement for Frank.

| Spike | Engine | What it established |
|---|---|---|
| `webview/` | WebKitGTK (WebKit2, GTK3) | A CGO-embedded OS engine *can* render full JS + WASM + WebGL. Probes the live page after load for canvas/WebGL/WebAssembly presence and the GL `RENDERER` string, and writes a PNG snapshot to `/tmp` so a run can be verified without watching the screen. |
| `webkit6/` | WebKitGTK-6.0 (GTK4) | The GTK4 twin of the above. Same WebKit core (2.52.4), but GTK4 composites the WebView through `GskRenderer` + GTK4 dmabuf textures instead of the GTK3 path. The WebGL problem — **dmabuf zero-copy freeze / readback at half speed** — lives in that compositing layer, so this tests whether the GTK4 path presents the fast zero-copy route without freezing on old Intel. Runs with dmabuf **enabled** (the whole point); `SPIKE_NO_DMABUF=1` disables it for an A/B against the GTK3 result. |
| `cef/` | Chromium via CEF (C API, Views) | The alternative to WebKitGTK. Prints fps via an injected rAF counter and a console-message handler, uses a unique cache dir per launch so Chromium's single-instance lock does not stop the window opening, and passes `--no-zygote` / `--no-sandbox` so CEF's process model does not fight the Go runtime. |

Each takes an optional URL argument; `webview` defaults to the wasm-stuff WebGL
test page.

Frank's `master` is the WebKitGTK-6.0 / GTK4 line, which is the path
`webkit6/` was written to validate.
