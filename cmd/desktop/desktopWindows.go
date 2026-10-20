package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type desktopRect struct{ X, Y, Width, Height int }
type desktopNode struct {
	AppID            string `json:"app_id"`
	WindowProperties struct {
		Class string `json:"class"`
	} `json:"window_properties"`
	ID            int64         `json:"id"`
	Marks         []string      `json:"marks"`
	FullscreenMode int          `json:"fullscreen_mode"`
	Name          string        `json:"name"`
	PID           int           `json:"pid"`
	Rect          desktopRect   `json:"rect"`
	Nodes         []desktopNode `json:"nodes"`
	FloatingNodes []desktopNode `json:"floating_nodes"`
}

func desktopTree() (desktopNode, error) {
	var tree desktopNode
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	data, err := exec.CommandContext(ctx, "swaymsg", "-t", "get_tree", "-r").Output()
	if err != nil {
		return tree, err
	}
	err = json.Unmarshal(data, &tree)
	return tree, err
}

func windowSignature(tree desktopNode) string {
	var signature strings.Builder
	var walk func(desktopNode)
	walk = func(node desktopNode) {
		fmt.Fprintf(&signature, "(%d:%d:%s;", node.ID, node.PID, node.Name)
		for _, child := range append(node.Nodes, node.FloatingNodes...) {
			walk(child)
		}
		signature.WriteString(")")
	}
	walk(tree)
	return signature.String()
}

func findDesktopWindow(node desktopNode, title string) (desktopNode, bool) {
	if node.PID == os.Getpid() && node.Name == title {
		return node, true
	}
	for _, child := range append(node.Nodes, node.FloatingNodes...) {
		if result, ok := findDesktopWindow(child, title); ok {
			return result, true
		}
	}
	return desktopNode{}, false
}

func desktopWindowCommand(title, command string) bool {
	selector := fmt.Sprintf("[pid=%d title=%s] ", os.Getpid(), strconv.Quote("^"+regexp.QuoteMeta(title)+"$"))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	data, err := exec.CommandContext(ctx, "swaymsg", "-r", selector+command).Output()
	if err != nil {
		return false
	}
	var replies []struct {
		Success bool `json:"success"`
	}
	if json.Unmarshal(data, &replies) != nil || len(replies) == 0 {
		return false
	}
	return replies[0].Success
}

func minimizedDesktopWindows() []desktopNode {
	tree, err := desktopTree()
	if err != nil {
		return nil
	}
	return minimizedWindowsFromTree(tree)
}
func minimizedWindowsFromTree(tree desktopNode) []desktopNode {
	var result []desktopNode
	var walk func(desktopNode)
	walk = func(node desktopNode) {
		if node.Name == "__i3_scratch" {
			var collect func(desktopNode)
			collect = func(child desktopNode) {
				for _,mark:=range child.Marks { if mark=="__miracleos_tab_hidden" {return} }
				if child.PID > 0 && child.Name != "" {
					result = append(result, child)
					return
				}
				for _, descendant := range append(child.Nodes, child.FloatingNodes...) {
					collect(descendant)
				}
			}
			for _, child := range append(node.Nodes, node.FloatingNodes...) {
				collect(child)
			}
			return
		}
		for _, child := range append(node.Nodes, node.FloatingNodes...) {
			walk(child)
		}
	}
	walk(tree)
	return result
}

func glassWindowHeader(win *gtk.Window, title string, canClose ...func() bool) *gtk.Box {
	menuClass(win, "miracleos-glass-window")
	win.SetDecorated(false)
	win.SetAppPaintable(true)
	if screen := win.GetScreen(); screen != nil {
		if visual, visualErr := screen.GetRGBAVisual(); visualErr == nil {
			win.SetVisual(visual)
		}
	}
	if serverWindowDecorations() {
 menuClass(win, "miracleos-server-window")
		win.SetDecorated(true)
		if len(canClose) > 0 {
			win.Connect("delete-event", func() bool { return !canClose[0]() })
		}
		empty, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
		return empty
	}
	bar, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 8)
	menuClass(bar, "glass-titlebar")
	iconID := "files"
	if strings.Contains(title, "Paketinstaller") {
		iconID = "store"
	}
	icon := menuIcon(iconID, 26)
	if strings.Contains(title, "Einstellungen") {
		icon = settingIcon("preferences-system", 24)
	}
	bar.PackStart(icon, false, false, 0)
	caption := menuLabel(title, "glass-window-title")
	drag, _ := gtk.EventBoxNew()
	drag.SetVisibleWindow(false)
	drag.SetHExpand(true)
	dragContent, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	dragContent.PackStart(caption, false, false, 0)
	drag.Add(dragContent)
	drag.AddEvents(int(gdk.BUTTON_PRESS_MASK))
	drag.Connect("button-press-event", func(_ *gtk.EventBox, event *gdk.Event) bool {
		button := gdk.EventButtonNewFromEvent(event)
		if button.Button() != 1 {
			return false
		}
		win.BeginMoveDrag(button.Button(), int(button.XRoot()), int(button.YRoot()), button.Time())
		return true
	})
	bar.PackStart(drag, true, true, 0)
	minimize, _ := gtk.ButtonNewWithLabel("−")
	maximize, _ := gtk.ButtonNewWithLabel("□")
	close, _ := gtk.ButtonNewWithLabel("×")
	minimize.SetTooltipText("Minimieren · über die Taskleiste wieder öffnen")
	maximize.SetTooltipText("Maximieren / Wiederherstellen")
	close.SetTooltipText("Schließen")
	for _, button := range []*gtk.Button{minimize, maximize, close} {
		menuClass(button, "glass-window-control")
	}
	menuClass(close, "glass-window-close")
	minimize.Connect("clicked", func() {
		if !desktopWindowCommand(title, "move scratchpad") {
			win.Iconify()
		}
	})
	maximized := false
	previous := desktopRect{X: 100, Y: 75, Width: 1080, Height: 650}
	maximize.Connect("clicked", func() {
		if !maximized {
			if tree, err := desktopTree(); err == nil {
				if node, found := findDesktopWindow(tree, title); found {
					previous = node.Rect
				}
			}
			if !desktopWindowCommand(title, "floating disable") {
				win.Maximize()
			}
			maximize.SetLabel("▣")
		} else {
			command := fmt.Sprintf("floating enable, resize set width %d px height %d px, move position %d px %d px", previous.Width, previous.Height, previous.X, previous.Y)
			if !desktopWindowCommand(title, command) {
				win.Unmaximize()
			}
			maximize.SetLabel("□")
		}
		maximized = !maximized
	})
	close.Connect("clicked", func() {
		if len(canClose) > 0 && !canClose[0]() {
			return
		}
		win.Destroy()
	})
	bar.PackEnd(close, false, false, 0)
	bar.PackEnd(maximize, false, false, 0)
	bar.PackEnd(minimize, false, false, 0)
	return bar
}
