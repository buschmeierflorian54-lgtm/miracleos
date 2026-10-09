package main

/*
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>
static void overview_pointer(void *context, int *x, int *y) { GdkDevice *device=gdk_drag_context_get_device(GDK_DRAG_CONTEXT(context)); if(device) gdk_device_get_position(device,NULL,x,y); }
static void file_drag_set_uri_list(void *selection, const char *uris, int length) {
 // gotk3's SetURIs allocates an unterminated gchar** in the pinned version.
 // Set the standard URI-list bytes directly instead of reading past that array.
 gtk_selection_data_set((GtkSelectionData *)selection,
     gdk_atom_intern_static_string("text/uri-list"), 8,
     (const guchar *)uris, length);
}
static const guchar *file_drag_get_uri_list(void *selection, int *length) {
 if (!selection || gtk_selection_data_get_format((GtkSelectionData *)selection) != 8) {
  *length = -1;
  return NULL;
 }
 *length = gtk_selection_data_get_length((GtkSelectionData *)selection);
 return gtk_selection_data_get_data((GtkSelectionData *)selection);
}
static GdkDragAction file_drop_negotiate(void *widget, void *context, guint time) {
 GtkWidget *target_widget = GTK_WIDGET(widget);
 GdkDragContext *drag = GDK_DRAG_CONTEXT(context);
 GdkModifierType modifiers = 0;
 GdkDevice *device = gdk_drag_context_get_device(drag);
 GdkWindow *window = gtk_widget_get_window(target_widget);
 if (device && window)
  gdk_window_get_device_position(window, device, NULL, NULL, &modifiers);
 GdkDragAction allowed = gdk_drag_context_get_actions(drag);
 GdkDragAction action = 0;
 if (gtk_drag_dest_find_target(target_widget, drag, NULL) != GDK_NONE) {
  if (modifiers & GDK_CONTROL_MASK) {
   if (allowed & GDK_ACTION_COPY) action = GDK_ACTION_COPY;
  } else if (allowed & GDK_ACTION_MOVE) {
   action = GDK_ACTION_MOVE;
  } else if (allowed & GDK_ACTION_COPY) {
   action = GDK_ACTION_COPY;
  }
 }
 // Retain our explicit choice: Wayland action negotiation is asynchronous.
 // A stale selected_action must not silently turn a default move into a copy.
 g_object_set_data(G_OBJECT(drag), "miracleos-drop-action", GINT_TO_POINTER(action));
 gdk_drag_status(drag, action, time);
 if (action) gtk_drag_highlight(target_widget);
 else gtk_drag_unhighlight(target_widget);
 return action;
}
static gboolean file_drop_request(void *widget, void *context, guint time) {
 if (!file_drop_negotiate(widget, context, time)) return FALSE;
 GdkAtom target = gtk_drag_dest_find_target(GTK_WIDGET(widget), GDK_DRAG_CONTEXT(context), NULL);
 if (target == GDK_NONE) return FALSE;
 gtk_drag_get_data(GTK_WIDGET(widget), GDK_DRAG_CONTEXT(context), target, time);
 return TRUE;
}
static int file_drop_move(void *context) {
 return GPOINTER_TO_INT(g_object_get_data(G_OBJECT(context), "miracleos-drop-action")) == GDK_ACTION_MOVE;
}
static void file_drop_finish(void *context, gboolean success, guint time) {
 // We perform moves ourselves. Never ask the source to delete anything again.
 gtk_drag_finish(GDK_DRAG_CONTEXT(context), success, FALSE, time);
 g_object_unref(context);
}
*/
import "C"

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unsafe"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func fileDropPaths(uris []string) []string {
	var paths []string
	seen := map[string]bool{}
	for _, raw := range uris {
		raw = strings.TrimSpace(raw)
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		u, err := url.Parse(raw)
		if err != nil || u.Scheme != "file" || (u.Host != "" && u.Host != "localhost") || !filepath.IsAbs(u.Path) {
			continue
		}
		path := filepath.Clean(u.Path)
		if !seen[path] {
			paths = append(paths, path)
			seen[path] = true
		}
	}
	return paths
}

