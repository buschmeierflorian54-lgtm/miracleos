package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestFolderZIPContentsAndCollision(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "Pictures")
	if err := os.MkdirAll(filepath.Join(source, "empty"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "photo.png"), []byte("picture"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../outside", filepath.Join(source, "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "Pictures.zip"), []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	archive, err := createFolderZIP(source)
	if err != nil {
		t.Fatal(err)
	}
	if archive != filepath.Join(root, "Pictures (1).zip") {
		t.Fatal(archive)
	}
	reader, err := zip.OpenReader(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	contents := make(map[string]string)
	for _, file := range reader.File {
		stream, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(stream)
		stream.Close()
		if err != nil {
			t.Fatal(err)
		}
		contents[file.Name] = string(data)
		if file.Name == "Pictures/link" && file.Mode()&os.ModeSymlink == 0 {
			t.Fatal("link mode lost")
		}
	}
	if len(contents) != 4 || contents["Pictures/photo.png"] != "picture" || contents["Pictures/link"] != "../outside" {
		t.Fatal(contents)
	}
	if _, ok := contents["Pictures/empty/"]; !ok {
		t.Fatal("empty directory lost")
	}
	data, err := os.ReadFile(filepath.Join(root, "Pictures.zip"))
	if err != nil || string(data) != "keep" {
		t.Fatal("existing archive overwritten")
	}
}

func TestFolderZIPRejectsSymlinkRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.Symlink(root, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if _, err := createFolderZIP(filepath.Join(root, "link")); err == nil {
		t.Fatal("followed root symlink")
	}
}
