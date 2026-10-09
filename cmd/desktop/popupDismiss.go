package main

import (
	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"time"
)

// A transparent input surface below the popup catches outside clicks, including
// clicks on clients that do not transfer keyboard focus from an exclusive layer.
func dismissPopupOutside(win *gtk.Window, closePopup func()) {
	var shield *gtk.Window
	alive, closing := true, false
	generation := 0
	cleanup := func() {
		generation++
		if shield != nil {
			shield.Destroy()
			shield = nil
		}
	}
	win.Connect("destroy", func() { alive = false; cleanup() })
	win.Connect("unmap", func() { cleanup() })
	win.Connect("map", func() {
		cleanup()
		closing = false
		win.SetOpacity(1)
		shield, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
		shield.SetDecorated(false)
		shield.SetAppPaintable(true)
		if screen := shield.GetScreen(); screen != nil {
			if visual, e := screen.GetRGBAVisual(); e == nil {
				shield.SetVisual(visual)
			}
		}
		menuClass(shield, "popup-dismiss-shield")
		layershell.InitForWindow(shield)
		layershell.SetNamespace(shield, "miracleos-popup-dismiss")
		layershell.SetLayer(shield, layershell.LAYER_SHELL_LAYER_TOP)
		layershell.SetKeyboardMode(shield, layershell.LAYER_SHELL_KEYBOARD_MODE_NONE)
		for _, edge := range []layershell.LayerShellEdgeFlags{layershell.LAYER_SHELL_EDGE_TOP, layershell.LAYER_SHELL_EDGE_BOTTOM, layershell.LAYER_SHELL_EDGE_LEFT, layershell.LAYER_SHELL_EDGE_RIGHT} {
			layershell.SetAnchor(shield, edge, true)
		}
		events, _ := gtk.EventBoxNew()
		events.SetVisibleWindow(false)
		events.AddEvents(int(gdk.BUTTON_PRESS_MASK))
		events.Connect("button-press-event", func() bool {
			if closing {
				return true
			}
			closing = true
			token := generation
			started := time.Now()
			glib.TimeoutAdd(16, func() bool {
				if !alive || token != generation {
					return false
				}
				progress := float64(time.Since(started)) / float64(160*time.Millisecond)
				if progress >= 1 {
					closePopup()
					return false
				}
				win.SetOpacity(1 - progress*progress*(3-2*progress))
				return true
			})
			return true
		})
		layoutDropTarget(events,func(p layoutDrag,x,y int)bool{
 if y>shield.GetAllocatedHeight()-90 && centerDockWidth>0 && x>shield.GetAllocatedWidth()/2-centerDockWidth/2-60 && x<shield.GetAllocatedWidth()/2+centerDockWidth/2+60 {
  app,ok:=draggedApp(p);if !ok{return false};return pinApp(app,true)==nil
 }
 ok:=dropOnSurface("desktop",p,x,y);if ok{glib.IdleAdd(closePopup)};return ok
})
shield.Add(events)
		shield.ShowAll()
	})
}
