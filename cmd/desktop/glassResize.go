package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func resizableGlass(win *gtk.Window, content gtk.IWidget) *gtk.Grid {
	win.SetResizable(true)
	grid, _ := gtk.GridNew()
	grid.SetHExpand(true)
	grid.SetVExpand(true)
	content.ToWidget().SetHExpand(true)
	content.ToWidget().SetVExpand(true)
	grid.Attach(content, 1, 1, 1, 1)
	if serverWindowDecorations() {
		return grid
	}
	edges := []struct {
		x, y, w, h int
		edge       gdk.WindowEdge
	}{
		{0, 0, 1, 1, gdk.WINDOW_EDGE_NORTH_WEST}, {1, 0, 1, 1, gdk.WINDOW_EDGE_NORTH}, {2, 0, 1, 1, gdk.WINDOW_EDGE_NORTH_EAST},
		{0, 1, 1, 1, gdk.WINDOW_EDGE_WEST}, {2, 1, 1, 1, gdk.WINDOW_EDGE_EAST},
		{0, 2, 1, 1, gdk.WINDOW_EDGE_SOUTH_WEST}, {1, 2, 1, 1, gdk.WINDOW_EDGE_SOUTH}, {2, 2, 1, 1, gdk.WINDOW_EDGE_SOUTH_EAST}}
	for _, e := range edges {
		e := e
		handle, _ := gtk.EventBoxNew()
		handle.SetVisibleWindow(false)
		handle.SetSizeRequest(6, 6)
		handle.AddEvents(int(gdk.BUTTON_PRESS_MASK))
		handle.Connect("button-press-event", func(_ interface{}, event *gdk.Event) bool {
			b := gdk.EventButtonNewFromEvent(event)
			if b.Button() != 1 {
				return false
			}
			win.BeginResizeDrag(e.edge, 1, int(b.XRoot()), int(b.YRoot()), b.Time())
			return true
		})
		grid.Attach(handle, e.x, e.y, e.w, e.h)
	}
	return grid
}
