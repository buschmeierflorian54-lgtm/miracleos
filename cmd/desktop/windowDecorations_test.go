package main

import (
	"strings"
	"testing"
)

func TestServerDecorationsOptIn(t *testing.T) {
	t.Setenv("MIRACLEOS_SERVER_DECORATIONS", "")
	if serverWindowDecorations() {
		t.Fatal("stock compositor must keep fallback frame")
	}
	t.Setenv("MIRACLEOS_SERVER_DECORATIONS", "1")
	if !serverWindowDecorations() {
		t.Fatal("patched session should use server frame")
	}
}
func TestWindowThemeUsesSystemPalette(t *testing.T) {
	p := systemPreferences{Accent: "#329df0", FontSize: 12}
	light := windowThemeConfiguration(p)
	p.Dark = true
	dark := windowThemeConfiguration(p)
	if light == dark {
		t.Fatal("dark palette unchanged")
	}
	for _, css := range []string{light, dark} {
		if !strings.Contains(css, "#329df0") || !strings.Contains(css, "client.focused") || !strings.Contains(css, "titlebar_padding") {
			t.Fatal("missing shared window theme")
		}
	}
}