func fileDropSelectionPaths(data *gtk.SelectionData) []string {
	var length C.int
	bytes := C.file_drag_get_uri_list(unsafe.Pointer(data), &length)
	// Read only the bounded byte payload. gotk3's GetURIs string-array conversion
	// can dereference invalid pointers, even with a valid GTK URI list.
	if bytes == nil || length <= 0 || length > 1024*1024 {
		return nil
	}
	payload := C.GoBytes(unsafe.Pointer(bytes), length)
	if strings.IndexByte(string(payload), 0) >= 0 {
		return nil
	}
	return fileDropPaths(strings.Split(string(payload), "\n"))
}

func (fm *glassFileManager) enableFileDrag(widget gtk.IWidget, path string, flow *gtk.FlowBox, child *gtk.FlowBoxChild) {
	target, _ := gtk.TargetEntryNew("text/uri-list", 0, 0)
	widget.ToWidget().DragSourceSet(gdk.BUTTON1_MASK, []gtk.TargetEntry{*target}, gdk.ACTION_COPY|gdk.ACTION_MOVE)
	var dragged []string
	widget.ToWidget().Connect("drag-begin", func(_ interface{}, context *gdk.DragContext) {
		if child != nil && flow != nil && !child.IsSelected() {
			flow.UnselectAll()
			flow.SelectChild(child)
		}
		dragged = nil
		if flow != nil {
			dragged = append([]string(nil), fm.selectedPaths()...)
		}
		if len(dragged) == 0 {
			dragged = []string{path}
		}
		gtk.DragSetIconName(context, "document-send-symbolic", 0, 0)

		fm.beginDragOverview(context)
		fm.status.SetText("Ziehen zum Verschieben · Strg gedrückt halten zum Kopieren")
	})
	widget.ToWidget().Connect("drag-data-get", func(_ interface{}, _ *gdk.DragContext, data *gtk.SelectionData, _ uint, _ uint) {
		uris := make([]string, 0, len(dragged))
		for _, p := range dragged {
			uris = append(uris, (&url.URL{Scheme: "file", Path: p}).String())
		}
		payload := strings.Join(uris, "\r\n") + "\r\n"
		text := C.CString(payload)
		defer C.free(unsafe.Pointer(text))
		// SelectionData's marshaler returns the boxed GtkSelectionData pointer
		// directly; its exported struct field is not a wrapped native pointer.
		C.file_drag_set_uri_list(unsafe.Pointer(data), text, C.int(len(payload)))
	})
	widget.ToWidget().Connect("drag-end", func() {
		fm.endDragOverview()
		// A move into another window must remove the old tile here as well.
		glib.IdleAdd(func() {
			if fm.alive.Load() {
				fm.render()
			}
		})
	})
}

