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
	"os"
	"runtime"
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
	if len(os.Args) > 1 {
		url = os.Args[1]
	}

	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))
	C.frank_run(curl)
}
