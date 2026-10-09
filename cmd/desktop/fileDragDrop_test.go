package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileDropPaths(t *testing.T) {
	paths := fileDropPaths([]string{"file:///tmp/a%20b.png", "https://example.org/photo.png", "file://remote/tmp/a", "file:///tmp/a%20b.png"})
	if len(paths) != 1 || paths[0] != "/tmp/a b.png" {
		t.Fatal(paths)
	}
}

func TestFileDropTransfer(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	dest := filepath.Join(root, "dest")
	if err := os.MkdirAll(source, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		t.Fatal(err)
	}
	image := filepath.Join(source, "image.png")
	if err := os.WriteFile(image, []byte("image"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := pasteFilePaths([]string{image}, false, dest); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(image); err != nil {
		t.Fatal("copy removed source", err)
	}
	moved, err := pasteFilePaths([]string{image}, true, dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(moved) != 1 || filepath.Base(moved[0]) != "image (Kopie).png" {
		t.Fatal(moved)
	}
	if _, err := os.Stat(image); !os.IsNotExist(err) {
		t.Fatal("move retained source", err)
	}
	if _, err := pasteFilePaths([]string{dest}, false, dest); err == nil {
		t.Fatal("folder copied into itself")
	}
}
