/* Frank — preferences, persisted to ~/.config/frank/frank.ini via GKeyFile. */
#include "browser.h"

/* General */
static char *s_homepage = NULL; /* "" / NULL = built-in welcome */
static gdouble s_zoom = 1.0;
static int s_font_size = 16;
static int s_min_font_size = 0;
/* Content */
static gboolean s_javascript = TRUE;
static gboolean s_images = TRUE;
static gboolean s_webgl = TRUE;
static gboolean s_webaudio = TRUE;
static gboolean s_smooth_scroll = TRUE;
static gboolean s_block_popups = TRUE;
static gboolean s_allow_fullscreen = TRUE;
/* Privacy */
static int s_cookie_policy = 1; /* 0 always, 1 no-third-party, 2 never */
static gboolean s_itp = TRUE;
static gboolean s_local_storage = TRUE;
/* Advanced */
static gboolean s_devtools = FALSE;
static gboolean s_hw_accel = TRUE; /* always vs never */
static char *s_user_agent = NULL;  /* "" / NULL = engine default */

const char *frank_homepage(void) { return s_homepage ? s_homepage : ""; }

static char *config_path(void) {
	char *dir = g_build_filename(g_get_user_config_dir(), "frank", NULL);
	g_mkdir_with_parents(dir, 0700);
	char *path = g_build_filename(dir, "frank.ini", NULL);
	g_free(dir);
	return path;
}

static gboolean kf_bool(GKeyFile *kf, const char *g, const char *k, gboolean def) {
	GError *e = NULL;
	gboolean v = g_key_file_get_boolean(kf, g, k, &e);
	if (e) { g_error_free(e); return def; }
	return v;
}
static int kf_int(GKeyFile *kf, const char *g, const char *k, int def) {
	GError *e = NULL;
	int v = g_key_file_get_integer(kf, g, k, &e);
	if (e) { g_error_free(e); return def; }
	return v;
}
static gdouble kf_double(GKeyFile *kf, const char *g, const char *k, gdouble def) {
	GError *e = NULL;
	gdouble v = g_key_file_get_double(kf, g, k, &e);
	if (e) { g_error_free(e); return def; }
	return v;
}
static char *kf_str(GKeyFile *kf, const char *g, const char *k) {
	return g_key_file_get_string(kf, g, k, NULL); /* NULL if absent */
}

void frank_settings_load(void) {
	char *path = config_path();
	GKeyFile *kf = g_key_file_new();
	if (g_key_file_load_from_file(kf, path, G_KEY_FILE_NONE, NULL)) {
		char *hp = kf_str(kf, "general", "homepage"); if (hp) { g_free(s_homepage); s_homepage = hp; }
		s_zoom = kf_double(kf, "general", "zoom", s_zoom); if (s_zoom <= 0) s_zoom = 1.0;
		s_font_size = kf_int(kf, "general", "font_size", s_font_size);
		s_min_font_size = kf_int(kf, "general", "min_font_size", s_min_font_size);

		s_javascript = kf_bool(kf, "content", "javascript", s_javascript);
		s_images = kf_bool(kf, "content", "images", s_images);
		s_webgl = kf_bool(kf, "content", "webgl", s_webgl);
		s_webaudio = kf_bool(kf, "content", "webaudio", s_webaudio);
		s_smooth_scroll = kf_bool(kf, "content", "smooth_scroll", s_smooth_scroll);
		s_block_popups = kf_bool(kf, "content", "block_popups", s_block_popups);
		s_allow_fullscreen = kf_bool(kf, "content", "allow_fullscreen", s_allow_fullscreen);

		s_cookie_policy = kf_int(kf, "privacy", "cookie_policy", s_cookie_policy);
		s_itp = kf_bool(kf, "privacy", "tracking_prevention", s_itp);
		s_local_storage = kf_bool(kf, "privacy", "local_storage", s_local_storage);

		s_devtools = kf_bool(kf, "advanced", "devtools", s_devtools);
		s_hw_accel = kf_bool(kf, "advanced", "hardware_acceleration", s_hw_accel);
		char *ua = kf_str(kf, "advanced", "user_agent"); if (ua) { g_free(s_user_agent); s_user_agent = ua; }
	}
	g_key_file_free(kf);
	g_free(path);
}

