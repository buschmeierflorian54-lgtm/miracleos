package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClipboardTreeCopiesWithoutFollowingLinks(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	if err := os.Mkdir(source, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "photo.png"), []byte("image"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("photo.png", filepath.Join(source, "link")); err != nil {
		t.Fatal(err)
	}
	snapshot, err := snapshotClipboardSource(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := copyClipboardTree(source, target, snapshot); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(target, "photo.png"))
	if err != nil || string(data) != "image" {
		t.Fatalf("copy: %s %v", data, err)
	}
	link, err := os.Readlink(filepath.Join(target, "link"))
	if err != nil || link != "photo.png" {
		t.Fatalf("symlink: %s %v", link, err)
	}
	if err := copyClipboardTree(source, target, snapshot); !os.IsExist(err) {
		t.Fatalf("existing target overwritten: %v", err)
	}
}

func TestClipboardDetectsChangedSourceBeforeRemoval(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	if err := os.Mkdir(source, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(source, "data")
	if err := os.WriteFile(path, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	snapshot, err := snapshotClipboardSource(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := copyClipboardTree(source, target, snapshot); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("changed source"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := unchangedClipboardSource(source, snapshot); err == nil {
		t.Fatal("source change was ignored")
	}
	if err := removeClipboardSource(source, snapshot); err == nil {
		t.Fatal("changed source was removed")
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "changed source" {
		t.Fatal("changed source not preserved")
	}
}

func TestClipboardIntegration(t *testing.T) {
	if os.Getenv("MIRACLEOS_CLIPBOARD_TEST") != "1" {
		t.Skip("requires live Wayland clipboard")
	}
	root := t.TempDir()
	source := filepath.Join(root, "source")
	dest := filepath.Join(root, "dest")
	os.Mkdir(source, 0755)
	os.Mkdir(dest, 0755)
	os.WriteFile(filepath.Join(source, "photo.png"), []byte("picture"), 0644)
	if err := copyFilesToClipboard([]string{source}, false); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := pasteClipboardFiles(dest); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"source", "source (Kopie)"} {
		data, err := os.ReadFile(filepath.Join(dest, name, "photo.png"))
		if err != nil || string(data) != "picture" {
			t.Fatal(name, err)
		}
	}
	if err := copyFilesToClipboard([]string{source}, true); err != nil {
		t.Fatal(err)
	}
	if err := pasteClipboardFiles(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(source); err != nil {
		t.Fatal("same folder cut", err)
	}
	if err := copyFilesToClipboard([]string{source}, true); err != nil {
		t.Fatal(err)
	}
	if err := pasteClipboardFiles(dest); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatal("cut source remains", err)
	}
}
