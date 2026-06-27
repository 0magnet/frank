package main

// Run the real wasm-visor under the hood (experimental).
//
// Frank starts `skywire cli hv serve` as a managed subprocess on a fresh
// ephemeral loopback port for the browser's lifetime; the always-present anchor
// WebView then loads that URL, which boots the actual wasm-visor (dmsg client +
// transport.Manager + edge router) in a persistent page. Each browser that loads
// hv serve mints its own keyless ephemeral visor, so this never clashes with a
// visor the other agent may already be serving.
//
// This reuses hv serve (which handles asset + config + boot wiring) as the
// least-coupled way to get the visor running now. A later, self-contained step
// can serve the wasm + a minimal boot page via the frank:// scheme instead,
// dropping the subprocess and the loopback port. Set FRANK_NO_VISOR=1 to skip
// and fall back to the SharedWorker stub.

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// startVisorServer launches `skywire cli hv serve` on an ephemeral loopback port
// and waits for it, returning the URL the anchor loads and a cleanup func.
func startVisorServer() (string, func(), error) {
	bin := locateSkywire()
	if bin == "" {
		return "", nil, fmt.Errorf("skywire binary not found (set SKYWIRE_BIN, or put it on PATH)")
	}
	port, err := freePort()
	if err != nil {
		return "", nil, fmt.Errorf("pick ephemeral port: %w", err)
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	cmd := exec.Command(bin, "cli", "hv", "serve", "--addr", addr)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGTERM}
	if err := cmd.Start(); err != nil {
		return "", nil, fmt.Errorf("start hv serve: %w", err)
	}
	cleanup := func() {
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
			_ = cmd.Process.Kill()
		}
	}

	url := "http://" + addr + "/"
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for {
		if ctx.Err() != nil {
			cleanup()
			return "", nil, fmt.Errorf("hv serve did not come up at %s within timeout", addr)
		}
		if c, derr := net.DialTimeout("tcp", addr, 500*time.Millisecond); derr == nil {
			c.Close()
			if resp, herr := http.Get(url); herr == nil {
				resp.Body.Close()
				fmt.Fprintf(os.Stderr, "frank: wasm-visor server up at %s (anchor will boot it)\n", url)
				return url, cleanup, nil
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
}
