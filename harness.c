/* Frank — SharedWorker persistence harness (experimental: --sharedworker-demo).
 *
 * Serves three assets from ONE origin (frank://frank/...) so a SharedWorker is
 * shared across pages, marks the scheme secure+CORS (SharedWorker needs a proper
 * secure origin), and provides an always-present "anchor" WebView that holds a
 * worker connection so the worker (the stand-in visor) survives main-view
 * reloads. This is the Frank-side proof of the pattern in harness/sharedworker/.
 *
 * To confirm persistence WITHOUT a screenshot: the worker logs "BOOTED" exactly
 * once, and the anchor + tab pages log the SAME uptime (one shared worker). If
 * SharedWorker did not work on this origin we'd see an error or a second BOOT.
 */
#include "browser.h"

static const char *WORKER_JS =
	"const t0=Date.now();let booted=false,ticks=0;const ports=new Set();\n"
	/* forward worker logs to connected ports so they surface in a page console
	   (the real visor would route its debug output through here too). */
	"function vlog(){var s=Array.prototype.slice.call(arguments).join(' ');\n"
	" ports.forEach(function(p){try{p.postMessage({type:'log',line:s});}catch(e){}});}\n"
	"function boot(){if(booted)return;booted=true;vlog('BOOTED (exactly once)');\n"
	" setInterval(function(){ticks++;var m={type:'status',up:Date.now()-t0,ticks:ticks,ports:ports.size};\n"
	"  ports.forEach(function(p){try{p.postMessage(m);}catch(e){ports.delete(p);}});},1000);}\n"
	"onconnect=function(e){var p=e.ports[0];ports.add(p);boot();\n"
	" vlog('client connected; ports='+ports.size);\n"
	" p.onmessage=function(m){if(m.data==='status')p.postMessage({type:'status',up:Date.now()-t0,ticks:ticks,ports:ports.size});};\n"
	" p.start();};\n";

static const char *ANCHOR_HTML =
	"<!doctype html><meta charset=utf-8><title>anchor</title>"
	"<body style='font:12px sans-serif;margin:4px;color:#2b6cb0'>"
	"<span id=o>visor anchor starting…</span>"
	"<script>var o=document.getElementById('o');"
	"var w=new SharedWorker('visor-worker.js',{name:'frank-visor'});w.port.start();"
	"w.port.onmessage=function(m){var d=m.data;"
	"if(d.type==='status'){o.textContent='visor alive — uptime '+Math.round(d.up/1000)+'s, ports '+d.ports;}"
	"else if(d.type==='log'){console.log('[visor] '+d.line);}};"
	"</script></body>";

static const char *TAB_HTML =
	"<!doctype html><meta charset=utf-8><title>visor tab</title>"
	"<body style='font:15px sans-serif;padding:2em'>"
	"<h2>SharedWorker persistence demo</h2>"
	"<p>Reload this page (Ctrl+R). The uptime keeps climbing — the visor did NOT "
	"restart — because the anchor strip at the bottom holds the SharedWorker alive.</p>"
	"<pre id=o>connecting…</pre>"
	"<script>var o=document.getElementById('o');"
	"var w=new SharedWorker('visor-worker.js',{name:'frank-visor'});w.port.start();"
	"w.port.onmessage=function(m){if(m.data.type==='status'){"
	"o.textContent='visor uptime: '+Math.round(m.data.up/1000)+'s (ticks '+m.data.ticks+', ports '+m.data.ports+')';"
	"console.log('[tab] uptime '+Math.round(m.data.up/1000)+'s');}};"
	"w.port.postMessage('status');"
	"</script></body>";

static void on_frank_request(WebKitURISchemeRequest *req, gpointer user) {
	const char *path = webkit_uri_scheme_request_get_path(req);
	const char *body;
	const char *mime;
	if (path != NULL && strstr(path, "visor-worker.js") != NULL) {
		body = WORKER_JS; mime = "text/javascript";
	} else if (path != NULL && strstr(path, "tab") != NULL) {
		body = TAB_HTML; mime = "text/html";
	} else {
		body = ANCHOR_HTML; mime = "text/html";
	}
	gsize len = strlen(body);
	GInputStream *s = g_memory_input_stream_new_from_data(g_strdup(body), len, g_free);
	webkit_uri_scheme_request_finish(req, s, (gint64)len, mime);
	g_object_unref(s);
}

void frank_register_frank_scheme(void) {
	WebKitWebContext *ctx = webkit_web_context_get_default();
	webkit_web_context_register_uri_scheme(ctx, "frank", on_frank_request, NULL, NULL);
	WebKitSecurityManager *sm = webkit_web_context_get_security_manager(ctx);
	webkit_security_manager_register_uri_scheme_as_secure(sm, "frank");
	webkit_security_manager_register_uri_scheme_as_cors_enabled(sm, "frank");
}

static WebKitWebView *g_anchor_view = NULL;

GtkWidget *frank_make_anchor(void) {
	GtkWidget *a = webkit_web_view_new();
	g_anchor_view = WEBKIT_WEB_VIEW(a);
	WebKitSettings *s = webkit_web_view_get_settings(g_anchor_view);
	webkit_settings_set_enable_write_console_messages_to_stdout(s, TRUE);
	webkit_settings_set_enable_developer_extras(s, TRUE); /* so Visor Console can inspect it */
	gtk_widget_set_size_request(a, -1, 26); /* a thin always-present anchor strip */

	/* If a real visor server is up (Frank-managed `hv serve`), the anchor loads
	 * it so the actual wasm-visor boots under the hood; otherwise the SharedWorker
	 * stub keeps the persistence pattern working. */
	const char *vurl = g_getenv("FRANK_VISOR_URL");
	if (vurl != NULL && vurl[0] != '\0') {
		webkit_web_view_load_uri(g_anchor_view, vurl);
	} else {
		webkit_web_view_load_uri(g_anchor_view, "frank://frank/anchor.html");
	}
	return a;
}

/* Open DevTools on the anchor view, where the visor's forwarded logs land. */
void frank_show_visor_console(void) {
	if (g_anchor_view == NULL) {
		return;
	}
	WebKitWebInspector *insp = webkit_web_view_get_inspector(g_anchor_view);
	if (insp) {
		webkit_web_inspector_show(insp);
	}
}
