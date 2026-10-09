package main

import (
	"path/filepath"
	"strings"
)

const wallpaperSystemDir = "/usr/share/backgrounds/miracleos"

func defaultWallpaper(dark bool) string {
	if dark {
		return filepath.Join(wallpaperSystemDir, "dark.png")
	}
	return filepath.Join(wallpaperSystemDir, "light.png")
}
func isDefaultWallpaper(path string) bool {
	return path == "" || path == defaultWallpaper(false) || path == defaultWallpaper(true) || strings.HasSuffix(path, "/Pictures/MiracleOS.png") || strings.HasSuffix(path, "/miracleos-wallpaper.png")
}
func modeWallpaper(p systemPreferences) systemPreferences {
	if !p.CustomWallpaper && !isDefaultWallpaper(p.Wallpaper) {
		p.CustomWallpaper = true
	}
	if !p.CustomWallpaper {
		p.Wallpaper = defaultWallpaper(p.Dark)
	}
	return p
}
