package main

import (
	"context"
	"fmt"
	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type contextAction struct {
	label    string
	run      func()
	disabled bool
}

func glassContextMenu(parent gtk.IWidget, event *gdk.Event, actions []contextAction) {
	menu := buildSharedMenu(actions)
	_ = menu.SetProperty("attach-widget", parent.ToWidget().Object)
	menu.ShowAll()
	menu.PopupAtPointer(event)
	menu.Connect("selection-done", func() { menu.Destroy() })
}
func copyPath(path string) {
	clipboard, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
	if err == nil {
		clipboard.SetText(path)
	}
}

// Every representation of a file or folder uses the same actions for its path.
func (fm *glassFileManager) bindFileContext(widget gtk.IWidget, path string) {
	w := widget.ToWidget()
	w.AddEvents(int(gdk.BUTTON_PRESS_MASK))
	w.Connect("button-press-event", func(_ interface{}, event *gdk.Event) bool {
		if gdk.EventButtonNewFromEvent(event).Button() != 3 {
			return false
		}
		fm.contextMenu(path, event)
		return true
	})
}
func (fm *glassFileManager) selectedPaths() []string {
	return append([]string(nil), fm.selection...)
}

func (fm *glassFileManager) copySelection(cut bool) {
	fm.copyPaths(fm.selectedPaths(), cut)
}

func (fm *glassFileManager) copyPaths(paths []string, cut bool) {
	if len(paths) == 0 {
		return
	}
	paths = append([]string(nil), paths...)
	go func() {
		err := copyFilesToClipboard(paths, cut)
		glib.IdleAdd(func() {
			if !fm.alive.Load() {
				return
			}
			if err != nil {
				fm.status.SetText("Die Auswahl konnte nicht in die Zwischenablage übernommen werden: " + err.Error())
				return
			}
			verb := "kopiert"
			if cut {
				verb = "zum Verschieben vorgemerkt"
			}
			fm.status.SetText(fmt.Sprintf("%d Elemente %s.", len(paths), verb))
		})
	}()
}

func (fm *glassFileManager) pasteInto(destination string) {
	fm.status.SetText("Dateien werden eingefügt …")
	go func() {
		err := pasteClipboardFiles(destination)
		glib.IdleAdd(func() {
			if !fm.alive.Load() {
				return
			}
			fm.render()
			if err != nil {
				fm.status.SetText("Einfügen fehlgeschlagen: " + err.Error())
			} else {
				fm.status.SetText("Dateien eingefügt.")
			}
		})
	}()
}

func (fm *glassFileManager) contextMenu(path string, event *gdk.Event) {
	clipboardPaths, _ := clipboardFiles()
	canPaste := func(destination string) bool {
		return len(clipboardPaths) > 0 && unix.Access(destination, unix.W_OK|unix.X_OK) == nil
	}
	if path == "" {
		actions := []contextAction{{"Neuer Ordner", fm.newFolder, false}, {"Aktualisieren", fm.render, false}}
		if canPaste(fm.current) {
			destination := fm.current
			actions = append(actions, contextAction{"Einfügen", func() { fm.pasteInto(destination) }, false})
		}
		actions = append(actions, contextAction{"", nil, false}, contextAction{"Ordnerpfad kopieren", func() { copyPath(fm.current) }, false}, contextAction{"Einstellungen", launchSettings, false})
		glassContextMenu(fm.window, event, actions)
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	paths := []string{path}
	for _, selected := range fm.selection {
		if selected == path {
			paths = fm.selectedPaths()
			break
		}
	}
	actions := []contextAction{{"Öffnen", func() { fm.open(path) }, false}}
	if fm.isTag(path) {
		actions = append(actions, contextAction{"Farbe ändern …", func() { fm.chooseTagColor(path) }, false})
	}
	if info.IsDir() && fm.tabsBar != nil {
		actions = append(actions, contextAction{"In neuem Tab öffnen", func() { fm.openTab(path) }, false})
	}
	if info.IsDir() {
		actions = append(actions, contextAction{"In neuem Fenster öffnen", func() { launchFileManager(path) }, false})
	} else if isFileImage(path) {
		actions = append(actions, contextAction{"Bildvorschau", func() { fm.previewImage(path) }, false})
	}
	actions = append(actions, contextAction{"", nil, false}, contextAction{"Kopieren", func() { fm.copyPaths(paths, false) }, false}, contextAction{"Ausschneiden", func() { fm.copyPaths(paths, true) }, false})
	destination := filepath.Dir(path)
	if info.IsDir() {
		destination = path
	}
	actions = append(actions, contextAction{"Einfügen", func() { fm.pasteInto(destination) }, !canPaste(destination)})
	if info.IsDir() {
		actions = append(actions, contextAction{"Als ZIP komprimieren", func() { fm.zipFolder(path) }, false})
	}
	actions = append(actions, contextAction{"", nil, false}, contextAction{"Pfad kopieren", func() { copyPath(path) }, false}, contextAction{"Umbenennen …", func() { fm.renameFile(path) }, false})
	actions = append(actions, contextAction{"In den Papierkorb", func() {
		fm.status.SetText("Datei wird in den Papierkorb verschoben …")
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			err := exec.CommandContext(ctx, "gio", "trash", path).Run()
			glib.IdleAdd(func() {
				if !fm.alive.Load() {
					return
				}
				fm.render()
				if err != nil {
					fm.status.SetText("Die Datei konnte nicht in den Papierkorb verschoben werden.")
				} else {
					fm.status.SetText("Datei in den Papierkorb verschoben.")
				}
			})
		}()
	}, false}, contextAction{"", nil, false}, contextAction{"Informationen", func() { fm.showFileInfo(path) }, false})
	glassContextMenu(fm.window, event, actions)
}
func (fm *glassFileManager) zipFolder(path string) {
	fm.status.SetText("ZIP-Datei wird erstellt …")
	go func() {
		archive, err := createFolderZIP(path)
		glib.IdleAdd(func() {
			if !fm.alive.Load() {
				return
			}
			fm.render()
			if err != nil {
				fm.status.SetText("ZIP konnte nicht erstellt werden: " + err.Error())
			} else {
				fm.status.SetText("ZIP erstellt: " + filepath.Base(archive))
			}
		})
	}()
}
func (fm *glassFileManager) renameFile(path string) {
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Umbenennen")
	win.SetTransientFor(fm.window)
	win.SetModal(true)
	win.SetDefaultSize(380, 180)
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	menuClass(box, "fm-dialog")
	box.PackStart(menuLabel("Neuer Name", "fm-card-title"), false, false, 0)
	entry, _ := gtk.EntryNew()
	entry.SetText(filepath.Base(path))
	menuClass(entry, "fm-path")
	box.PackStart(entry, false, false, 0)
	status := menuLabel("", "fm-muted")
	box.PackStart(status, false, false, 0)
	save, _ := gtk.ButtonNewWithLabel("Umbenennen")
	menuClass(save, "dialog-action")
	menuClass(save, "dialog-primary")
	save.Connect("clicked", func() {
		name, _ := entry.GetText()
		name = strings.TrimSpace(name)
		if name == "" || name == "." || name == ".." || strings.Contains(name, "/") {
			status.SetText("Bitte einen gültigen Namen eingeben.")
			return
		}
		target := filepath.Join(filepath.Dir(path), name)
		if target != path {
			if err := unix.Renameat2(unix.AT_FDCWD, path, unix.AT_FDCWD, target, unix.RENAME_NOREPLACE); err != nil {
				status.SetText("Dieser Name ist bereits vergeben oder kann nicht verwendet werden.")
				return
			}
		}
		win.Destroy()
		fm.render()
	})
	box.PackEnd(save, false, false, 0)
	title, _ := win.GetTitle()
	shell, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	menuClass(shell, "fm-shell")
	menuClass(shell, "glass-dialog")
	shell.PackStart(glassDialogHeader(win, title), false, false, 0)
	shell.PackStart(box, true, true, 0)
	dialogEscape(win)
	win.Add(shell)
	win.ShowAll()
}
func (fm *glassFileManager) showFileInfo(path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Dateiinformationen")
	win.SetTransientFor(fm.window)
	win.SetDefaultSize(420, 240)
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	menuClass(box, "fm-dialog")
	box.PackStart(menuLabel(info.Name(), "fm-card-title"), false, false, 0)
	box.PackStart(settingsText(fmt.Sprintf("%s\n\nGröße: %d Bytes\nGeändert: %s", path, info.Size(), info.ModTime().Format("02.01.2006 15:04"))), false, false, 0)
	box.PackEnd(settingsButton("Schließen", func() { win.Destroy() }), false, false, 0)
	title, _ := win.GetTitle()
	shell, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	menuClass(shell, "fm-shell")
	menuClass(shell, "glass-dialog")
	shell.PackStart(glassDialogHeader(win, title), false, false, 0)
	shell.PackStart(box, true, true, 0)
	dialogEscape(win)
	win.Add(shell)
	win.ShowAll()
}

var desktopBackground *gtk.Window

func createDesktopBackground() {
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	desktopBackground = win
	win.SetTitle("MiracleOS Hintergrund")
	win.SetDecorated(false)
	win.SetSkipTaskbarHint(true)
	win.SetSkipPagerHint(true)
	layershell.InitForWindow(win)
	layershell.SetNamespace(win, "miracleos-background")
	layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_BACKGROUND)
	for _, edge := range []layershell.LayerShellEdgeFlags{layershell.LAYER_SHELL_EDGE_TOP, layershell.LAYER_SHELL_EDGE_BOTTOM, layershell.LAYER_SHELL_EDGE_LEFT, layershell.LAYER_SHELL_EDGE_RIGHT} {
		layershell.SetAnchor(win, edge, true)
	}
	image, _ := gtk.ImageNew()
	events, _ := gtk.EventBoxNew()
	events.Add(image)
	events.AddEvents(int(gdk.BUTTON_PRESS_MASK))
	events.Connect("button-press-event", func(_ *gtk.EventBox, event *gdk.Event) bool {
		if gdk.EventButtonNewFromEvent(event).Button() != 3 {
			return false
		}
		glassContextMenu(win, event, []contextAction{{"Dateien öffnen", func() { launchFileManager("") }, false}, {"", nil, false}, {"Hintergrund ändern …", func() { exe, _ := os.Executable(); _ = exec.Command(exe, "--settings", "Hintergrund").Start() }, false}, {"Theme & Farben …", func() { exe, _ := os.Executable(); _ = exec.Command(exe, "--settings", "Farben").Start() }, false}, {"Einstellungen", launchSettings, false}})
		return true
	})
	overlay,_:=gtk.OverlayNew();overlay.Add(events)
 canvas:=newWidgetCanvas(win,"desktop");overlay.AddOverlay(canvas.root)
 win.Add(overlay)
	last := ""
	update := func() bool {
		p := readSystemPreferences()
		path := p.Wallpaper
		if path == "" {
			home, _ := os.UserHomeDir()
			path = filepath.Join(home, ".local/share/backgrounds/miracleos-wallpaper.png")
		}
		if path != last {
			width, height := win.GetAllocatedWidth(), win.GetAllocatedHeight()
			if width <= 1 || height <= 1 {
				return true
			}
			if pixbuf, err := gdk.PixbufNewFromFileAtScale(path, width, height, false); err == nil {
				image.SetFromPixbuf(pixbuf)
				last = path
			}
		}
		return true
	}
	update()
	glib.TimeoutAdd(800, update)
	win.ShowAll()
}