void frank_settings_save(void) {
	char *path = config_path();
	GKeyFile *kf = g_key_file_new();
	g_key_file_set_string(kf, "general", "homepage", s_homepage ? s_homepage : "");
	g_key_file_set_double(kf, "general", "zoom", s_zoom);
	g_key_file_set_integer(kf, "general", "font_size", s_font_size);
	g_key_file_set_integer(kf, "general", "min_font_size", s_min_font_size);
	g_key_file_set_boolean(kf, "content", "javascript", s_javascript);
	g_key_file_set_boolean(kf, "content", "images", s_images);
	g_key_file_set_boolean(kf, "content", "webgl", s_webgl);
	g_key_file_set_boolean(kf, "content", "webaudio", s_webaudio);
	g_key_file_set_boolean(kf, "content", "smooth_scroll", s_smooth_scroll);
	g_key_file_set_boolean(kf, "content", "block_popups", s_block_popups);
	g_key_file_set_boolean(kf, "content", "allow_fullscreen", s_allow_fullscreen);
	g_key_file_set_integer(kf, "privacy", "cookie_policy", s_cookie_policy);
	g_key_file_set_boolean(kf, "privacy", "tracking_prevention", s_itp);
	g_key_file_set_boolean(kf, "privacy", "local_storage", s_local_storage);
	g_key_file_set_boolean(kf, "advanced", "devtools", s_devtools);
	g_key_file_set_boolean(kf, "advanced", "hardware_acceleration", s_hw_accel);
	g_key_file_set_string(kf, "advanced", "user_agent", s_user_agent ? s_user_agent : "");
	g_key_file_save_to_file(kf, path, NULL);
	g_key_file_free(kf);
	g_free(path);
}

void frank_settings_apply(WebKitWebView *wv) {
	if (wv == NULL) return;
	WebKitSettings *s = webkit_web_view_get_settings(wv);
	webkit_settings_set_enable_javascript(s, s_javascript);
	webkit_settings_set_auto_load_images(s, s_images);
	webkit_settings_set_enable_webgl(s, s_webgl);
	webkit_settings_set_enable_webaudio(s, s_webaudio);
	webkit_settings_set_enable_smooth_scrolling(s, s_smooth_scroll);
	webkit_settings_set_javascript_can_open_windows_automatically(s, !s_block_popups);
	webkit_settings_set_enable_fullscreen(s, s_allow_fullscreen);
	webkit_settings_set_enable_html5_local_storage(s, s_local_storage);
	webkit_settings_set_enable_developer_extras(s, s_devtools);
	webkit_settings_set_default_font_size(s, s_font_size);
	webkit_settings_set_minimum_font_size(s, s_min_font_size);
	webkit_settings_set_hardware_acceleration_policy(s,
		s_hw_accel ? WEBKIT_HARDWARE_ACCELERATION_POLICY_ALWAYS
		           : WEBKIT_HARDWARE_ACCELERATION_POLICY_NEVER);
	if (s_user_agent && s_user_agent[0])
		webkit_settings_set_user_agent(s, s_user_agent);
	webkit_web_view_set_zoom_level(wv, s_zoom);

	WebKitNetworkSession *ns = webkit_web_view_get_network_session(wv);
	if (ns) {
		WebKitCookieManager *cm = webkit_network_session_get_cookie_manager(ns);
		WebKitCookieAcceptPolicy cp = WEBKIT_COOKIE_POLICY_ACCEPT_NO_THIRD_PARTY;
		if (s_cookie_policy == 0) cp = WEBKIT_COOKIE_POLICY_ACCEPT_ALWAYS;
		else if (s_cookie_policy == 2) cp = WEBKIT_COOKIE_POLICY_ACCEPT_NEVER;
		webkit_cookie_manager_set_accept_policy(cm, cp);
		webkit_network_session_set_itp_enabled(ns, s_itp);
	}
}

/* ── preferences dialog ─────────────────────────────────────────────────── */

static void changed(void) {
	frank_settings_apply(g_webview);
	frank_settings_save();
}

