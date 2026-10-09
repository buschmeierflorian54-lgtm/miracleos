package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstalledApplicationCatalog(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_DATA_HOME", root)
	t.Setenv("XDG_DATA_DIRS", root)
	dir := filepath.Join(root, "applications")
	os.MkdirAll(dir, 0755)
	entries := map[string]string{"visible.desktop": "Name=Catalog Visible\nExec=/bin/true\n", "hidden.desktop": "Name=Catalog Hidden\nExec=/bin/true\nHidden=true\n", "nodisplay.desktop": "Name=Catalog NoDisplay\nExec=/bin/true\nNoDisplay=true\n"}
	for name, body := range entries {
		os.WriteFile(filepath.Join(dir, name), []byte("[Desktop Entry]\nType=Application\n"+body), 0644)
	}
	apps := installedApplications()
	if len(apps) != 1 || apps[0].Name != "Catalog Visible" || apps[0].File != filepath.Join(dir, "visible.desktop") {
		t.Fatalf("catalog: %+v", apps)
	}
}
