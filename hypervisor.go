package main

// Experimental Skywire integration (phase A): run the in-tab wasm-visor by
// spawning `skywire cli hv serve` and loading its local URL in the WebView. The
// wasm-visor boots inside the page and provides the hypervisor UI, the dmsg/
// skynet browse + host panels, and the in-tab skysocks-lite — all four target
// features, via the existing build. A later phase can embed the wasm+UI assets
// and serve them through Frank's in-process scheme handler to drop the
// subprocess and the local UI port.

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// locateSkywire finds the skywire binary: $SKYWIRE_BIN, then $PATH, then the
// local fork checkout.
func locateSkywire() string {
	if b := os.Getenv("SKYWIRE_BIN"); b != "" {
		return b
	}
	if p, err := exec.LookPath("skywire"); err == nil {
		return p
	}
	cand := os.ExpandEnv("$HOME/go/src/github.com/0pcom/skywire/skywire")
	if _, err := os.Stat(cand); err == nil {
		return cand
	}
	return ""
}

// startHypervisor launches `skywire cli hv serve` and waits for it to accept
// HTTP, returning the URL to load and a cleanup func that stops the subprocess.
func startHypervisor() (string, func(), error) {
	bin := locateSkywire()
	if bin == "" {
		return "", nil, fmt.Errorf("skywire binary not found (set SKYWIRE_BIN, or put it on PATH)")
	}

	addr := os.Getenv("FRANK_HV_ADDR")
	if addr == "" {
		addr = "127.0.0.1:7999"
	}

	cmd := exec.Command(bin, "cli", "hv", "serve", "--addr", addr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return "", nil, fmt.Errorf("start %s: %w", bin, err)
	}
	cleanup := func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}

	url := "http://" + addr + "/"
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	for {
		if ctx.Err() != nil {
			cleanup()
			return "", nil, fmt.Errorf("hv serve did not come up at %s within timeout", addr)
		}
		if c, err := net.DialTimeout("tcp", addr, 500*time.Millisecond); err == nil {
			c.Close()
			if resp, err := http.Get(url); err == nil {
				resp.Body.Close()
				fmt.Fprintf(os.Stderr, "frank: hypervisor up at %s\n", url)
				return url, cleanup, nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
}