static GtkWidget *row(const char *label, GtkWidget *control) {
	GtkWidget *box = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 12);
	GtkWidget *l = gtk_label_new(label);
	gtk_widget_set_hexpand(l, TRUE);
	gtk_widget_set_halign(l, GTK_ALIGN_START);
	gtk_widget_set_valign(control, GTK_ALIGN_CENTER);
	gtk_box_append(GTK_BOX(box), l);
	gtk_box_append(GTK_BOX(box), control);
	return box;
}

static void on_bool(GObject *sw, GParamSpec *ps, gpointer ptr) {
	*(gboolean *)ptr = gtk_switch_get_active(GTK_SWITCH(sw));
	changed();
}
static GtkWidget *bool_row(const char *label, gboolean *val) {
	GtkWidget *sw = gtk_switch_new();
	gtk_switch_set_active(GTK_SWITCH(sw), *val);
	gtk_widget_set_halign(sw, GTK_ALIGN_END);
	g_signal_connect(sw, "notify::active", G_CALLBACK(on_bool), val);
	return row(label, sw);
}
static void on_int(GtkSpinButton *sb, gpointer ptr) {
	*(int *)ptr = gtk_spin_button_get_value_as_int(sb);
	changed();
}
static GtkWidget *int_row(const char *label, int *val, int lo, int hi) {
	GtkWidget *sb = gtk_spin_button_new_with_range(lo, hi, 1);
	gtk_spin_button_set_value(GTK_SPIN_BUTTON(sb), *val);
	g_signal_connect(sb, "value-changed", G_CALLBACK(on_int), val);
	return row(label, sb);
}
static void on_dbl(GtkSpinButton *sb, gpointer ptr) {
	*(gdouble *)ptr = gtk_spin_button_get_value(sb);
	changed();
}
static GtkWidget *dbl_row(const char *label, gdouble *val, double lo, double hi, double step) {
	GtkWidget *sb = gtk_spin_button_new_with_range(lo, hi, step);
	gtk_spin_button_set_value(GTK_SPIN_BUTTON(sb), *val);
	g_signal_connect(sb, "value-changed", G_CALLBACK(on_dbl), val);
	return row(label, sb);
}
static void on_entry(GtkEditable *e, gpointer ptr) {
	char **val = (char **)ptr;
	g_free(*val);
	*val = g_strdup(gtk_editable_get_text(e));
	changed();
}
static GtkWidget *entry_row(const char *label, char **val, const char *placeholder) {
	GtkWidget *e = gtk_entry_new();
	gtk_widget_set_hexpand(e, TRUE);
	if (*val) gtk_editable_set_text(GTK_EDITABLE(e), *val);
	if (placeholder) gtk_entry_set_placeholder_text(GTK_ENTRY(e), placeholder);
	g_signal_connect(e, "changed", G_CALLBACK(on_entry), val);
	return row(label, e);
}
static void on_combo(GObject *dd, GParamSpec *ps, gpointer ptr) {
	*(int *)ptr = (int)gtk_drop_down_get_selected(GTK_DROP_DOWN(dd));
	changed();
}
static GtkWidget *combo_row(const char *label, int *val, const char *const *opts) {
	GtkWidget *dd = gtk_drop_down_new_from_strings(opts);
	gtk_drop_down_set_selected(GTK_DROP_DOWN(dd), (guint)*val);
	gtk_widget_set_halign(dd, GTK_ALIGN_END);
	g_signal_connect(dd, "notify::selected", G_CALLBACK(on_combo), val);
	return row(label, dd);
}

static void on_clear_cache(GtkButton *b, gpointer d) {
	if (!g_webview) return;
	WebKitNetworkSession *ns = webkit_web_view_get_network_session(g_webview);
	WebKitWebsiteDataManager *m = webkit_network_session_get_website_data_manager(ns);
	webkit_website_data_manager_clear(m, WEBKIT_WEBSITE_DATA_DISK_CACHE | WEBKIT_WEBSITE_DATA_MEMORY_CACHE, 0, NULL, NULL, NULL);
}
static void on_clear_cookies(GtkButton *b, gpointer d) {
	if (!g_webview) return;
	WebKitNetworkSession *ns = webkit_web_view_get_network_session(g_webview);
	WebKitWebsiteDataManager *m = webkit_network_session_get_website_data_manager(ns);
	webkit_website_data_manager_clear(m, WEBKIT_WEBSITE_DATA_COOKIES, 0, NULL, NULL, NULL);
}

