package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func glassDialogHeader(win *gtk.Window, title string) *gtk.Box {
	setup := glassWindowHeader(win, title)
	if serverWindowDecorations() {
		return setup
	}
	setup.Destroy()
	header, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	menuClass(header, "dialog-titlebar")
	header.PackStart(fmIcon("utilities-terminal-symbolic", gtk.ICON_SIZE_MENU), false, false, 0)
	drag, _ := gtk.EventBoxNew()
	drag.SetVisibleWindow(false)
	drag.SetHExpand(true)
	caption := menuLabel(title, "glass-window-title")
	caption.SetXAlign(0)
	drag.Add(caption)
	drag.AddEvents(int(gdk.BUTTON_PRESS_MASK))
	drag.Connect("button-press-event", func(_ interface{}, event *gdk.Event) bool {
		b := gdk.EventButtonNewFromEvent(event)
		if b.Button() != 1 {
			return false
		}
		win.BeginMoveDrag(1, int(b.XRoot()), int(b.YRoot()), b.Time())
		return true
	})
	header.PackStart(drag, true, true, 0)
	close, _ := gtk.ButtonNewWithLabel("×")
	menuClass(close, "glass-window-control")
	close.SetTooltipText("Schließen")
	close.Connect("clicked", func() { win.Destroy() })
	header.PackEnd(close, false, false, 0)
	return header
}
func dialogAction(label, icon, key string, primary bool, run func()) *gtk.Button {
	b, _ := gtk.ButtonNew()
	menuClass(b, "dialog-action")
	if primary {
		menuClass(b, "dialog-primary")
	}
	row, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 12)
	if icon != "" {
		row.PackStart(fmIcon(icon, gtk.ICON_SIZE_MENU), false, false, 0)
	}
	text := menuLabel(label, "")
	text.SetXAlign(0)
	row.PackStart(text, true, true, 0)
	if key != "" {
		hint := menuLabel(key, "dialog-key")
		row.PackEnd(hint, false, false, 0)
	}
	b.Add(row)
	b.Connect("clicked", run)
	return b
}
func dialogEscape(win *gtk.Window) {
	win.Connect("key-press-event", func(_ interface{}, event *gdk.Event) bool {
		if gdk.EventKeyNewFromEvent(event).KeyVal() == gdk.KEY_Escape {
			win.Destroy()
			return true
		}
		return false
	})
}
