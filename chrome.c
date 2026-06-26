/* Frank — browser chrome: toolbar with back/forward/reload + address bar. */
#include "browser.h"

static void on_address_activate(GtkEntry *entry, gpointer user) {
	frank_load(gtk_editable_get_text(GTK_EDITABLE(entry)));
}

static void on_back(GtkButton *b, gpointer user) {
	if (webkit_web_view_can_go_back(g_webview)) {
		webkit_web_view_go_back(g_webview);
	}
}

static void on_forward(GtkButton *b, gpointer user) {
	if (webkit_web_view_can_go_forward(g_webview)) {
		webkit_web_view_go_forward(g_webview);
	}
}

static void on_reload(GtkButton *b, gpointer user) {
	webkit_web_view_reload(g_webview);
}

GtkWidget *frank_make_toolbar(void) {
	GtkWidget *bar = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 6);
	gtk_widget_set_margin_top(bar, 6);
	gtk_widget_set_margin_bottom(bar, 6);
	gtk_widget_set_margin_start(bar, 6);
	gtk_widget_set_margin_end(bar, 6);

	g_back_btn = gtk_button_new_from_icon_name("go-previous-symbolic");
	g_fwd_btn = gtk_button_new_from_icon_name("go-next-symbolic");
	GtkWidget *reload = gtk_button_new_from_icon_name("view-refresh-symbolic");
	gtk_widget_set_sensitive(g_back_btn, FALSE);
	gtk_widget_set_sensitive(g_fwd_btn, FALSE);

	g_url_entry = gtk_entry_new();
	gtk_widget_set_hexpand(g_url_entry, TRUE);
	gtk_entry_set_placeholder_text(GTK_ENTRY(g_url_entry), "Enter a web address");

	g_signal_connect(g_back_btn, "clicked", G_CALLBACK(on_back), NULL);
	g_signal_connect(g_fwd_btn, "clicked", G_CALLBACK(on_forward), NULL);
	g_signal_connect(reload, "clicked", G_CALLBACK(on_reload), NULL);
	g_signal_connect(g_url_entry, "activate", G_CALLBACK(on_address_activate), NULL);

	gtk_box_append(GTK_BOX(bar), g_back_btn);
	gtk_box_append(GTK_BOX(bar), g_fwd_btn);
	gtk_box_append(GTK_BOX(bar), reload);
	gtk_box_append(GTK_BOX(bar), g_url_entry);
	return bar;
}
