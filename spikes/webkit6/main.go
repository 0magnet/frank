// Command webkit6-spike is the GTK4 / WebKitGTK-6.0 twin of the GTK3 spike.
// The WebKit engine core is identical (2.52.4), but the GTK4 port composites the
// WebView into the window through a different pipeline (GskRenderer + GTK4 dmabuf
// textures) than the GTK3 port. Our WebGL problem (dmabuf zero-copy freeze /
// readback half-speed) lives in that compositing layer, so this tests whether the
// GTK4 path presents the fast zero-copy path without freezing on old Intel.
//
// Tested with dmabuf ENABLED (the whole point). Sets the GL_KHR_robustness Mesa
// override (harmless; needed only if WebKit falls back to readback). Set
// SPIKE_NO_DMABUF=1 to also disable dmabuf for an A/B against the GTK3 result.
//
// Usage: webkit6-spike [url]
package main

/*
#cgo pkg-config: gtk4 webkitgtk-6.0
#include <gtk/gtk.h>
#include <webkit/webkit.h>
#include <stdio.h>

static const char *g_url;
static int g_autoquit;

static void on_js_done(GObject *obj, GAsyncResult *res, gpointer u) {
	GError *err = NULL;
	JSCValue *val = webkit_web_view_evaluate_javascript_finish(WEBKIT_WEB_VIEW(obj), res, &err);
	if (err) { printf("[probe] JS error: %s\n", err->message); g_error_free(err); fflush(stdout); return; }
	if (val) { char *s = jsc_value_to_string(val); printf("[probe] %s\n", s); g_free(s); }
	fflush(stdout);
}

static void on_fps_done(GObject *obj, GAsyncResult *res, gpointer u) {
	GError *err = NULL;
	JSCValue *val = webkit_web_view_evaluate_javascript_finish(WEBKIT_WEB_VIEW(obj), res, &err);
	if (err) { printf("[fps] JS error: %s\n", err->message); g_error_free(err); fflush(stdout); return; }
	if (val) { char *s = jsc_value_to_string(val); printf("[fps] %s\n", s); g_free(s); }
	fflush(stdout);
}

static void on_snapshot_done(GObject *obj, GAsyncResult *res, gpointer u) {
	GError *err = NULL;
	GdkTexture *tex = webkit_web_view_get_snapshot_finish(WEBKIT_WEB_VIEW(obj), res, &err);
	if (err) { printf("[snapshot] error: %s\n", err->message); g_error_free(err); fflush(stdout); return; }
	if (tex) {
		gboolean ok = gdk_texture_save_to_png(tex, "/tmp/frank_webkit6_spike.png");
		printf("[snapshot] saved=%d -> /tmp/frank_webkit6_spike.png\n", ok);
		g_object_unref(tex);
	}
	fflush(stdout);
}

static gboolean on_fps_timeout(gpointer data) {
	WebKitWebView *wv = WEBKIT_WEB_VIEW(data);
	const char *script =
		"(function(){var dt=(performance.now()-(window.__t0||performance.now()))/1000;"
		"return 'fps='+(dt>0?(window.__n/dt).toFixed(1):'0')+' frames='+(window.__n||0)+' over '+dt.toFixed(1)+'s';})()";
	webkit_web_view_evaluate_javascript(wv, script, -1, NULL, NULL, NULL, on_fps_done, NULL);
	webkit_web_view_get_snapshot(wv, WEBKIT_SNAPSHOT_REGION_VISIBLE, WEBKIT_SNAPSHOT_OPTIONS_NONE, NULL, on_snapshot_done, NULL);
	return FALSE;
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
		"+' WebAssembly='+(typeof WebAssembly)+' title=['+document.title+']';"
		"}catch(e){return 'probe-exception: '+e;}})()";
	webkit_web_view_evaluate_javascript(wv, script, -1, NULL, NULL, NULL, on_js_done, NULL);
	const char *fpsStart =
		"(function(){window.__t0=performance.now();window.__n=0;"
		"function loop(){window.__n++;requestAnimationFrame(loop);}requestAnimationFrame(loop);"
		"return 'fps-counter-started';})()";
	webkit_web_view_evaluate_javascript(wv, fpsStart, -1, NULL, NULL, NULL, NULL, NULL);
	g_timeout_add(3000, on_fps_timeout, wv);
	return FALSE;
}

static void on_load_changed(WebKitWebView *wv, WebKitLoadEvent ev, gpointer data) {
	if (ev == WEBKIT_LOAD_FINISHED) {
		printf("[load] finished; probing in 4s, fps over next 3s...\n"); fflush(stdout);
		g_timeout_add(4000, on_probe_timeout, wv);
	}
}

static gboolean quit_cb(gpointer app) { g_application_quit(G_APPLICATION(app)); return FALSE; }

static void on_activate(GtkApplication *app, gpointer data) {
	GtkWidget *win = gtk_application_window_new(app);
	gtk_window_set_default_size(GTK_WINDOW(win), 1100, 800);
	gtk_window_set_title(GTK_WINDOW(win), "Frank WebKit6/GTK4 Spike");

	WebKitWebView *wv = WEBKIT_WEB_VIEW(webkit_web_view_new());
	WebKitSettings *st = webkit_web_view_get_settings(wv);
	webkit_settings_set_enable_webgl(st, TRUE);
	webkit_settings_set_enable_javascript(st, TRUE);
	webkit_settings_set_enable_write_console_messages_to_stdout(st, TRUE);
	webkit_settings_set_hardware_acceleration_policy(st, WEBKIT_HARDWARE_ACCELERATION_POLICY_ALWAYS);

	g_signal_connect(wv, "load-changed", G_CALLBACK(on_load_changed), NULL);
	gtk_window_set_child(GTK_WINDOW(win), GTK_WIDGET(wv));

	webkit_web_view_load_uri(wv, g_url);
	gtk_window_present(GTK_WINDOW(win));

	if (g_autoquit > 0) g_timeout_add_seconds(g_autoquit, quit_cb, app);
}

static void run_viewer(const char *url, int autoquit) {
	g_url = url;
	g_autoquit = autoquit;
	GtkApplication *app = gtk_application_new("org.frank.webkit6spike", G_APPLICATION_DEFAULT_FLAGS);
	g_signal_connect(app, "activate", G_CALLBACK(on_activate), NULL);
	g_application_run(G_APPLICATION(app), 0, NULL);
	g_object_unref(app);
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
	// Robust readPixels override (harmless; only matters if WebKit falls back to
	// the readback path). dmabuf is left ENABLED on purpose — testing whether the
	// GTK4 compositing path handles the zero-copy path that the GTK3 port couldn't.
	setIfUnset("MESA_EXTENSION_OVERRIDE", "+GL_KHR_robustness")
	if os.Getenv("SPIKE_NO_DMABUF") == "1" {
		setIfUnset("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	}

	url := "file:///home/d0mo/go/src/github.com/0magnet/wasm-stuff/index.html"
	if len(os.Args) > 1 {
		url = os.Args[1]
	}
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
