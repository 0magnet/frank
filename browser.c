/* Frank — application + main window assembly. */
#include "browser.h"

WebKitWebView *g_webview = NULL;
GtkWidget *g_url_entry = NULL;
GtkWidget *g_back_btn = NULL;
GtkWidget *g_fwd_btn = NULL;
GtkWidget *g_main_window = NULL;
const char *g_proxy_uri = "";

static char g_initial_url[4096];

static void on_activate(GtkApplication *app, gpointer data) {
	frank_settings_load();
	frank_install_actions(app);

	gboolean demo = (g_getenv("FRANK_SHAREDWORKER_DEMO") != NULL);
	if (demo) {
		frank_register_frank_scheme(); /* must precede any WebView creation */
	}

	GtkWidget *win = gtk_application_window_new(app);
	g_main_window = win;
	gtk_window_set_title(GTK_WINDOW(win), "Frank");
	gtk_window_set_default_size(GTK_WINDOW(win), 1280, 900);

	GtkWidget *vbox = gtk_box_new(GTK_ORIENTATION_VERTICAL, 0);
	gtk_box_append(GTK_BOX(vbox), frank_make_menubar());
	gtk_box_append(GTK_BOX(vbox), frank_make_toolbar());
	gtk_box_append(GTK_BOX(vbox), frank_make_webview());
	if (demo) {
		gtk_box_append(GTK_BOX(vbox), frank_make_anchor()); /* always-present anchor strip */
	}
	gtk_window_set_child(GTK_WINDOW(win), vbox);

	if (demo) {
		webkit_web_view_load_uri(g_webview, "frank://frank/tab.html");
	} else if (g_initial_url[0] != '\0') {
		gtk_editable_set_text(GTK_EDITABLE(g_url_entry), g_initial_url);
		webkit_web_view_load_uri(g_webview, g_initial_url);
	} else {
		frank_load_home();
	}
	gtk_window_present(GTK_WINDOW(win));
}

void frank_run(const char *initial_url, const char *proxy_uri) {
	if (initial_url != NULL) {
		strncpy(g_initial_url, initial_url, sizeof(g_initial_url) - 1);
	}
	if (proxy_uri != NULL && proxy_uri[0] != '\0') {
		g_proxy_uri = strdup(proxy_uri);
	}
	// GApplication IDs forbid an element starting with a digit, so "0magnet" is
	// escaped to "_0magnet". NON_UNIQUE: every launch is its own process for now
	// (no single-instance forwarding); revisit if we want one-process/many-windows.
	GtkApplication *app = gtk_application_new("io.github._0magnet.frank", G_APPLICATION_NON_UNIQUE);
	g_signal_connect(app, "activate", G_CALLBACK(on_activate), NULL);
	g_application_run(G_APPLICATION(app), 0, NULL);
	g_object_unref(app);
}
