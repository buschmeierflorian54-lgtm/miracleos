package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

const fileClipboardType = "x-special/gnome-copied-files"

func clipboardFiles() ([]string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	data, err := exec.CommandContext(ctx, "wl-paste", "--no-newline", "--type", fileClipboardType).Output()
	if err != nil {
		return nil, false
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 || (lines[0] != "copy" && lines[0] != "cut") {
		return nil, false
	}
	var paths []string
	for _, line := range lines[1:] {
		u, err := url.Parse(strings.TrimSpace(line))
		if err != nil || u.Scheme != "file" || (u.Host != "" && u.Host != "localhost") || !filepath.IsAbs(u.Path) {
			continue
		}
		paths = append(paths, filepath.Clean(u.Path))
	}
	return paths, lines[0] == "cut"
}

func copyFilesToClipboard(paths []string, cut bool) error {
	action := "copy"
	if cut {
		action = "cut"
	}
	lines := []string{action}
	for _, path := range paths {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		lines = append(lines, (&url.URL{Scheme: "file", Path: absolute}).String())
	}
	if len(lines) == 1 {
		return errors.New("Keine Dateien ausgewählt")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "wl-copy", "--type", fileClipboardType)
	cmd.Stdin = strings.NewReader(strings.Join(lines, "\n") + "\n")
	return cmd.Run()
}

type clipboardSnapshot map[string]os.FileInfo

func snapshotClipboardSource(path string) (clipboardSnapshot, error) {
	result := clipboardSnapshot{}
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		result[p] = info
		return nil
	})
	return result, err
}
func unchangedClipboardSource(path string, previous clipboardSnapshot) error {
	current, err := snapshotClipboardSource(path)
	if err != nil {
		return err
	}
	if len(current) != len(previous) {
		return errors.New("Quelldateien wurden während des Kopierens verändert")
	}
	for p, old := range previous {
		now, ok := current[p]
		if !ok || !os.SameFile(old, now) || old.Mode() != now.Mode() || old.Size() != now.Size() || !old.ModTime().Equal(now.ModTime()) {
			return fmt.Errorf("Quelle wurde verändert: %s", p)
		}
	}
	return nil
}

func copyClipboardTree(source, target string, snapshot clipboardSnapshot) error {
	info := snapshot[source]
	if info == nil {
		return errors.New("Quelle fehlt")
	}
	if info.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(source)
		if err != nil {
			return err
		}
		return os.Symlink(link, target)
	}
	if info.IsDir() {
		if err := os.Mkdir(target, 0700); err != nil {
			return err
		}
		entries, err := os.ReadDir(source)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := copyClipboardTree(filepath.Join(source, entry.Name()), filepath.Join(target, entry.Name()), snapshot); err != nil {
				return err
			}
		}
		return os.Chmod(target, info.Mode().Perm())
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("Dieser Dateityp kann nicht kopiert werden: %s", source)
	}
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	opened, err := input.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(info, opened) {
		return errors.New("Quelle wurde ausgetauscht")
	}
	output, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return err
	}
	_, err = io.Copy(output, input)
	if err == nil {
		err = output.Sync()
	}
	closeErr := output.Close()
	if err == nil {
		err = closeErr
	}
	return err
}

// Removal is deliberately per entry, with identity checks; newly created files
// are never recursively removed if the original tree changes during a move.
func removeClipboardSource(path string, snapshot clipboardSnapshot) error {
	old := snapshot[path]
	now, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if old == nil || !os.SameFile(old, now) || old.Mode() != now.Mode() || (!old.IsDir() && (old.Size() != now.Size() || !old.ModTime().Equal(now.ModTime()))) {
		return fmt.Errorf("Quelle wurde verändert und bleibt erhalten: %s", path)
	}
	if old.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if snapshot[filepath.Join(path, entry.Name())] == nil {
				return errors.New("Quelle enthält neue Dateien und bleibt erhalten")
			}
		}
		for _, entry := range entries {
			if err := removeClipboardSource(filepath.Join(path, entry.Name()), snapshot); err != nil {
				return err
			}
		}
	}
	return os.Remove(path)
}

func pasteClipboardFiles(destination string) error {
	paths, cut := clipboardFiles()
	if len(paths) == 0 {
		return errors.New("Keine Dateien in der Zwischenablage")
	}
	pasted, err := pasteFilePaths(paths, cut, destination)
	if err != nil {
		return err
	}
	if cut {
		return copyFilesToClipboard(pasted, false)
	}
	return nil
}
func pasteFilePaths(paths []string, cut bool, destination string) ([]string, error) {
	if len(paths) == 0 {
		return nil, errors.New("Keine Dateien ausgewählt")
	}
	destination, err := filepath.EvalSymlinks(destination)
	if err != nil {
		return nil, err
	}
	destination, err = filepath.Abs(destination)
	if err != nil {
		return nil, err
	}
	dir, err := os.Stat(destination)
	if err != nil {
		return nil, err
	}
	if !dir.IsDir() {
		return nil, errors.New("Ziel ist kein Ordner")
	}
	var pasted []string
	for _, source := range paths {
		if cut {
			parent, parentErr := filepath.EvalSymlinks(filepath.Dir(source))
			if parentErr == nil && parent == destination {
				pasted = append(pasted, source)
				continue
			}
		}
		source, err = filepath.Abs(source)
		if err != nil {
			return nil, err
		}
		snapshot, err := snapshotClipboardSource(source)
		if err != nil {
			return nil, err
		}
		if snapshot[source].IsDir() {
			real, err := filepath.EvalSymlinks(source)
			if err != nil {
				return nil, err
			}
			relative, err := filepath.Rel(real, destination)
			if err != nil {
				return nil, err
			}
			if relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
				return nil, errors.New("Ein Ordner kann nicht in sich selbst kopiert werden")
			}
		}
		base := filepath.Base(source)
		var target string
		for attempt := 0; ; attempt++ {
			name := base
			if attempt > 0 {
				ext := filepath.Ext(base)
				if snapshot[source].IsDir() {
					ext = ""
				}
				suffix := " (Kopie)"
				if attempt > 1 {
					suffix = fmt.Sprintf(" (Kopie%d)", attempt)
				}
				name = strings.TrimSuffix(base, ext) + suffix + ext
			}
			target = filepath.Join(destination, name)
			if cut {
				err = unix.Renameat2(unix.AT_FDCWD, source, unix.AT_FDCWD, target, unix.RENAME_NOREPLACE)
				if err == nil {
					break
				}
				if errors.Is(err, unix.EEXIST) {
					continue
				}
				if !errors.Is(err, unix.EXDEV) {
					return nil, err
				}
			}
			err = copyClipboardTree(source, target, snapshot)
			if errors.Is(err, os.ErrExist) {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("Kopieren fehlgeschlagen; unvollständiges Ziel %s: %w", target, err)
			}
			if err = unchangedClipboardSource(source, snapshot); err != nil {
				return nil, fmt.Errorf("%w; Quelle bleibt erhalten", err)
			}
			if cut {
				if err = removeClipboardSource(source, snapshot); err != nil {
					return nil, err
				}
			}
			break
		}
		pasted = append(pasted, target)
	}
	return pasted, nil
}