func (fm *glassFileManager) enableFileDrop(widget gtk.IWidget, destination string) {
	w := widget.ToWidget()
	target, _ := gtk.TargetEntryNew("text/uri-list", 0, 0)
	w.DragDestSet(0, []gtk.TargetEntry{*target}, gdk.ACTION_COPY|gdk.ACTION_MOVE)
	w.Connect("drag-motion", func(_ interface{}, context *gdk.DragContext, _ int, _ int, time uint) bool {
		action := C.file_drop_negotiate(unsafe.Pointer(w.GObject), unsafe.Pointer(context.GObject), C.guint(time))
		if action == C.GDK_ACTION_COPY {
			fm.status.SetText("Loslassen zum Kopieren")
		}
		if action == C.GDK_ACTION_MOVE {
			fm.status.SetText("Loslassen zum Verschieben · Strg zum Kopieren")
		}
		return action != 0
	})
	w.Connect("drag-leave", func() { C.gtk_drag_unhighlight((*C.GtkWidget)(unsafe.Pointer(w.GObject))) })
	w.Connect("drag-drop", func(_ interface{}, context *gdk.DragContext, _ int, _ int, time uint) bool {
		return C.file_drop_request(unsafe.Pointer(w.GObject), unsafe.Pointer(context.GObject), C.guint(time)) != 0
	})
	w.Connect("drag-data-received", func(_ interface{}, context *gdk.DragContext, _ int, _ int, data *gtk.SelectionData, _ uint, time uint) {
		paths := fileDropSelectionPaths(data)
		move := C.file_drop_move(unsafe.Pointer(context.GObject)) != 0
		pointer := unsafe.Pointer(context.GObject)
		C.g_object_ref(C.gpointer(pointer))
		fm.status.SetText("Dateien werden übertragen …")
		go func() {
			_, err := pasteFilePaths(paths, move, destination)
			glib.IdleAdd(func() {
				success := C.gboolean(0)
				if err == nil {
					success = 1
				}
				C.file_drop_finish(pointer, success, C.guint(time))
				if !fm.alive.Load() {
					return
				}
				fm.render()
				if err != nil {
					fm.status.SetText("Übertragen fehlgeschlagen: " + err.Error())
				} else {
					verb := "kopiert"
					if move {
						verb = "verschoben"
					}
					fm.status.SetText(fmt.Sprintf("%d Elemente %s.", len(paths), verb))
				}
			})
		}()
	})
}

