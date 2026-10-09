package main

import (
	_ "embed"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

//go:embed images/logo.png
var startupLogo []byte

func showStartupScreen(preview bool) {
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Willkommen — MiracleOS")
	win.SetDecorated(false)
	win.Fullscreen()
	menuClass(win, "startup-screen")
	root, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	content, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 18)
	content.SetHAlign(gtk.ALIGN_CENTER)
	root.SetCenterWidget(content)
	if pixbuf, err := gdk.PixbufNewFromBytesOnly(startupLogo); err == nil {
		logo, _ := gtk.ImageNewFromPixbuf(scalePixbuf(pixbuf, 104, 104))
		content.PackStart(logo, false, false, 10)
	}
	title := menuLabel("MiracleOS", "startup-title")
	title.SetHAlign(gtk.ALIGN_CENTER)
	content.PackStart(title, false, false, 0)
	subtitle := menuLabel("Willkommen zu Hause", "startup-subtitle")
	subtitle.SetHAlign(gtk.ALIGN_CENTER)
	content.PackStart(subtitle, false, false, 0)
	spinner, _ := gtk.SpinnerNew()
	spinner.SetHAlign(gtk.ALIGN_CENTER)
	spinner.SetSizeRequest(26, 26)
	spinner.Start()
	content.PackStart(spinner, false, false, 18)
	win.Add(root)
	win.ShowAll()
	duration := uint(2400)
	if preview {
		duration = 8000
	}
	glib.TimeoutAdd(duration, func() bool {
		win.Destroy()
		if preview {
			gtk.MainQuit()
		}
		return false
	})
}
