package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// createFolderZIP publishes a complete archive beside the source without
// replacing an existing file. Symbolic links are stored, never traversed.
func createFolderZIP(path string) (result string, err error) {
	source, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Lstat(source)
	if err != nil {
		return "", err
	}
	if !info.IsDir() || filepath.Dir(source) == source {
		return "", fmt.Errorf("Bitte einen Ordner auswählen")
	}
	parent, base := filepath.Dir(source), filepath.Base(source)
	tmp, err := os.CreateTemp(parent, ".miracleos-archive-*.tmp")
	if err != nil {
		return "", err
	}
	defer func() { tmp.Close(); os.Remove(tmp.Name()) }()
	writer := zip.NewWriter(tmp)
	err = filepath.WalkDir(source, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		mode := info.Mode()
		if !mode.IsRegular() && !mode.IsDir() && mode&os.ModeSymlink == 0 {
			return fmt.Errorf("Dateityp kann nicht archiviert werden: %s", current)
		}
		relative, err := filepath.Rel(parent, current)
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relative)
		if mode.IsDir() {
			header.Name += "/"
		} else if mode.IsRegular() {
			header.Method = zip.Deflate
		}
		output, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}
		if mode.IsDir() {
			return nil
		}
		if mode&os.ModeSymlink != 0 {
			target, err := os.Readlink(current)
			if err != nil {
				return err
			}
			_, err = io.WriteString(output, target)
			return err
		}
		fd, err := unix.Open(current, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
		if err != nil {
			return err
		}
		input := os.NewFile(uintptr(fd), current)
		defer input.Close()
		opened, err := input.Stat()
		if err != nil {
			return err
		}
		if !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
			return fmt.Errorf("Datei wurde während des Archivierens geändert: %s", current)
		}
		if _, err = io.Copy(output, input); err != nil {
			return err
		}
		after, err := input.Stat()
		if err != nil {
			return err
		}
		if opened.Size() != after.Size() || !opened.ModTime().Equal(after.ModTime()) {
			return fmt.Errorf("Datei wurde während des Archivierens geändert: %s", current)
		}
		return nil
	})
	closeErr := writer.Close()
	if err != nil {
		return "", err
	}
	if closeErr != nil {
		return "", closeErr
	}
	if err = tmp.Sync(); err != nil {
		return "", err
	}
	if err = tmp.Close(); err != nil {
		return "", err
	}
	for count := 0; ; count++ {
		name := base + ".zip"
		if count > 0 {
			name = fmt.Sprintf("%s (%d).zip", base, count)
		}
		result = filepath.Join(parent, name)
		if err = os.Link(tmp.Name(), result); os.IsExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		return result, nil
	}
}
