// Command cef-spike embeds Chromium via CEF (C API, Views framework) in a native
// Go program and loads a URL, to evaluate embedded-Chromium rendering vs WebKitGTK.
//
// - resizable window, larger default size
// - unique cache dir per launch (avoids Chromium's single-instance lock so the
//   window opens on every run)
// - fps printed to stdout via an injected rAF counter + console-message handler
//   (the in-page --show-fps-counter HUD is hidden behind the test page's top bar)
// - --no-zygote / --no-sandbox so CEF's process model doesn't fight the Go runtime
//
// Usage: cef-spike [url]
package main

/*
#cgo CFLAGS: -I/usr/include/cef
#cgo LDFLAGS: -L/usr/lib/cef -lcef -Wl,-rpath,/usr/lib/cef

#define CEF_API_VERSION 14800

#include <stdio.h>
#include <string.h>
#include "include/cef_api_hash.h"
#include "include/capi/cef_app_capi.h"
#include "include/capi/cef_client_capi.h"
#include "include/capi/cef_life_span_handler_capi.h"
#include "include/capi/cef_display_handler_capi.h"
#include "include/capi/cef_load_handler_capi.h"
#include "include/capi/cef_frame_capi.h"
#include "include/capi/cef_browser_process_handler_capi.h"
#include "include/capi/cef_command_line_capi.h"
#include "include/capi/views/cef_browser_view_capi.h"
#include "include/capi/views/cef_window_capi.h"
#include "include/capi/views/cef_window_delegate_capi.h"
#include "include/capi/views/cef_panel_capi.h"

static cef_app_t g_app;
static cef_client_t g_client;
static cef_life_span_handler_t g_lsh;
static cef_display_handler_t g_dh;
static cef_load_handler_t g_lh;
static cef_browser_process_handler_t g_bph;
static cef_window_delegate_t g_wd;
static cef_browser_view_t *g_browser_view;
static const char *g_url;
static const char *g_cache;

static void setstr(cef_string_t *out, const char *s) {
	cef_string_utf8_to_utf16(s, strlen(s), out);
}

// --- no-op singleton ref counting (objects live for the whole process) ---
static void noop_add_ref(cef_base_ref_counted_t *self) {}
static int  noop_release(cef_base_ref_counted_t *self) { return 0; }
static int  has_one_ref(cef_base_ref_counted_t *self) { return 1; }
static int  has_at_least_one_ref(cef_base_ref_counted_t *self) { return 1; }
static void rc(cef_base_ref_counted_t *b, size_t sz) {
	b->size = sz;
	b->add_ref = noop_add_ref;
	b->release = noop_release;
	b->has_one_ref = has_one_ref;
	b->has_at_least_one_ref = has_at_least_one_ref;
}

// --- app ---
static void app_on_cmdline(cef_app_t *self, const cef_string_t *process_type, cef_command_line_t *cl) {
	cef_string_t s = {0};
	setstr(&s, "no-zygote");        cl->append_switch(cl, &s); cef_string_clear(&s);
	setstr(&s, "show-fps-counter"); cl->append_switch(cl, &s); cef_string_clear(&s);
}
static cef_browser_process_handler_t *app_get_bph(cef_app_t *self) { return &g_bph; }

// --- client ---
static cef_life_span_handler_t *client_get_lsh(cef_client_t *self) { return &g_lsh; }
static cef_display_handler_t   *client_get_dh(cef_client_t *self)  { return &g_dh; }
static cef_load_handler_t      *client_get_lh(cef_client_t *self)  { return &g_lh; }

static void lsh_on_before_close(cef_life_span_handler_t *self, cef_browser_t *browser) {
	cef_quit_message_loop();
}

// print page console output (carries our fps log) to stdout
static int dh_on_console_message(cef_display_handler_t *self, cef_browser_t *browser,
                                 cef_log_severity_t level, const cef_string_t *message,
                                 const cef_string_t *source, int line) {
	if (message && message->str) {
		cef_string_utf8_t out;
		memset(&out, 0, sizeof(out));
		cef_string_utf16_to_utf8(message->str, message->length, &out);
		if (out.str) printf("[page] %s\n", out.str);
		fflush(stdout);
		cef_string_utf8_clear(&out);
	}
	return 0;
}

// inject an fps counter once the page finishes loading
static void lh_on_load_end(cef_load_handler_t *self, cef_browser_t *browser,
                           cef_frame_t *frame, int httpStatusCode) {
	const char *js =
		"(function(){if(window.__fps)return;window.__fps=1;var n=0,t=performance.now();"
		"function L(){n++;var d=(performance.now()-t)/1000;"
		"if(d>=1){console.log('FPS '+(n/d).toFixed(1));n=0;t=performance.now();}"
		"requestAnimationFrame(L);}requestAnimationFrame(L);})()";
	cef_string_t code = {0}, surl = {0};
	setstr(&code, js);
	setstr(&surl, "frank://fps");
	frame->execute_java_script(frame, &code, &surl, 0);
	cef_string_clear(&code);
	cef_string_clear(&surl);
}

// --- window delegate (resizable, sized) ---
static int wd_can_resize(cef_window_delegate_t *self, cef_window_t *w)   { return 1; }
static int wd_can_maximize(cef_window_delegate_t *self, cef_window_t *w) { return 1; }
static int wd_can_minimize(cef_window_delegate_t *self, cef_window_t *w) { return 1; }
static int wd_can_close(cef_window_delegate_t *self, cef_window_t *w)    { return 1; }

static void wd_on_window_created(cef_window_delegate_t *self, cef_window_t *window) {
	cef_panel_t *panel = (cef_panel_t *)window;
	panel->add_child_view(panel, (cef_view_t *)g_browser_view);
	cef_string_t title = {0};
	setstr(&title, "Frank CEF Spike");
	window->set_title(window, &title);
	cef_string_clear(&title);
	cef_size_t sz; sz.width = 1280; sz.height = 900;
	window->center_window(window, &sz);
	window->show(window);
}

static void bph_on_context_initialized(cef_browser_process_handler_t *self) {
	cef_string_t url = {0};
	setstr(&url, g_url);
	cef_browser_settings_t bs;
	memset(&bs, 0, sizeof(bs));
	bs.size = sizeof(bs);
	g_browser_view = cef_browser_view_create(&g_client, &url, &bs, NULL, NULL, NULL);
	cef_string_clear(&url);
	cef_window_create_top_level(&g_wd);
}

static void init_structs(void) {
	rc(&g_app.base, sizeof(g_app));
	g_app.on_before_command_line_processing = app_on_cmdline;
	g_app.get_browser_process_handler = app_get_bph;

	rc(&g_bph.base, sizeof(g_bph));
	g_bph.on_context_initialized = bph_on_context_initialized;

	rc(&g_client.base, sizeof(g_client));
	g_client.get_life_span_handler = client_get_lsh;
	g_client.get_display_handler = client_get_dh;
	g_client.get_load_handler = client_get_lh;

	rc(&g_lsh.base, sizeof(g_lsh));
	g_lsh.on_before_close = lsh_on_before_close;

	rc(&g_dh.base, sizeof(g_dh));
	g_dh.on_console_message = dh_on_console_message;

	rc(&g_lh.base, sizeof(g_lh));
	g_lh.on_load_end = lh_on_load_end;

	rc(&g_wd.base.base.base, sizeof(g_wd)); // window_delegate -> panel_delegate -> view_delegate -> base
	g_wd.on_window_created = wd_on_window_created;
	g_wd.can_resize = wd_can_resize;
	g_wd.can_maximize = wd_can_maximize;
	g_wd.can_minimize = wd_can_minimize;
	g_wd.can_close = wd_can_close;
}

static int run_cef(int argc, char **argv, const char *url, const char *cache) {
	// Register our compiled API version before any other call (else libcef sees -1).
	cef_api_hash(CEF_API_VERSION, 0);

	g_url = url;
	g_cache = cache;
	init_structs();

	cef_main_args_t main_args;
	main_args.argc = argc;
	main_args.argv = argv;

	int code = cef_execute_process(&main_args, &g_app, NULL);
	if (code >= 0) return code; // this invocation was a CEF subprocess

	cef_settings_t settings;
	memset(&settings, 0, sizeof(settings));
	settings.size = sizeof(settings);
	settings.no_sandbox = 1;
	setstr(&settings.resources_dir_path, "/usr/lib/cef");
	setstr(&settings.locales_dir_path, "/usr/lib/cef/locales");
	setstr(&settings.root_cache_path, g_cache);

	cef_initialize(&main_args, &settings, &g_app, NULL);
	cef_run_message_loop();
	cef_shutdown();
	return 0;
}
*/
import "C"

import (
	"os"
	"runtime"
	"strings"
	"unsafe"
)

func main() {
	runtime.LockOSThread()

	url := "file:///home/d0mo/go/src/github.com/0magnet/chaosrack/index.html"
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		url = os.Args[1]
	}

	// Unique cache dir per launch so Chromium's single-instance lock never blocks
	// a relaunch (the "window won't open the second time" problem).
	cache, err := os.MkdirTemp("", "frank-cef-")
	if err != nil {
		cache = "/tmp/frank-cef-cache"
	}

	argv := make([]*C.char, len(os.Args)+1)
	for i, a := range os.Args {
		argv[i] = C.CString(a)
	}
	curl := C.CString(url)
	ccache := C.CString(cache)

	code := C.run_cef(C.int(len(os.Args)), (**C.char)(unsafe.Pointer(&argv[0])), curl, ccache)
	os.Exit(int(code))
}
