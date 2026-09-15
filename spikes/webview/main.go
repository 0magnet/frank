// Command webview-spike embeds a WebKitGTK (WebKit2) browser view in a native
// Go program and loads a URL, to prove that a CGO-embedded OS engine can render
// full JS + WebAssembly + WebGL content (the gating requirement for Frank).
//
// It also probes the live page after load (canvas/WebGL/WebAssembly presence,
// GL RENDERER string) and writes a PNG snapshot to /tmp so the result can be
// verified without watching the screen.
//
// Usage: webview-spike [url]   (default: the wasm-stuff WebGL test page)
package main

/*
#cgo pkg-config: webkit2gtk-4.1
#include <gtk/gtk.h>
#include <webkit2/webkit2.h>
#include <cairo.h>
#include <stdio.h>

// JS probe result -> stdout
static void on_js_done(GObject *obj, GAsyncResult *res, gpointer user_data) {
	GError *err = NULL;
	JSCValue *val = webkit_web_view_evaluate_javascript_finish(WEBKIT_WEB_VIEW(obj), res, &err);
	if (err) { printf("[probe] JS error: %s\n", err->message); g_error_free(err); fflush(stdout); return; }
	if (val) {
		char *s = jsc_value_to_string(val);
		printf("[probe] %s\n", s);
		g_free(s);
	}
	fflush(stdout);
}

// snapshot -> /tmp/frank_webview_spike.png
static void on_snapshot_done(GObject *obj, GAsyncResult *res, gpointer user_data) {
	GError *err = NULL;
	cairo_surface_t *surf = webkit_web_view_get_snapshot_finish(WEBKIT_WEB_VIEW(obj), res, &err);
	if (err) { printf("[snapshot] error: %s\n", err->message); g_error_free(err); fflush(stdout); return; }
	if (surf) {
		cairo_status_t st = cairo_surface_write_to_png(surf, "/tmp/frank_webview_spike.png");
		printf("[snapshot] write status=%d -> /tmp/frank_webview_spike.png\n", (int)st);
		cairo_surface_destroy(surf);
	}
	fflush(stdout);
}

// fps read-back result -> stdout
static void on_fps_done(GObject *obj, GAsyncResult *res, gpointer user_data) {
	GError *err = NULL;
	JSCValue *val = webkit_web_view_evaluate_javascript_finish(WEBKIT_WEB_VIEW(obj), res, &err);
	if (err) { printf("[fps] JS error: %s\n", err->message); g_error_free(err); fflush(stdout); return; }
	if (val) { char *s = jsc_value_to_string(val); printf("[fps] %s\n", s); g_free(s); }
	fflush(stdout);
}

// read the fps counter (started at probe time), then snapshot
static gboolean on_fps_timeout(gpointer data) {
	WebKitWebView *wv = WEBKIT_WEB_VIEW(data);
	const char *script =
		"(function(){var dt=(performance.now()-(window.__t0||performance.now()))/1000;"
		"return 'fps='+(dt>0?(window.__n/dt).toFixed(1):'0')+' frames='+(window.__n||0)+' over '+dt.toFixed(1)+'s';})()";
	webkit_web_view_evaluate_javascript(wv, script, -1, NULL, NULL, NULL, on_fps_done, NULL);
	webkit_web_view_get_snapshot(wv, WEBKIT_SNAPSHOT_REGION_VISIBLE, WEBKIT_SNAPSHOT_OPTIONS_NONE, NULL, on_snapshot_done, NULL);
	return FALSE; // run once
}

static gboolean on_probe_timeout(gpointer data) {
	WebKitWebView *wv = WEBKIT_WEB_VIEW(data);
	const char *script =
		"(function(){try{"
		"var c=document.querySelector('canvas');"
		"var gl=c?(c.getContext('webgl2')||c.getContext('webgl')):null;"
		"var ver=gl?gl.getParameter(gl.VERSION):'none';"
		"var rend=gl?gl.getParameter(gl.RENDERER):'none';"
		"return 'canvas='+(!!c)+' webgl='+(!!gl)+' GL_VERSION=['+ver+'] RENDERER=['+rend+']'"
		"+' WebAssembly='+(typeof WebAssembly)+' title=['+document.title+']'"
		"+' bodyLen='+(document.body?document.body.innerHTML.length:0);"
		"}catch(e){return 'probe-exception: '+e;}})()";
	webkit_web_view_evaluate_javascript(wv, script, -1, NULL, NULL, NULL, on_js_done, NULL);
	// start an independent rAF frame counter to measure sustained fps
	const char *fpsStart =
		"(function(){window.__t0=performance.now();window.__n=0;"
		"function loop(){window.__n++;requestAnimationFrame(loop);}requestAnimationFrame(loop);"
		"return 'fps-counter-started';})()";
	webkit_web_view_evaluate_javascript(wv, fpsStart, -1, NULL, NULL, NULL, NULL, NULL);
	g_timeout_add(3000, on_fps_timeout, wv); // measure over 3s
	return FALSE; // run once
}

static void on_load_changed(WebKitWebView *wv, WebKitLoadEvent ev, gpointer data) {
	if (ev == WEBKIT_LOAD_STARTED)  { printf("[load] started\n");  fflush(stdout); }
	if (ev == WEBKIT_LOAD_COMMITTED){ printf("[load] committed\n");fflush(stdout); }
	if (ev == WEBKIT_LOAD_FINISHED) {
		printf("[load] finished; probing in 4s, fps over next 3s...\n"); fflush(stdout);
		g_timeout_add(4000, on_probe_timeout, wv);
	}
}

static gboolean quit_cb(gpointer d) { gtk_main_quit(); return FALSE; }

static void run_viewer(const char *url, int autoquit_secs) {
	gtk_init(NULL, NULL);
	if (autoquit_secs > 0) {
		g_timeout_add_seconds(autoquit_secs, quit_cb, NULL);
	}
	GtkWidget *win = gtk_window_new(GTK_WINDOW_TOPLEVEL);
	gtk_window_set_default_size(GTK_WINDOW(win), 1100, 800);
	gtk_window_set_title(GTK_WINDOW(win), "Frank WebView Spike");

	WebKitWebView *wv = WEBKIT_WEB_VIEW(webkit_web_view_new());
	WebKitSettings *st = webkit_web_view_get_settings(wv);
	webkit_settings_set_enable_webgl(st, TRUE);
	webkit_settings_set_enable_javascript(st, TRUE);
	webkit_settings_set_enable_write_console_messages_to_stdout(st, TRUE);
	webkit_settings_set_enable_developer_extras(st, TRUE);
	webkit_settings_set_hardware_acceleration_policy(st, WEBKIT_HARDWARE_ACCELERATION_POLICY_ALWAYS);

	g_signal_connect(wv, "load-changed", G_CALLBACK(on_load_changed), NULL);
	gtk_container_add(GTK_CONTAINER(win), GTK_WIDGET(wv));
	g_signal_connect(win, "destroy", G_CALLBACK(gtk_main_quit), NULL);

	webkit_web_view_load_uri(wv, url);
	gtk_widget_show_all(win);
	gtk_main();
}
*/
import "C"