static GtkWidget *page(void) {
	GtkWidget *b = gtk_box_new(GTK_ORIENTATION_VERTICAL, 10);
	gtk_widget_set_margin_top(b, 16);
	gtk_widget_set_margin_bottom(b, 16);
	gtk_widget_set_margin_start(b, 16);
	gtk_widget_set_margin_end(b, 16);
	return b;
}

void frank_show_preferences(GtkWidget *parent) {
	GtkWidget *win = gtk_window_new();
	gtk_window_set_title(GTK_WINDOW(win), "Frank Preferences");
	gtk_window_set_default_size(GTK_WINDOW(win), 520, 420);
	if (parent) gtk_window_set_transient_for(GTK_WINDOW(win), GTK_WINDOW(parent));

	GtkWidget *nb = gtk_notebook_new();

	/* General */
	GtkWidget *g = page();
	gtk_box_append(GTK_BOX(g), entry_row("Homepage", &s_homepage, "blank = built-in welcome page"));
	gtk_box_append(GTK_BOX(g), dbl_row("Default zoom", &s_zoom, 0.5, 3.0, 0.1));
	gtk_box_append(GTK_BOX(g), int_row("Default font size", &s_font_size, 6, 48));
	gtk_box_append(GTK_BOX(g), int_row("Minimum font size", &s_min_font_size, 0, 48));
	gtk_notebook_append_page(GTK_NOTEBOOK(nb), g, gtk_label_new("General"));

	/* Content */
	GtkWidget *c = page();
	gtk_box_append(GTK_BOX(c), bool_row("Enable JavaScript", &s_javascript));
	gtk_box_append(GTK_BOX(c), bool_row("Load images automatically", &s_images));
	gtk_box_append(GTK_BOX(c), bool_row("Enable WebGL", &s_webgl));
	gtk_box_append(GTK_BOX(c), bool_row("Enable Web Audio", &s_webaudio));
	gtk_box_append(GTK_BOX(c), bool_row("Smooth scrolling", &s_smooth_scroll));
	gtk_box_append(GTK_BOX(c), bool_row("Block pop-up windows", &s_block_popups));
	gtk_box_append(GTK_BOX(c), bool_row("Allow fullscreen", &s_allow_fullscreen));
	gtk_notebook_append_page(GTK_NOTEBOOK(nb), c, gtk_label_new("Content"));

	/* Privacy */
	GtkWidget *p = page();
	static const char *cookie_opts[] = { "Always accept", "Block third-party", "Never accept", NULL };
	gtk_box_append(GTK_BOX(p), combo_row("Cookies", &s_cookie_policy, cookie_opts));
	gtk_box_append(GTK_BOX(p), bool_row("Tracking prevention (ITP)", &s_itp));
	gtk_box_append(GTK_BOX(p), bool_row("Allow local storage", &s_local_storage));
	GtkWidget *clr = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 8);
	GtkWidget *bc = gtk_button_new_with_label("Clear cache");
	GtkWidget *bk = gtk_button_new_with_label("Clear cookies");
	g_signal_connect(bc, "clicked", G_CALLBACK(on_clear_cache), NULL);
	g_signal_connect(bk, "clicked", G_CALLBACK(on_clear_cookies), NULL);
	gtk_box_append(GTK_BOX(clr), bc);
	gtk_box_append(GTK_BOX(clr), bk);
	gtk_box_append(GTK_BOX(p), clr);
	gtk_notebook_append_page(GTK_NOTEBOOK(nb), p, gtk_label_new("Privacy"));

	/* Advanced */
	GtkWidget *a = page();
	gtk_box_append(GTK_BOX(a), bool_row("Enable Developer Tools", &s_devtools));
	gtk_box_append(GTK_BOX(a), bool_row("Hardware acceleration", &s_hw_accel));
	gtk_box_append(GTK_BOX(a), entry_row("User agent", &s_user_agent, "blank = engine default"));
	gtk_notebook_append_page(GTK_NOTEBOOK(nb), a, gtk_label_new("Advanced"));

	gtk_window_set_child(GTK_WINDOW(win), nb);
	gtk_window_present(GTK_WINDOW(win));
}
