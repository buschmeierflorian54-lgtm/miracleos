package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// The session sets this only when the patched compositor supplies the controls.
// Stock Sway retains the existing GTK frame as a fallback.
func serverWindowDecorations() bool { return os.Getenv("MIRACLEOS_SERVER_DECORATIONS") == "1" }
func windowFrameSurface(p systemPreferences) string {
	if p.FrameColor != "" {
		if p.Dark {
			return frameSwatch(p.FrameColor, true)
		}
		return p.FrameColor
	}
	if p.Dark {
		return tintedColor(p.Accent, 0, .78)
	}
	return tintedColor(p.Accent, 255, .72)
}
func windowClientSurface(p systemPreferences) string {
	if p.FrameColor == "#000000" {
		p.FrameColor = ""
	}
	return windowFrameSurface(p)
}
func windowThemeConfiguration(p systemPreferences) string {
	surface := windowFrameSurface(p)
	text := "#352737"
	if p.Dark || p.FrameColor == "#000000" {
		text = "#fff0fb"
	}
	return fmt.Sprintf("# Generated from MiracleOS system preferences.\nfont pango:DejaVu Sans %d\ntitlebar_padding 12 10\ntitlebar_border_thickness 0\ndefault_border normal 6\ndefault_floating_border normal 6\nclient.focused %s %s %s %s %s\nclient.unfocused %s %s %s %s %s\n", p.FontSize, surface, surface, text, p.Accent, surface, surface, surface, text, p.Accent, surface)
}
func writeWindowTheme(p systemPreferences) error {
	path := filepath.Join(filepath.Dir(preferencesPath()), "window-theme.conf")
	if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		return e
	}
	if e := writeAppFrames(p); e != nil {
		return e
	}
	data := []byte(windowThemeConfiguration(p))
	if old, e := os.ReadFile(path); e == nil && string(old) == string(data) {
		return nil
	}
	tmp := path + ".tmp"
	if e := os.WriteFile(tmp, data, 0600); e != nil {
		return e
	}
	return os.Rename(tmp, path)
}
