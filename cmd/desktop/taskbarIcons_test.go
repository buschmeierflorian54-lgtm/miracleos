package main

import (
	"strings"
	"testing"
)

func TestTaskbarAppIcon(t *testing.T) {
	apps := []installedApp{{ID: "firefox-esr.desktop", Icon: "firefox-esr"}, {ID: "steam.desktop", WMClass: "Steam", Icon: "steam"}}
	for _, tc := range []struct{ id, title, want string }{
		{"firefox-esr", "Browser", "firefox-esr"},
		{"Steam", "Steam", "steam"},
		{"miracleos-desktop", "Terminal — MiracleOS", "utilities-terminal"},
		{"miracleos-desktop", "Dateien — MiracleOS", "files.png"},
		{"miracleos-desktop", "Einstellungen — MiracleOS", "settings.png"},
		{"unknown", "Unknown", "application-x-executable"},
	} {
		if got := taskbarAppIcon(tc.id, tc.title, apps); !strings.HasSuffix(got, tc.want) {
			t.Errorf("%s: got %s want %s", tc.title, got, tc.want)
		}
	}
}
