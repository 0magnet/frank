/* Frank — preferences, persisted to ~/.config/frank/frank.ini via GKeyFile. */
#include "browser.h"

gboolean g_setting_javascript = TRUE;
gboolean g_setting_webgl = TRUE;
gboolean g_setting_devtools = FALSE;
gdouble g_setting_zoom = 1.0;

static char *config_path(void) {
	char *dir = g_build_filename(g_get_user_config_dir(), "frank", NULL);
	g_mkdir_with_parents(dir, 0700);
	char *path = g_build_filename(dir, "frank.ini", NULL);
	g_free(dir);
	return path;
}

void frank_settings_load(void) {
	char *path = config_path();
	GKeyFile *kf = g_key_file_new();
	if (g_key_file_load_from_file(kf, path, G_KEY_FILE_NONE, NULL)) {
		GError *e = NULL;
		gboolean b;
		b = g_key_file_get_boolean(kf, "content", "javascript", &e); if (!e) g_setting_javascript = b; g_clear_error(&e);
		b = g_key_file_get_boolean(kf, "content", "webgl", &e); if (!e) g_setting_webgl = b; g_clear_error(&e);
		b = g_key_file_get_boolean(kf, "content", "devtools", &e); if (!e) g_setting_devtools = b; g_clear_error(&e);
		gdouble d = g_key_file_get_double(kf, "content", "zoom", &e); if (!e && d > 0.0) g_setting_zoom = d; g_clear_error(&e);
	}
	g_key_file_free(kf);
	g_free(path);
}

void frank_settings_save(void) {
	char *path = config_path();
	GKeyFile *kf = g_key_file_new();
	g_key_file_set_boolean(kf, "content", "javascript", g_setting_javascript);
	g_key_file_set_boolean(kf, "content", "webgl", g_setting_webgl);
	g_key_file_set_boolean(kf, "content", "devtools", g_setting_devtools);
	g_key_file_set_double(kf, "content", "zoom", g_setting_zoom);
	g_key_file_save_to_file(kf, path, NULL);
	g_key_file_free(kf);
	g_free(path);
}

void frank_settings_apply(WebKitWebView *wv) {
	if (wv == NULL) {
		return;
	}
	WebKitSettings *s = webkit_web_view_get_settings(wv);
	webkit_settings_set_enable_javascript(s, g_setting_javascript);
	webkit_settings_set_enable_webgl(s, g_setting_webgl);
	webkit_settings_set_enable_developer_extras(s, g_setting_devtools);
	webkit_web_view_set_zoom_level(wv, g_setting_zoom);
}

/* --- preferences dialog --- */

static void on_js_toggled(GObject *sw, GParamSpec *ps, gpointer d) {
	g_setting_javascript = gtk_switch_get_active(GTK_SWITCH(sw));
	frank_settings_apply(g_webview);
	frank_settings_save();
}
static void on_webgl_toggled(GObject *sw, GParamSpec *ps, gpointer d) {
	g_setting_webgl = gtk_switch_get_active(GTK_SWITCH(sw));
	frank_settings_apply(g_webview);
	frank_settings_save();
}
static void on_devtools_toggled(GObject *sw, GParamSpec *ps, gpointer d) {
	g_setting_devtools = gtk_switch_get_active(GTK_SWITCH(sw));
	frank_settings_apply(g_webview);
	frank_settings_save();
}
static void on_zoom_changed(GtkSpinButton *sb, gpointer d) {
	g_setting_zoom = gtk_spin_button_get_value(sb);
	frank_settings_apply(g_webview);
	frank_settings_save();
}

static GtkWidget *pref_row(const char *label, GtkWidget *control) {
	GtkWidget *box = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 12);
	GtkWidget *l = gtk_label_new(label);
	gtk_widget_set_hexpand(l, TRUE);
	gtk_widget_set_halign(l, GTK_ALIGN_START);
	gtk_widget_set_valign(control, GTK_ALIGN_CENTER);
	gtk_box_append(GTK_BOX(box), l);
	gtk_box_append(GTK_BOX(box), control);
	return box;
}

void frank_show_preferences(GtkWidget *parent) {
	GtkWidget *win = gtk_window_new();
	gtk_window_set_title(GTK_WINDOW(win), "Frank Preferences");
	gtk_window_set_default_size(GTK_WINDOW(win), 440, 260);
	if (parent != NULL) {
		gtk_window_set_transient_for(GTK_WINDOW(win), GTK_WINDOW(parent));
	}

	GtkWidget *box = gtk_box_new(GTK_ORIENTATION_VERTICAL, 12);
	gtk_widget_set_margin_top(box, 16);
	gtk_widget_set_margin_bottom(box, 16);
	gtk_widget_set_margin_start(box, 16);
	gtk_widget_set_margin_end(box, 16);

	GtkWidget *js = gtk_switch_new();
	gtk_switch_set_active(GTK_SWITCH(js), g_setting_javascript);
	gtk_widget_set_halign(js, GTK_ALIGN_END);
	g_signal_connect(js, "notify::active", G_CALLBACK(on_js_toggled), NULL);
	gtk_box_append(GTK_BOX(box), pref_row("Enable JavaScript", js));

	GtkWidget *gl = gtk_switch_new();
	gtk_switch_set_active(GTK_SWITCH(gl), g_setting_webgl);
	gtk_widget_set_halign(gl, GTK_ALIGN_END);
	g_signal_connect(gl, "notify::active", G_CALLBACK(on_webgl_toggled), NULL);
	gtk_box_append(GTK_BOX(box), pref_row("Enable WebGL", gl));

	GtkWidget *dev = gtk_switch_new();
	gtk_switch_set_active(GTK_SWITCH(dev), g_setting_devtools);
	gtk_widget_set_halign(dev, GTK_ALIGN_END);
	g_signal_connect(dev, "notify::active", G_CALLBACK(on_devtools_toggled), NULL);
	gtk_box_append(GTK_BOX(box), pref_row("Enable Developer Tools", dev));

	GtkWidget *zoom = gtk_spin_button_new_with_range(0.5, 3.0, 0.1);
	gtk_spin_button_set_value(GTK_SPIN_BUTTON(zoom), g_setting_zoom);
	g_signal_connect(zoom, "value-changed", G_CALLBACK(on_zoom_changed), NULL);
	gtk_box_append(GTK_BOX(box), pref_row("Default Zoom", zoom));

	gtk_window_set_child(GTK_WINDOW(win), box);
	gtk_window_present(GTK_WINDOW(win));
}
