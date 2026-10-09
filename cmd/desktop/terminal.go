package main

/*
#cgo pkg-config: vte-2.91
#include <vte/vte.h>
#include <stdlib.h>
static void miracle_terminal_spawn_done(VteTerminal *terminal, GPid pid, GError *error, gpointer data) { if (error) g_printerr("Terminal spawn: %s\n",error->message); }
static void miracle_terminal_spawn(VteTerminal *terminal, const char *shell, const char *directory) {
 char *argv[]={(char*)shell,NULL};
 vte_terminal_spawn_async(terminal,VTE_PTY_DEFAULT,directory,argv,NULL,G_SPAWN_DEFAULT,NULL,NULL,NULL,-1,NULL,miracle_terminal_spawn_done,NULL);
}
static void miracle_terminal_colors(VteTerminal *terminal,const char *foreground,const char *background,const char *accent) {
 GdkRGBA fg,bg,cursor;gdk_rgba_parse(&fg,foreground);gdk_rgba_parse(&bg,background);gdk_rgba_parse(&cursor,accent);
 vte_terminal_set_colors(terminal,&fg,&bg,NULL,0);vte_terminal_set_color_cursor(terminal,&cursor);
}
*/
import "C"
import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"os"
	"unsafe"
)

func showTerminal() {
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Terminal — MiracleOS")
	win.SetDefaultSize(850, 540)
	root, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	menuClass(root, "fm-shell")
	root.PackStart(glassWindowHeader(win, "Terminal — MiracleOS"), false, false, 0)
	native := C.vte_terminal_new()
	terminal := (*C.VteTerminal)(unsafe.Pointer(native))
	widget := &gtk.Widget{InitiallyUnowned: glib.InitiallyUnowned{Object: glib.Take(unsafe.Pointer(native))}}
	enableTerminalDrop(win, widget, func(text string) {
		value := C.CString(text)
		C.vte_terminal_paste_text(terminal, value)
		C.free(unsafe.Pointer(value))
	})
	widget.SetHExpand(true)
	widget.SetVExpand(true)
	widget.SetMarginStart(12)
	widget.SetMarginEnd(12)
	widget.SetMarginTop(10)
	widget.SetMarginBottom(10)
	C.vte_terminal_set_scrollback_lines(terminal, 10000)
	font := C.CString("Monospace 11")
	description := C.pango_font_description_from_string(font)
	C.free(unsafe.Pointer(font))
	C.vte_terminal_set_font(terminal, description)
	C.pango_font_description_free(description)
	applyColors := func() {
		prefs := readSystemPreferences()
		fg := C.CString("#f4eaf5")
		bg := C.CString("#241c2b")
		accent := C.CString(prefs.Accent)
		C.miracle_terminal_colors(terminal, fg, bg, accent)
		C.free(unsafe.Pointer(fg))
		C.free(unsafe.Pointer(bg))
		C.free(unsafe.Pointer(accent))
	}
	applyColors()
	copySelection := func() { C.vte_terminal_copy_clipboard_format(terminal, C.VTE_FORMAT_TEXT) }
	paste := func() { C.vte_terminal_paste_clipboard(terminal) }
	scale := 1.0
	terminalMenu := buildSharedMenu([]contextAction{
		{"Kopieren", copySelection, false}, {"Einfügen", paste, false},
		{"Alles auswählen", func() { C.vte_terminal_select_all(terminal) }, false}, {"", nil, false},
		{"Schrift vergrößern", func() {
			if scale < 2 {
				scale += .1
				C.vte_terminal_set_font_scale(terminal, C.double(scale))
			}
		}, false},
		{"Schrift verkleinern", func() {
			if scale > .6 {
				scale -= .1
				C.vte_terminal_set_font_scale(terminal, C.double(scale))
			}
		}, false},
	})
	terminalMenu.ShowAll()
	C.vte_terminal_set_context_menu(terminal, (*C.GtkWidget)(unsafe.Pointer(terminalMenu.GObject)))
	widget.Connect("setup-context-menu", func() {
		first := true
		terminalMenu.GetChildren().Foreach(func(item interface{}) {
			if first {
				if row, ok := item.(gtk.IWidget); ok {
					row.ToWidget().SetSensitive(C.vte_terminal_get_has_selection(terminal) != 0)
				}
				first = false
			}
		})
	})
	widget.Connect("key-press-event", func(_ interface{}, event *gdk.Event) bool {
		key := gdk.EventKeyNewFromEvent(event)
		if key.State()&uint(gdk.CONTROL_MASK|gdk.SHIFT_MASK) != uint(gdk.CONTROL_MASK|gdk.SHIFT_MASK) {
			return false
		}
		switch key.KeyVal() {
		case 'c', 'C':
			copySelection()
			return true
		case 'v', 'V':
			paste()
			return true
		}
		return false
	})
	widget.Connect("child-exited", func() { win.Destroy() })
	alive := true
	win.Connect("destroy", func() { alive = false; gtk.MainQuit() })
	glib.TimeoutAdd(1000, func() bool {
		if !alive {
			return false
		}
		applyColors()
		return true
	})
	root.PackStart(widget, true, true, 0)
	win.Add(root)
	win.ShowAll()
	widget.GrabFocus()
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}
	home, _ := os.UserHomeDir()
	cshell := C.CString(shell)
	cdir := C.CString(home)
	C.miracle_terminal_spawn(terminal, cshell, cdir)
	C.free(unsafe.Pointer(cshell))
	C.free(unsafe.Pointer(cdir))
}
