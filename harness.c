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
	"function boot(){if(booted)return;booted=true;console.log('[worker] BOOTED (exactly once)');\n"
	" setInterval(function(){ticks++;var m={type:'status',up:Date.now()-t0,ticks:ticks,ports:ports.size};\n"
	"  ports.forEach(function(p){try{p.postMessage(m);}catch(e){ports.delete(p);}});},1000);}\n"
	"onconnect=function(e){var p=e.ports[0];ports.add(p);boot();\n"
	" console.log('[worker] connect; ports='+ports.size);\n"
	" p.onmessage=function(m){if(m.data==='status')p.postMessage({type:'status',up:Date.now()-t0,ticks:ticks,ports:ports.size});};\n"
	" p.start();};\n";

static const char *ANCHOR_HTML =
	"<!doctype html><meta charset=utf-8><title>anchor</title>"
	"<body style='font:12px sans-serif;margin:4px;color:#2b6cb0'>"
	"<span id=o>visor anchor starting…</span>"
	"<script>var o=document.getElementById('o');"
	"var w=new SharedWorker('visor-worker.js',{name:'frank-visor'});w.port.start();"
	"w.port.onmessage=function(m){if(m.data.type==='status'){"
	"o.textContent='visor alive — uptime '+Math.round(m.data.up/1000)+'s, ports '+m.data.ports;"
	"if(m.data.ticks%5===0)console.log('[anchor] uptime '+Math.round(m.data.up/1000)+'s');}};"
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

GtkWidget *frank_make_anchor(void) {
	GtkWidget *a = webkit_web_view_new();
	WebKitSettings *s = webkit_web_view_get_settings(WEBKIT_WEB_VIEW(a));
	webkit_settings_set_enable_write_console_messages_to_stdout(s, TRUE);
	gtk_widget_set_size_request(a, -1, 26); /* a thin always-present anchor strip */
	webkit_web_view_load_uri(WEBKIT_WEB_VIEW(a), "frank://frank/anchor.html");
	return a;
}
