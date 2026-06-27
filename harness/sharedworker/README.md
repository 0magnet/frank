# SharedWorker persistence harness

Proves the lifecycle pattern for running the wasm-visor **persistently under the
hood**, decoupled from any tab — so opening/closing/reloading the hypervisor or
browse tabs does **not** restart the visor (today, reloading the tab restarts it).

## The pattern

```
            ┌─────────────────────────────────────────┐
            │  SharedWorker  "frank-visor"             │   ← the wasm-visor
            │  (boots ONCE, holds dmsg/WS sessions)    │      runs here
            └───▲───────────────▲──────────────▲───────┘
                │               │              │
        anchor.html        tab (hv UI)    tab (browse)
        (hidden, never      (open/close/   (open/close/
         closed → keeps      reload freely) reload freely)
         worker alive)
```

- **`visor-worker.js`** — the visor, in a `SharedWorker`. Boots exactly once;
  keeps state (here just uptime/ticks; really: the dmsg client + transport
  manager + router). It lives as long as ≥1 page holds a connection.
- **`anchor.html`** — Frank loads this **hidden, at startup, and never reloads
  it** (an off-screen WebView). Holding one connection open keeps the worker —
  and thus the visor — alive for the browser's whole lifetime.
- **`tab.html`** — any visible tab (hypervisor UI, browse). Connects to the same
  worker; open/close/reload it freely, the visor keeps running.

## Run the demo

SharedWorkers are shared by **same-origin** pages — `file://` pages each get a
*unique* origin and won't share, so serve over http:

```sh
cd harness/sharedworker && python3 -m http.server 8088
# open http://localhost:8088/anchor.html  AND  http://localhost:8088/tab.html
```

Reload `tab.html`: uptime keeps climbing (visor persists). Close `anchor.html`,
then close + reopen the last tab: uptime resets (no anchor → worker terminated) —
demonstrating *why* the anchor is required.

## Mapping to the real wasm-visor

- Replace `bootOnce()` with `importScripts('wasm_exec.js')` + instantiate
  `wasm-visor.wasm` + `skywireVisor.boot(...)`. The wasm-visor's API is installed
  on the global (`js.Global().Set(...)`), which works in a worker scope — so the
  core is already worker-portable. The boot/loader (currently DOM-based
  `hv-boot.js`) needs a worker variant. **This is the other agent's change.**
- Expose `skywireVisor.fetchDmsg / hvApi / serveContent` over the worker port
  (`postMessage`) instead of `globalThis`, so tabs and Frank's native address-bar
  bridge call the one persistent visor.
- **Frank side (this repo):** create the hidden anchor WebView at startup, serve
  these assets from a single origin (custom scheme or localhost), and route the
  address bar through the persistent visor. The anchor + same-origin serving is
  what this harness establishes.
