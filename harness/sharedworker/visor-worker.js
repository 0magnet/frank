// visor-worker.js — a SharedWorker standing in for the persistent wasm-visor.
//
// In the real system, bootOnce() would do:
//     importScripts('wasm_exec.js');
//     const go = new Go();
//     const { instance } = await WebAssembly.instantiateStreaming(fetch('wasm-visor.wasm'), go.importObject);
//     go.run(instance);                 // installs globalThis.skywireVisor (works in a worker)
//     await self.skywireVisor.boot(...);// dmsg client + transport.Manager + edge router come up
// ...and the worker would proxy skywireVisor.fetchDmsg / hvApi / serveContent to
// connected ports. Here it just keeps state so we can SEE that the visor boots
// exactly once and survives tab open/close/reload as long as the anchor port stays.

const bootMs = Date.now();
let booted = false;
let ticks = 0;
const ports = new Set();

function bootOnce() {
  if (booted) return;
  booted = true;
  console.log('[visor-worker] BOOTED — this must happen exactly ONCE');
  setInterval(() => {
    ticks++;
    const msg = { type: 'status', uptimeMs: Date.now() - bootMs, ticks, ports: ports.size };
    for (const p of ports) {
      try { p.postMessage(msg); } catch (e) { ports.delete(p); }
    }
  }, 1000);
}

// SharedWorker: every same-origin page that does `new SharedWorker(...)` triggers
// onconnect against the SAME worker instance. The worker lives while >=1 page
// holds it; the hidden anchor is what guarantees "browser lifetime".
onconnect = (e) => {
  const port = e.ports[0];
  ports.add(port);
  bootOnce();
  console.log('[visor-worker] port connected; live ports =', ports.size);
  port.onmessage = (m) => {
    if (m.data === 'status') {
      port.postMessage({ type: 'status', uptimeMs: Date.now() - bootMs, ticks, ports: ports.size });
    }
  };
  port.start();
};
