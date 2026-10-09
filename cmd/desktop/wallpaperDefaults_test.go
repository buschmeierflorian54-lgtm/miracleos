package main

import "testing"

func TestWallpaperDefaults(t *testing.T) {
	for _, dark := range []bool{false, true} {
		p := modeWallpaper(systemPreferences{Dark: dark, Wallpaper: "/home/test/Pictures/MiracleOS.png"})
		if p.Wallpaper != defaultWallpaper(dark) || p.CustomWallpaper {
			t.Fatal(p)
		}
	}
	p := modeWallpaper(systemPreferences{Dark: true, Wallpaper: "/home/test/custom.jpg", CustomWallpaper: true})
	if p.Wallpaper != "/home/test/custom.jpg" {
		t.Fatal("custom wallpaper overwritten")
	}
	p = modeWallpaper(systemPreferences{Dark: true, Wallpaper: defaultWallpaper(false)})
	if p.Wallpaper != defaultWallpaper(true) {
		t.Fatal("mode switch failed")
	}
}
