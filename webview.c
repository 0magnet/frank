/* Frank — the embedded WebKitGTK-6.0 WebView and navigation. */
#include "browser.h"

GtkWidget *frank_make_webview(void) {
	GtkWidget *w = webkit_web_view_new();
	g_webview = WEBKIT_WEB_VIEW(w);

	WebKitSettings *s = webkit_web_view_get_settings(g_webview);
	webkit_settings_set_enable_write_console_messages_to_stdout(s, TRUE);
	webkit_settings_set_hardware_acceleration_policy(s, WEBKIT_HARDWARE_ACCELERATION_POLICY_ALWAYS);
	frank_settings_apply(g_webview); /* JavaScript, WebGL, devtools, zoom from preferences */

	gtk_widget_set_hexpand(w, TRUE);
	gtk_widget_set_vexpand(w, TRUE);

	// Route all WebView traffic through the visor's SOCKS resolving proxy when
	// configured, so http://<pk>.dmsg/ (and its subresources) resolve over the
	// Skywire network transparently. The proxy URI may carry credentials.
	if (g_proxy_uri != NULL && g_proxy_uri[0] != '\0') {
		WebKitNetworkSession *session = webkit_web_view_get_network_session(g_webview);
		WebKitNetworkProxySettings *ps = webkit_network_proxy_settings_new(g_proxy_uri, NULL);
		webkit_network_session_set_proxy_settings(session, WEBKIT_NETWORK_PROXY_MODE_CUSTOM, ps);
		webkit_network_proxy_settings_free(ps);
		printf("[frank] routing WebView via proxy %s\n", g_proxy_uri);
		fflush(stdout);
	}

	g_signal_connect(g_webview, "load-changed", G_CALLBACK(frank_on_load_changed), NULL);
	return w;
}

/* Navigate to user input: pass through anything with a scheme, otherwise treat
 * as a hostname and default to https. (dmsg://<pk> will be handled by a custom
 * scheme handler in a later step.) */
void frank_load(const char *text) {
	if (!text || !*text) {
		return;
	}
	if (strstr(text, "://") != NULL || strncmp(text, "about:", 6) == 0 ||
	    strncmp(text, "data:", 5) == 0) {
		webkit_web_view_load_uri(g_webview, text);
	} else {
		char buf[2048];
		snprintf(buf, sizeof(buf), "https://%s", text);
		webkit_web_view_load_uri(g_webview, buf);
	}
}

void frank_on_load_changed(WebKitWebView *wv, WebKitLoadEvent ev, gpointer data) {
	const char *uri = webkit_web_view_get_uri(wv);
	if (uri != NULL && (ev == WEBKIT_LOAD_COMMITTED || ev == WEBKIT_LOAD_FINISHED)) {
		gtk_editable_set_text(GTK_EDITABLE(g_url_entry), uri);
	}
	gtk_widget_set_sensitive(g_back_btn, webkit_web_view_can_go_back(wv));
	gtk_widget_set_sensitive(g_fwd_btn, webkit_web_view_can_go_forward(wv));
}
