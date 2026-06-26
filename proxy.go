package main

// Frank-managed Skywire resolving proxy (experimental).
//
// With --skynet, Frank starts a SOCKS5 resolving proxy on a fresh ephemeral
// loopback port for the browser's lifetime and points the WebView at it, so
// http://<pk>.dmsg/ resolves over Skywire transparently. Today this spawns the
// standalone `skywire dmsg web` (dmsg-relay only) as a stand-in; it will be
// replaced by an in-process routable visor (transports + router) exposing the
// same SOCKS interface — the WebView wiring does not change.
//
// Lifecycle: the proxy is a child of Frank and is killed when Frank exits, so
// the visor shares the browser's lifecycle. (The alternative model — a
// host-native visor that spawns Frank and outlives it — is supported instead by
// setting FRANK_PROXY; see main.go.)
//
// TODO(security): the loopback SOCKS port is reachable by any local process.
// The fix is SOCKS5 auth on the proxy (dmsgweb has none today) plus a random
// credential minted here and carried in the socks5://user:pass@ URI. The
// ephemeral random port is only interim obscurity.

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
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

// freePort asks the OS for an unused loopback TCP port.
func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// startManagedProxy launches the resolving proxy on a fresh ephemeral port for
// the browser's lifetime, returning its proxy URI and a cleanup func.
func startManagedProxy() (string, func(), error) {
	bin := locateSkywire()
	if bin == "" {
		return "", nil, fmt.Errorf("skywire binary not found (set SKYWIRE_BIN, or put it on PATH)")
	}
	port, err := freePort()
	if err != nil {
		return "", nil, fmt.Errorf("pick ephemeral port: %w", err)
	}

	cmd := exec.Command(bin, "dmsg", "web", "--socks", fmt.Sprint(port), "--loglvl", "error")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	// Own process group (so we can take down any children the proxy spawns) plus
	// Pdeathsig (so the kernel kills the proxy if Frank dies, even on a crash) —
	// together the proxy truly shares the browser's lifecycle and never orphans.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGTERM}
	if err := cmd.Start(); err != nil {
		return "", nil, fmt.Errorf("start dmsg web: %w", err)
	}
	cleanup := func() {
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM) // whole process group
			_ = cmd.Process.Kill()
		}
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for {
		if ctx.Err() != nil {
			cleanup()
			return "", nil, fmt.Errorf("resolving proxy not up on %s within timeout", addr)
		}
		if c, derr := net.DialTimeout("tcp", addr, 500*time.Millisecond); derr == nil {
			c.Close()
			fmt.Fprintf(os.Stderr, "frank: managed skynet proxy up at socks5://%s\n", addr)
			return "socks5://" + addr, cleanup, nil
		}
		time.Sleep(250 * time.Millisecond)
	}
}
