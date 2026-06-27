/* Frank — WebKit-GTK4 browser shell (C side).
 *
 * The pure-Go rendering engine (mustard/bun/mayo/ketchup/gg) is retained in the
 * module but no longer wired to the entry point; rendering is now an embedded
 * WebKitGTK-6.0 (GTK4) WebView. See main.go for the Go entry + GPU compat profile.
 */
#ifndef FRANK_BROWSER_H
#define FRANK_BROWSER_H

#include <gtk/gtk.h>
#include <webkit/webkit.h>
#include <stdio.h>
#include <string.h>

/* Shared widgets (defined in browser.c). */
extern WebKitWebView *g_webview;
extern GtkWidget *g_url_entry;
extern GtkWidget *g_back_btn;
extern GtkWidget *g_fwd_btn;
extern GtkWidget *g_main_window; /* for transient dialogs */

/* settings.c — GKeyFile-backed preferences. */
extern gboolean g_setting_javascript;
extern gboolean g_setting_webgl;
extern gboolean g_setting_devtools;
extern gdouble g_setting_zoom;
void frank_settings_load(void);
void frank_settings_save(void);
void frank_settings_apply(WebKitWebView *wv);
void frank_show_preferences(GtkWidget *parent);

/* menu.c — menu bar + application actions. */
void frank_install_actions(GtkApplication *app);
GtkWidget *frank_make_menubar(void);

/* webview.c */
GtkWidget *frank_make_webview(void);
void frank_load(const char *text); /* navigate; normalizes bare input to https */
void frank_on_load_changed(WebKitWebView *wv, WebKitLoadEvent ev, gpointer data);

/* chrome.c */
GtkWidget *frank_make_toolbar(void);

/* browser.c — entry point + shared state. */
extern const char *g_proxy_uri; /* WebView proxy URI (socks5://[user:pass@]host:port), or "" */
void frank_run(const char *initial_url, const char *proxy_uri);

#endif /* FRANK_BROWSER_H */
