package main

import (
	"github.com/gotk3/gotk3/gtk"
	"os"
	"path/filepath"
	"testing"
)

func TestSystemThemeVariantsParse(t *testing.T) {
	gtk.Init(nil)
	colors := []string{"#bf69a4", "#ec4779", "#3d9894", "#8183e7", "#ad62d7", "#e8ae29", "#ee9235", "#d35c45", "#dc3a53", "#ab3e79", "#2389cc", "#226fa8", "#36a7ba", "#2d9b79", "#378957"}
	for _, color := range colors {
		for _, dark := range []bool{false, true} {
			for _, transparent := range []bool{false, true} {
				provider, err := gtk.CssProviderNew()
				if err != nil {
					t.Fatal(err)
				}
				p := systemPreferences{Accent: color, Dark: dark, Transparency: transparent, FontSize: 14}
				if err := provider.LoadFromData(systemThemeCSS(p)); err != nil {
					t.Fatalf("Theme %s dark=%v transparent=%v: %v", color, dark, transparent, err)
				}
			}
		}
	}
}

func TestInvalidPreferencesHaveSafeDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := os.MkdirAll(filepath.Dir(preferencesPath()), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(preferencesPath(), []byte(`{"accent":"red; color: black","font_size":900}`), 0600); err != nil {
		t.Fatal(err)
	}
	p := readSystemPreferences()
	if p.Accent != "#bf69a4" || p.FontSize != 12 {
		t.Fatalf("Invalid preferences were not normalized: %+v", p)
	}
}
