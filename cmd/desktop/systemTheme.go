package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type systemPreferences struct {
	AppFrames       map[string]appFramePreference `json:"app_frames,omitempty"`
	FrameColor      string                        `json:"frame_color,omitempty"`
	Accent          string                        `json:"accent"`
	Dark            bool                          `json:"dark"`
	Transparency    bool                          `json:"transparency"`
	FontSize        int                           `json:"font_size"`
	CustomWallpaper bool                          `json:"custom_wallpaper,omitempty"`
	Wallpaper       string                        `json:"wallpaper,omitempty"`
}

func preferencesPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "miracleos", "settings.json")
}
func readSystemPreferences() systemPreferences {
	p := systemPreferences{Accent: "#bf69a4", Transparency: true, FontSize: 12}
	if data, err := os.ReadFile(preferencesPath()); err == nil {
		_ = json.Unmarshal(data, &p)
	}
	if !regexp.MustCompile(`^#[0-9a-fA-F]{6}$`).MatchString(p.Accent) {
		p.Accent = "#bf69a4"
	}
	if p.FontSize < 10 || p.FontSize > 16 {
		p.FontSize = 12
	}
	if !regexp.MustCompile(`^#[0-9a-fA-F]{6}$`).MatchString(p.FrameColor) {
		p.FrameColor = ""
	}
	return modeWallpaper(p)
}
func tintedColor(hex string, target int, amount float64) string {
	n, _ := strconv.ParseUint(strings.TrimPrefix(hex, "#"), 16, 32)
	r := int(float64((n>>16)&255)*(1-amount) + float64(target)*amount)
	g := int(float64((n>>8)&255)*(1-amount) + float64(target)*amount)
	b := int(float64(n&255)*(1-amount) + float64(target)*amount)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func systemThemeCSS(p systemPreferences) string {
	surface := tintedColor(p.Accent, 255, .72)
	card := tintedColor(p.Accent, 255, .88)
	text := "#352737"
	muted := "#826b80"
	opacity := "0.96"
	if p.Dark {
		surface = tintedColor(p.Accent, 0, .78)
		card = tintedColor(p.Accent, 0, .64)
		text = "#fff0fb"
		muted = "#c9b5c6"
	}
	if !p.Transparency {
		opacity = "1"
	}
	css := fmt.Sprintf(`
@define-color system_accent %s;
@define-color system_surface %s;
@define-color system_card %s;
@define-color system_text %s;
@define-color system_muted %s;
.fm-shell, .settings-shell { background-image: linear-gradient(135deg, alpha(@system_surface,%s), alpha(@system_card,%s)); color: @system_text; font-size: %dpx; }
.glass-dialog .fm-dialog { background: transparent; }
.glass-dialog .dialog-action {background-color: alpha(@system_card,.7);color:@system_text;}
.glass-dialog .dialog-primary {background-image:linear-gradient(140deg,@system_accent,shade(@system_accent,.82));color:white;}
.glass-dialog .dialog-primary label {color:white;}
.fm-shell label, .settings-shell label, .glass-titlebar label { color: @system_text; }
.fm-explorer .fm-tag-0 .fm-tag-dot { color: #ef5865; }
.fm-explorer .fm-tag-1 .fm-tag-dot { color: #329df0; }
.fm-explorer .fm-tag-2 .fm-tag-dot { color: #f4bf30; }
.fm-explorer .fm-tag-3 .fm-tag-dot { color: #36ba77; }
.fm-explorer .fm-tag-4 .fm-tag-dot { color: #ba56e5; }
.fm-explorer {background-color:alpha(@system_surface,.97); background-image:linear-gradient(135deg,alpha(@system_surface,.97),alpha(@system_card,.97));}
.fm-explorer .fm-content { background-color: alpha(@system_card,.94); border-color:alpha(@system_text,.12); }
.fm-explorer .fm-home-card {background-color:alpha(@system_surface,.5);border-color:alpha(@system_text,.12);}
.fm-explorer .fm-place, .fm-explorer .fm-control, .fm-explorer image {color:@system_text;}
.fm-explorer .fm-place image {opacity:.9;}
.fm-explorer .fm-status, .fm-explorer .fm-muted {color:@system_muted;}
.fm-explorer .fm-place:hover {background-color:alpha(@system_accent,.18);}
.fm-explorer .fm-place-active {background-color:alpha(@system_accent,.18);}
.fm-shell .fm-muted, .fm-shell .fm-section-label, .settings-shell .settings-muted { color: @system_muted; }
.glass-titlebar { background-image: linear-gradient(to bottom, alpha(@system_accent,.32),alpha(@system_accent,.18)); }
.glass-window-control, .fm-control, .fm-place-label { color: @system_text; }
.fm-home-card, .fm-content, .installer-card, .settings-card { background-color: alpha(@system_card,.60); background-image: none; }
.fm-path, .fm-search, .settings-search, .settings-shell entry, .settings-shell combobox button { background-color: alpha(@system_card,.78); color: @system_text; }
.settings-selected, .fm-place-active { box-shadow: inset 3px 0 @system_accent; background-color: alpha(@system_card,.90); }
.settings-shell .settings-sidebar button { background-image: none; background-color: transparent; border-color: transparent; box-shadow: none; }
.settings-shell .settings-sidebar .settings-selected { background-color: transparent; box-shadow: inset 3px 0 @system_accent; }
.settings-shell .settings-sidebar button:hover { background-color: alpha(@system_card,.65); }
.dialog-primary, .settings-primary, .installer-primary { background-image: linear-gradient(140deg,@system_accent,shade(@system_accent,.82)); color: white; }
.dialog-primary label, .settings-primary label, .installer-primary label { color: white; }
.settings-shell switch:checked, .settings-shell scale highlight, .installer-progress progress { background-color: @system_accent; background-image: none; }
.settings-shell switch slider { background-color: @system_card; }
.settings-shell button:hover { background-color: alpha(@system_card,.90); }
.settings-shell button.settings-primary:hover { background-color: @system_accent; }
.settings-shell .settings-theme-sample-light label { color: #493247; }
.settings-shell .settings-theme-sample-dark label { color: #fff1fa; }
.mm_menu_m2 { background-image: linear-gradient(140deg,alpha(@system_surface,.97),alpha(@system_surface,.94)); color: @system_text; }
.mm_menu_m2 label { color: @system_text; }
.mm_menu_m2 .mm_tab { background-color: alpha(@system_card,.35); background-image: none; }
.mm_menu_m2 entry { background-color: alpha(@system_card,.65); color: @system_text; }
.startup-screen { background-image: linear-gradient(135deg,shade(@system_accent,.35),shade(@system_accent,.78)); }
.glass-context-menu { background-image: none; background-color: alpha(@system_card,.98); color: @system_text; }
.glass-context-menu label { color: @system_text; }
.glass-context-menu menuitem:hover { background-color: alpha(@system_accent,.22); }
`, p.Accent, surface, card, text, muted, opacity, opacity, p.FontSize)
	if p.Dark {
		css += " .fm-explorer {background-color:@system_surface;background-image:linear-gradient(135deg,@system_surface,@system_card);} .fm-explorer .fm-content {background-color:@system_card;}"
	}
	css += fmt.Sprintf(` .miracleos-server-window .fm-shell, .miracleos-server-window .fm-explorer, .miracleos-server-window .settings-shell { background-image:none; background-color:%s; border:0; border-radius:0; box-shadow:none; }`, windowClientSurface(appClientPreferences(p)))
	return css
}
func watchSystemTheme() {
	provider, _ := gtk.CssProviderNew()
	screen, _ := gdk.ScreenGetDefault()
	gtk.AddProviderForScreen(screen, provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION+1)
	last := ""
	update := func() bool {
		data := systemThemeCSS(readSystemPreferences())
		if data != last {
			if provider.LoadFromData(data) == nil {
				last = data
				_ = writeWindowTheme(readSystemPreferences())
				writeFirefoxContextTheme(readSystemPreferences())
				writeLoginAppearance(readSystemPreferences())
			}
		}
		return true
	}
	update()
	glib.TimeoutAdd(600, update)
}

func saveSystemPreferences(p systemPreferences) error {
	p = modeWallpaper(p)
	dir := filepath.Dir(preferencesPath())
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	temp := preferencesPath() + ".tmp"
	if err = os.WriteFile(temp, data, 0600); err != nil {
		return err
	}
	if err = os.Rename(temp, preferencesPath()); err != nil {
		return err
	}
	_ = writeWindowTheme(p)
	writeLoginAppearance(p)
	if serverWindowDecorations() {
		go quickCommand("swaymsg", "reload")
	}
	// Native GTK dialogs use the same palette. Keep active authentication intact.
	home, _ := os.UserHomeDir()
	gtkDir := filepath.Join(home, ".config", "gtk-3.0")
	_ = os.MkdirAll(gtkDir, 0755)
	surface := tintedColor(p.Accent, 255, .87)
	text := "#422a43"
	if p.Dark {
		surface = tintedColor(p.Accent, 0, .75)
		text = "#fff0fb"
	}
	native := fmt.Sprintf(`.miracleos-auth { background-image:none; background-color:%s; color:%s; }
.miracleos-auth label { color:%s; }
.miracleos-auth entry, .miracleos-auth combobox button { background-image:none; background-color:%s; color:%s; }
.miracleos-auth button.default { background-image:none; background-color:%s; border-color:%s; }
.miracleos-auth button.default label { color:white; text-shadow:none; }
headerbar, .titlebar { background-image:none; background-color:%s; color:%s; }
`, surface, text, text, tintedColor(surface, 255, .12), text, p.Accent, p.Accent, surface, text)
	_ = os.WriteFile(filepath.Join(gtkDir, "miracleos-theme.css"), []byte(native), 0644)
	userCSS := filepath.Join(gtkDir, "gtk.css")
	css, _ := os.ReadFile(userCSS)
	const marker = "/* MiracleOS selected system palette */"
	base := strings.Split(string(css), marker)[0]
	_ = os.WriteFile(userCSS, []byte(base+"\n"+marker+"\n"+native), 0644)
	go func() {
		if exec.Command("pgrep", "-x", "pkexec").Run() != nil {
			_ = exec.Command("pkill", "-x", "lxpolkit").Run()
			_ = exec.Command("swaymsg", "exec", "env GTK_CSD=1 /usr/bin/lxpolkit").Run()
		}
	}()
	return nil
}

// Publish only non-sensitive appearance settings; the greeter cannot read homes.
func writeLoginAppearance(p systemPreferences) {
	account, err := user.Current()
	if err != nil { return }
	dir := filepath.Join("/var/lib/miracleos/appearance", filepath.Base(account.Username))
	data, _ := json.Marshal(struct { Dark bool `json:"dark"`; Accent string `json:"accent"` }{p.Dark, p.Accent})
	if os.WriteFile(filepath.Join(dir, "settings.tmp"), data, 0644) == nil {
		_ = os.Rename(filepath.Join(dir, "settings.tmp"), filepath.Join(dir, "settings.json"))
	}
}