// Hold a dragged item still for one second to reveal window drop targets.
func (fm *glassFileManager) beginDragOverview(context *gdk.DragContext) {
	fm.endDragOverview()
	pointer := unsafe.Pointer(context.GObject)
	C.g_object_ref(C.gpointer(pointer))
	active, shown := true, false
	fm.dragOverviewStop = func() {
		if active {
			active = false
			C.g_object_unref(C.gpointer(pointer))
		}
	}
	var nodes []desktopNode
	go func() {
		tree, err := desktopTree()
		if err != nil {
			return
		}
		var result []desktopNode
		var walk func(desktopNode)
		walk = func(n desktopNode) {
			if n.PID > 0 && n.PID != os.Getpid() && n.Name != "" {
				result = append(result, n)
			}
			for _, c := range append(n.Nodes, n.FloatingNodes...) {
				walk(c)
			}
		}
		walk(tree)
		sort.SliceStable(result, func(i, j int) bool { return result[i].ID > result[j].ID })
		glib.IdleAdd(func() {
			if active {
				nodes = result
			}
		})
	}()
	var lastX, lastY C.int
	C.overview_pointer(pointer, &lastX, &lastY)
	stillSince := time.Now()
	glib.TimeoutAdd(100, func() bool {
		if !active || !fm.alive.Load() {
			return false
		}
		var x, y C.int
		C.overview_pointer(pointer, &x, &y)
		if absDragPosition(int(x-lastX))+absDragPosition(int(y-lastY)) > 6 {
			stillSince = time.Now()
			lastX, lastY = x, y
		}
		if !shown && time.Since(stillSince) >= time.Second && len(nodes) > 0 {
			shown = true
			fm.showDragOverview(nodes)
		}
		return true
	})
}
func absDragPosition(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
func (fm *glassFileManager) endDragOverview() {
	if fm.dragOverviewStop != nil {
		fm.dragOverviewStop()
		fm.dragOverviewStop = nil
	}
	if fm.dragOverview != nil {
		fm.dragOverview.ToWidget().Destroy()
		fm.dragOverview = nil
	}
}
func (fm *glassFileManager) showDragOverview(nodes []desktopNode) {
	if fm.overlay == nil {
		return
	}
	shell, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 8)
	menuClass(shell, "drag-window-overview")
	shell.SetHAlign(gtk.ALIGN_CENTER)
	shell.SetVAlign(gtk.ALIGN_END)
	shell.SetMarginBottom(54)
	shell.SetMarginStart(18)
	shell.SetMarginEnd(18)
	shell.PackStart(menuLabel("In ein geöffnetes Fenster ziehen", "fm-card-title"), false, false, 0)
	scroll, _ := gtk.ScrolledWindowNew(nil, nil)
	scroll.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_NEVER)
	scroll.SetMinContentHeight(126)
	scroll.SetSizeRequest(560, -1)
	row, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	catalog := installedApplications()
	for _, node := range nodes {
		node := node
		card, _ := gtk.EventBoxNew()
		card.SetVisibleWindow(false)
		box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 8)
		menuClass(box, "drag-window-card")
		box.SetSizeRequest(155, 108)
		app := node.AppID
		if app == "" {
			app = node.WindowProperties.Class
		}
		image := installedAppIcon(installedApp{Icon: taskbarAppIcon(app, node.Name, catalog)})
		image.SetPixelSize(42)
		box.PackStart(image, true, true, 0)
		title := menuLabel(node.Name, "fm-file-name")
		title.SetMaxWidthChars(19)
		title.SetLineWrap(true)
		box.PackStart(title, false, false, 0)
		supported := false
		if _, err := os.Stat(windowDropSocket(node.PID)); err == nil {
			supported = true
		}
		hint := "Fenster anzeigen"
		if supported {
			hint = "Hier ablegen"
		}
		box.PackStart(menuLabel(hint, "fm-muted"), false, false, 0)
		card.Add(box)
		target, _ := gtk.TargetEntryNew("text/uri-list", 0, 0)
		card.DragDestSet(0, []gtk.TargetEntry{*target}, gdk.ACTION_COPY|gdk.ACTION_MOVE)
		hovered := false
		card.Connect("drag-motion", func(_ interface{}, ctx *gdk.DragContext, _ int, _ int, stamp uint) bool {
			if !supported {
				C.gdk_drag_status((*C.GdkDragContext)(unsafe.Pointer(ctx.GObject)), 0, C.guint(stamp))
				if !hovered {
					hovered = true
					glib.TimeoutAdd(350, func() bool {
						if hovered && fm.alive.Load() && fm.dragOverview != nil {
							focusDropWindow(node.ID)
							if fm.dragOverview != nil {
								fm.dragOverview.ToWidget().Hide()
							}
							fm.status.SetText("Jetzt direkt im geöffneten Programm ablegen.")
						}
						return false
					})
				}
				return true
			}
			return C.file_drop_negotiate(unsafe.Pointer(card.GObject), unsafe.Pointer(ctx.GObject), C.guint(stamp)) != 0
		})
		card.Connect("drag-leave", func() { hovered = false })
		card.Connect("drag-drop", func(_ interface{}, ctx *gdk.DragContext, _ int, _ int, stamp uint) bool {
			if !supported {
				return false
			}
			return C.file_drop_request(unsafe.Pointer(card.GObject), unsafe.Pointer(ctx.GObject), C.guint(stamp)) != 0
		})
		card.Connect("drag-data-received", func(_ interface{}, ctx *gdk.DragContext, _ int, _ int, data *gtk.SelectionData, _ uint, stamp uint) {
			paths := fileDropSelectionPaths(data)
			move := C.file_drop_move(unsafe.Pointer(ctx.GObject)) != 0
			ptr := unsafe.Pointer(ctx.GObject)
			C.g_object_ref(C.gpointer(ptr))
			go func() {
				focusDropWindowSync(node.ID)
				err := sendWindowDrop(node.PID, windowDropRequest{Paths: paths, Move: move})
				glib.IdleAdd(func() {
					ok := C.gboolean(0)
					if err == nil {
						ok = 1
					}
					C.file_drop_finish(ptr, ok, C.guint(stamp))
					if err != nil && fm.alive.Load() {
						fm.status.SetText("Dieses Fenster konnte die Datei nicht übernehmen.")
					}
				})
			}()
		})
		row.PackStart(card, false, false, 0)
	}
	scroll.Add(row)
	shell.PackStart(scroll, false, false, 0)
	fm.overlay.AddOverlay(shell)
	fm.dragOverview = shell
	shell.ShowAll()
}
