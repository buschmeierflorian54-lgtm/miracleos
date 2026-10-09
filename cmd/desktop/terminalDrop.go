package main

/*
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>
static void terminal_drop_finish(void *widget,void *context,guint time,gboolean ok) {
 g_signal_stop_emission_by_name(widget,"drag-data-received");
 gtk_drag_finish(GDK_DRAG_CONTEXT(context),ok,FALSE,time);
}
*/
import "C"
import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unsafe"
)

func terminalQuotedPaths(paths []string) string {
	result := []string{}
	for _, path := range paths {
		result = append(result, "'"+strings.ReplaceAll(path, "'", "'\"'\"'")+"'")
	}
	return strings.Join(result, " ") + " "
}
func enableTerminalDrop(win *gtk.Window, widget *gtk.Widget, insert func(string)) {
	target, _ := gtk.TargetEntryNew("text/uri-list", 0, 0)
	widget.DragDestSet(gtk.DEST_DEFAULT_ALL, []gtk.TargetEntry{*target}, gdk.ACTION_COPY)
	widget.Connect("drag-data-received", func(_ interface{}, context *gdk.DragContext, _ int, _ int, data *gtk.SelectionData, _ uint, time uint) {
		paths := fileDropSelectionPaths(data)
		ok := C.gboolean(0)
		if len(paths) > 0 {
			ok = 1
		}
		C.terminal_drop_finish(unsafe.Pointer(widget.GObject), unsafe.Pointer(context.GObject), C.guint(time), ok)
		if len(paths) == 0 {
			return
		}
		showTerminalDropChoices(win, widget, insert, paths)
	})
	registerWindowDrop(win, func(request windowDropRequest, complete func(error)) {
		showTerminalDropChoices(win, widget, insert, request.Paths)
		complete(nil)
	})
}
func showTerminalDropChoices(win *gtk.Window, widget *gtk.Widget, insert func(string), paths []string) {
	dialog, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	dialog.SetTitle("Dateien im Terminal")
	dialog.SetTransientFor(win)
	dialog.SetModal(true)
	dialog.SetDefaultSize(420, 380)
	root, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	menuClass(root, "fm-shell")
	menuClass(root, "glass-dialog")
	root.PackStart(glassDialogHeader(dialog, "Dateien im Terminal"), false, false, 0)
	body, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	menuClass(body, "fm-dialog")
	body.PackStart(menuLabel("Was soll passieren?", "dialog-heading"), false, false, 0)
	label := menuLabel(fmt.Sprintf("%d ausgewählte Elemente", len(paths)), "fm-muted")
	body.PackStart(label, false, false, 0)
	preview, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 12)
	menuClass(preview, "dialog-preview")
	preview.PackStart(fmIcon("text-x-generic-symbolic", gtk.ICON_SIZE_DIALOG), false, false, 0)
	detail, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 3)
	name := menuLabel(filepath.Base(paths[0]), "dialog-filename")
	name.SetXAlign(0)
	name.SetMaxWidthChars(30)
	name.SetLineWrap(true)
	detail.PackStart(name, false, false, 0)
	location := menuLabel(filepath.Dir(paths[0]), "fm-muted")
	location.SetMaxWidthChars(35)
	location.SetLineWrap(true)
	location.SetXAlign(0)
	detail.PackStart(location, false, false, 0)
	preview.PackStart(detail, true, true, 0)
	body.PackStart(preview, false, false, 0)
	var primary, openButton *gtk.Button
	for _, choice := range []struct {
		name string
		run  func()
	}{
		{"Pfad einfügen", func() { insert(terminalQuotedPaths(paths)); dialog.Destroy(); widget.GrabFocus() }},
		{"Öffnen", func() {
			for _, path := range paths {
				if info, err := os.Stat(path); err == nil && info.IsDir() {
					launchFileManager(path)
				} else {
					exec.Command("gio", "open", path).Start()
				}
			}
			dialog.Destroy()
		}},
		{"Abbrechen", func() { dialog.Destroy() }},
	} {
		choice := choice
		icon, key := "document-open-symbolic", "O"
		if choice.name == "Pfad einfügen" {
			icon, key = "edit-paste-symbolic", "Enter"
		}
		if choice.name == "Abbrechen" {
			icon, key = "", "Esc"
		}
		b := dialogAction(choice.name, icon, key, choice.name == "Pfad einfügen", choice.run)
		if choice.name == "Pfad einfügen" {
			primary = b
		}
		if choice.name == "Öffnen" {
			openButton = b
		}
		if choice.name == "Abbrechen" {
			menuClass(b, "dialog-cancel")
		}
		body.PackStart(b, false, false, 0)
	}
	dialog.Connect("key-press-event", func(_ interface{}, event *gdk.Event) bool {
		key := gdk.EventKeyNewFromEvent(event).KeyVal()
		if key == gdk.KEY_Return && primary != nil {
			primary.Clicked()
			return true
		}
		if (key == 'o' || key == 'O') && openButton != nil {
			openButton.Clicked()
			return true
		}
		if key == gdk.KEY_Escape {
			dialog.Destroy()
			return true
		}
		return false
	})
	root.PackStart(body, true, true, 0)
	dialog.Add(root)
	dialog.ShowAll()
	if primary != nil {
		primary.GrabFocus()
	}
}
