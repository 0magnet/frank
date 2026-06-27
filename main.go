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

	url := "" // empty -> the configured homepage (or built-in welcome), chosen C-side
	skynet := false
	for _, a := range os.Args[1:] {
		switch {
		case a == "--skynet" || a == "-skynet":
			skynet = true
		case a == "--sharedworker-demo":
			os.Setenv("FRANK_SHAREDWORKER_DEMO", "1")
		case !strings.HasPrefix(a, "-"):
			url = a
		}
	}

	// Two ways to reach the Skywire network, both wired through WebKit's proxy:
	//   FRANK_PROXY set -> bring-your-own: a visor that spawns Frank (or an
	//                      already-running one) passes its resolving-proxy URI and
	//                      owns its own lifecycle (it survives Frank closing).
	//   --skynet        -> Frank starts and manages a resolving proxy on an
	//                      ephemeral loopback port for the browser's lifetime
	//                      (stand-in for the in-process routable visor to come).
	proxy := os.Getenv("FRANK_PROXY")
	if proxy == "" && skynet {
		p, cleanup, err := startManagedProxy()
		if err != nil {
			fmt.Fprintln(os.Stderr, "frank: skynet:", err)
			os.Exit(1)
		}
		defer cleanup()
		proxy = p
	}

	curl := C.CString(url)
	cproxy := C.CString(proxy)
	defer C.free(unsafe.Pointer(curl))
	defer C.free(unsafe.Pointer(cproxy))
	C.frank_run(curl, cproxy)
}
