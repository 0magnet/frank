/* Frank — menu bar (File/View/History/Tools/Help) + application actions. */
#include "browser.h"

static void act_reload(GSimpleAction *a, GVariant *p, gpointer app) {
	if (g_webview) webkit_web_view_reload(g_webview);
}
static void act_back(GSimpleAction *a, GVariant *p, gpointer app) {
	if (g_webview && webkit_web_view_can_go_back(g_webview)) webkit_web_view_go_back(g_webview);
}
static void act_forward(GSimpleAction *a, GVariant *p, gpointer app) {
	if (g_webview && webkit_web_view_can_go_forward(g_webview)) webkit_web_view_go_forward(g_webview);
}
static void act_zoom_in(GSimpleAction *a, GVariant *p, gpointer app) {
	if (g_webview) webkit_web_view_set_zoom_level(g_webview, webkit_web_view_get_zoom_level(g_webview) * 1.1);
}
static void act_zoom_out(GSimpleAction *a, GVariant *p, gpointer app) {
	if (g_webview) webkit_web_view_set_zoom_level(g_webview, webkit_web_view_get_zoom_level(g_webview) / 1.1);
}
static void act_zoom_reset(GSimpleAction *a, GVariant *p, gpointer app) {
	if (g_webview) webkit_web_view_set_zoom_level(g_webview, 1.0);
}
static void act_home(GSimpleAction *a, GVariant *p, gpointer app) {
	frank_load_home();
}
static void act_focus_address(GSimpleAction *a, GVariant *p, gpointer app) {
	if (g_url_entry) gtk_widget_grab_focus(g_url_entry);
}
static void act_new_window(GSimpleAction *a, GVariant *p, gpointer app) {
	char *argv[] = { (char *)"/proc/self/exe", NULL };
	g_spawn_async(NULL, argv, NULL, G_SPAWN_DEFAULT, NULL, NULL, NULL, NULL);
}
static void act_quit(GSimpleAction *a, GVariant *p, gpointer app) {
	g_application_quit(G_APPLICATION(app));
}
static void act_preferences(GSimpleAction *a, GVariant *p, gpointer app) {
	frank_show_preferences(g_main_window);
}
static void act_about(GSimpleAction *a, GVariant *p, gpointer app) {
	GtkWidget *d = gtk_about_dialog_new();
	gtk_about_dialog_set_program_name(GTK_ABOUT_DIALOG(d), "Frank");
	gtk_about_dialog_set_comments(GTK_ABOUT_DIALOG(d),
		"A native web browser on WebKitGTK-6.0, with Skywire skynet integration.");
	gtk_about_dialog_set_website(GTK_ABOUT_DIALOG(d), "https://github.com/0magnet/frank");
	gtk_about_dialog_set_license_type(GTK_ABOUT_DIALOG(d), GTK_LICENSE_MIT_X11);
	if (g_main_window) gtk_window_set_transient_for(GTK_WINDOW(d), GTK_WINDOW(g_main_window));
	gtk_window_present(GTK_WINDOW(d));
}

static const GActionEntry app_actions[] = {
	{ "reload", act_reload, NULL, NULL, NULL },
	{ "back", act_back, NULL, NULL, NULL },
	{ "forward", act_forward, NULL, NULL, NULL },
	{ "zoom-in", act_zoom_in, NULL, NULL, NULL },
	{ "zoom-out", act_zoom_out, NULL, NULL, NULL },
	{ "zoom-reset", act_zoom_reset, NULL, NULL, NULL },
	{ "home", act_home, NULL, NULL, NULL },
	{ "focus-address", act_focus_address, NULL, NULL, NULL },
	{ "new-window", act_new_window, NULL, NULL, NULL },
	{ "preferences", act_preferences, NULL, NULL, NULL },
	{ "about", act_about, NULL, NULL, NULL },
	{ "quit", act_quit, NULL, NULL, NULL },
};

static void set_accel(GtkApplication *app, const char *action, const char *accel) {
	const char *accels[] = { accel, NULL };
	gtk_application_set_accels_for_action(app, action, accels);
}

void frank_install_actions(GtkApplication *app) {
	g_action_map_add_action_entries(G_ACTION_MAP(app), app_actions, G_N_ELEMENTS(app_actions), app);
	set_accel(app, "app.reload", "<Control>r");
	set_accel(app, "app.zoom-in", "<Control>plus");
	set_accel(app, "app.zoom-out", "<Control>minus");
	set_accel(app, "app.zoom-reset", "<Control>0");
	set_accel(app, "app.home", "<Alt>Home");
	set_accel(app, "app.focus-address", "<Control>l");
	set_accel(app, "app.new-window", "<Control>n");
	set_accel(app, "app.preferences", "<Control>comma");
	set_accel(app, "app.quit", "<Control>q");
}

static void add_item(GMenu *m, const char *label, const char *action) {
	g_menu_append(m, label, action);
}

GtkWidget *frank_make_menubar(void) {
	GMenu *bar = g_menu_new();

	GMenu *file = g_menu_new();
	add_item(file, "New Window", "app.new-window");
	add_item(file, "Quit", "app.quit");
	g_menu_append_submenu(bar, "_File", G_MENU_MODEL(file));
	g_object_unref(file);

	GMenu *view = g_menu_new();
	add_item(view, "Home", "app.home");
	add_item(view, "Reload", "app.reload");
	add_item(view, "Zoom In", "app.zoom-in");
	add_item(view, "Zoom Out", "app.zoom-out");
	add_item(view, "Reset Zoom", "app.zoom-reset");
	g_menu_append_submenu(bar, "_View", G_MENU_MODEL(view));
	g_object_unref(view);

	GMenu *history = g_menu_new();
	add_item(history, "Back", "app.back");
	add_item(history, "Forward", "app.forward");
	g_menu_append_submenu(bar, "_History", G_MENU_MODEL(history));
	g_object_unref(history);

	GMenu *tools = g_menu_new();
	add_item(tools, "Preferences", "app.preferences");
	g_menu_append_submenu(bar, "_Tools", G_MENU_MODEL(tools));
	g_object_unref(tools);

	GMenu *help = g_menu_new();
	add_item(help, "About Frank", "app.about");
	g_menu_append_submenu(bar, "_Help", G_MENU_MODEL(help));
	g_object_unref(help);

	GtkWidget *w = gtk_popover_menu_bar_new_from_model(G_MENU_MODEL(bar));
	g_object_unref(bar);
	return w;
}