import (
	"os"
	"strconv"
	"unsafe"
)

func setIfUnset(k, v string) {
	if os.Getenv(k) == "" {
		os.Setenv(k, v)
	}
}

func main() {
	// WebGL compatibility profile for old Intel/Mesa GPUs (e.g. HD 4000 / IvyBridge,
	// GLES 3.0): expose the robust readPixels entrypoint ANGLE needs for canvas
	// readback, and avoid the dmabuf zero-copy compositing path that stalls on this
	// hardware. Set before any GTK/WebKit/Mesa init; child WebProcesses inherit these.
	// NOTE: this is a COMPATIBILITY PROFILE, not universal — on modern GPUs disabling
	// dmabuf costs zero-copy; production Frank should gate this on GPU detection.
	// Set SPIKE_NO_COMPAT=1 to skip; existing env values are not overridden.
	if os.Getenv("SPIKE_NO_COMPAT") == "" {
		setIfUnset("MESA_EXTENSION_OVERRIDE", "+GL_KHR_robustness")
		setIfUnset("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	}

	url := "file:///home/d0mo/go/src/github.com/0magnet/wasm-stuff/index.html"
	if len(os.Args) > 1 {
		url = os.Args[1]
	}
	// SPIKE_AUTOQUIT=<seconds> makes the window self-close, for headless A/B testing.
	autoquit := 0
	if v := os.Getenv("SPIKE_AUTOQUIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			autoquit = n
		}
	}
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))
	C.run_viewer(curl, C.int(autoquit))
}
