package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func writeFirefoxContextTheme(p systemPreferences) {
	home, _ := os.UserHomeDir()
	surface := tintedColor(p.Accent, 255, .88)
	text := "#352737"
	if p.Dark {
		surface = tintedColor(p.Accent, 0, .76)
		text = "#fff0fb"
	}
	css := fmt.Sprintf(`/* MiracleOS context menus: preserves Firefox's native actions. */
menupopup { --panel-background-color: %s !important; --panel-text-color: %s !important; --panel-border-color: color-mix(in srgb,%s 30%%,transparent) !important; --panel-border-radius: 13px !important; --panel-padding: 7px !important; }
menupopup::part(content) { background: %s !important; color: %s !important; border-radius: 13px !important; padding: 7px !important; box-shadow: 0 8px 24px #22102244 !important; }
menupopup menuitem, menupopup menu { border-radius: 8px !important; padding-block: 8px !important; margin-inline: 2px !important; color: %s !important; }
menupopup menuitem[_moz-menuactive="true"],menupopup menu[_moz-menuactive="true"] { background-color: color-mix(in srgb,%s 24%%,transparent) !important; }
menupopup menuitem[disabled="true"] { opacity: .45 !important; }
menupopup menuseparator { margin: 5px 10px !important; }
`, surface, text, p.Accent, surface, text, text, p.Accent)
	for _, base := range []string{filepath.Join(home, ".config/mozilla/firefox"), filepath.Join(home, ".mozilla/firefox")} {
		profiles, _ := os.ReadDir(base)
		for _, profile := range profiles {
			if !profile.IsDir() {
				continue
			}
			folder := filepath.Join(base, profile.Name())
			if _, err := os.Stat(filepath.Join(folder, "prefs.js")); err != nil {
				continue
			}
			chrome := filepath.Join(folder, "chrome")
			_ = os.MkdirAll(chrome, 0755)
			_ = os.WriteFile(filepath.Join(chrome, "miracleos-context.css"), []byte(css), 0644)
			path := filepath.Join(chrome, "userChrome.css")
			data, _ := os.ReadFile(path)
			const include = "@import url(\"miracleos-context.css\");"
			if !strings.Contains(string(data), include) {
				_ = os.WriteFile(path, []byte(include+"\n"+string(data)), 0644)
			}
			path = filepath.Join(folder, "user.js")
			data, _ = os.ReadFile(path)
			s := string(data)
			for _, line := range []string{`user_pref("toolkit.legacyUserProfileCustomizations.stylesheets", true);`, `user_pref("widget.gtk.native-context-menus", false);`, `user_pref("widget.use-xdg-desktop-portal.file-picker", 1);`} {
				if !strings.Contains(s, line) {
					s += "\n" + line + "\n"
				}
			}
			_ = os.WriteFile(path, []byte(s), 0644)
		}
	}
}
