// Command frank is a native web browser built on an embedded WebKitGTK-6.0
// (GTK4) engine, so it renders modern JS/WebAssembly/WebGL/CSS at full fidelity.
//
// This is the WebKit-GTK4 line of Frank. The original pure-Go (GLFW/OpenGL)
// renderer is preserved on the `pure-go` branch; its engine packages
// (mustard/bun/mayo/ketchup/gg) remain in this module but are no longer wired to
// the entry point.
package main

/*
#cgo pkg-config: gtk4 webkitgtk-6.0
#include <stdlib.h>
#include "browser.h"
*/
import "C"

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"unsafe"
)

func setIfUnset(k, v string) {
	if os.Getenv(k) == "" {
		os.Setenv(k, v)
	}
}

// Default homepage: a self-contained welcome page (no network needed).
const defaultURL = "data:text/html,<html><body style='font-family:sans-serif;padding:2em;color:%23222'>" +
	"<h1>Frank</h1><p>WebKitGTK-6.0 browser shell &mdash; type a web address above.</p></body></html>"

func main() {
	// GTK/WebKit must run on the main thread.
	runtime.LockOSThread()

	// Old-GPU compatibility profile. WebKit's fast dmabuf zero-copy compositing
	// path can't be presented by some older drivers (e.g. Gen7 Intel / GLES 3.0),
	// where it renders one frame then freezes. Default is the fast path (correct
	// on modern GPUs); set FRANK_GL_COMPAT=1 to force the stable readback path
	// (slower but freeze-free) on affected hardware.
	if os.Getenv("FRANK_GL_COMPAT") == "1" {
		setIfUnset("MESA_EXTENSION_OVERRIDE", "+GL_KHR_robustness")
		setIfUnset("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	}

	url := defaultURL
	hypervisor := false
	for _, a := range os.Args[1:] {
		switch {
		case a == "--hypervisor" || a == "-hypervisor":
			hypervisor = true
		case !strings.HasPrefix(a, "-"):
			url = a
		}
	}

	// Experimental: --hypervisor spawns `skywire cli hv serve` and points the
	// WebView at the in-tab wasm-visor (hypervisor UI + dmsg/skynet browse + host
	// + skysocks-lite). The subprocess is stopped when the window closes.
	if hypervisor {
		hvURL, cleanup, err := startHypervisor()
		if err != nil {
			fmt.Fprintln(os.Stderr, "frank: hypervisor:", err)
			os.Exit(1)
		}
		defer cleanup()
		url = hvURL
	}

	// FRANK_PROXY routes the WebView through a SOCKS/HTTP proxy, e.g. the visor's
	// dmsgweb resolving proxy: socks5://[user:pass@]127.0.0.1:<ephemeral-port>.
	proxy := os.Getenv("FRANK_PROXY")

	curl := C.CString(url)
	cproxy := C.CString(proxy)
	defer C.free(unsafe.Pointer(curl))
	defer C.free(unsafe.Pointer(cproxy))
	C.frank_run(curl, cproxy)
}
