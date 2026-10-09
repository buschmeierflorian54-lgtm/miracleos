package main

import (
	"context"

	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"os/exec"
	"strings"

	"time"
)

func taskbarCommand(id int64, action string) {
	go func() {
		if action == "focus" || action == "scratchpad show, focus" {
			if tree, err := desktopTree(); err == nil {
				id = modalFocusTarget(tree, id)
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		exec.CommandContext(ctx, "swaymsg", fmt.Sprintf("[con_id=%d] %s", id, action)).Run()
	}()
}

var windowContextColors = map[int64]string{}

func pruneWindowColors(snapshot taskbarSnapshot) {
	live := map[int64]bool{}
	for _, entry := range snapshot.Windows {
		live[entry.ID] = true
	}
	for _, node := range snapshot.Minimized {
		live[node.ID] = true
	}
	for id := range windowContextColors {
		if !live[id] {
			delete(windowContextColors, id)
		}
	}
}
func taskbarWindowButton(id int64, app, title, icon string, minimized bool) *gtk.Button {
	button, _ := gtk.ButtonNew()
	menuClass(button, "app")
	button.SetTooltipText(title)
	identity := installedApp{Icon: icon}
	if app == "miracleos-files" || strings.HasPrefix(title, "Dateien —") || strings.HasPrefix(title, "Dateiauswahl —") {
		identity.Name = "Dateien"
	} else if app == "miracleos-settings" || strings.HasPrefix(title, "Einstellungen —") {
		identity.Name = "Einstellungen"
	}
	button.Add(installedAppIcon(identity))

	provider, _ := gtk.CssProviderNew()
	style, _ := button.GetStyleContext()
	style.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER+1)
	apply := func(color string) {
		valid := false
		for _, c := range tagPalette {
			if color == c {
				valid = true
			}
		}
		css := "button {border: 2px solid transparent; border-radius: 999px;}"
		if valid {
			css = fmt.Sprintf("button {border: 2px solid %s; border-radius: 999px; box-shadow: 0 0 7px alpha(%s,.35);}", color, color)
		}
		provider.LoadFromData(css)
	}
	apply(windowContextColors[id])
	button.Connect("clicked", func() {

		hideStartMenu()
		action := "focus"
		if minimized {
			action = "scratchpad show, focus"
		}
		taskbarCommand(id, action)
	})
	button.AddEvents(int(gdk.BUTTON_PRESS_MASK))
	button.Connect("button-press-event", func(_ interface{}, event *gdk.Event) bool {
		if gdk.EventButtonNewFromEvent(event).Button() != 3 {
			return false
		}
		save := func(color string) {
			if color == "" {
				delete(windowContextColors, id)
			} else {
				windowContextColors[id] = color
			}
			apply(color)
		}
		menu, _ := gtk.MenuNew()
		menuClass(menu, "glass-context-menu")
		menuClass(menu, "taskbar-color-menu")
		names := []string{"Rot", "Blau", "Gelb", "Grün"}
		for i, color := range tagPalette[:4] {
			color := color
			item, _ := gtk.MenuItemNew()
			menuClass(item, "taskbar-color-choice")
			dot := menuLabel("●", "taskbar-color-dot")
			css, _ := gtk.CssProviderNew()
			css.LoadFromData("label {color: " + color + ";font-size:26px;}")
			sc, _ := dot.GetStyleContext()
			sc.AddProvider(css, gtk.STYLE_PROVIDER_PRIORITY_USER+1)
			item.Add(dot)
			item.SetTooltipText(names[i] + " · erneut auswählen entfernt den Farbring")
			item.Connect("activate", func() {
				if windowContextColors[id] == color {
					save("")
				} else {
					save(color)
				}
			})
			menu.Attach(item, uint(i%4), uint(i%4+1), uint(i/4), uint(i/4+1))
		}
		separator, _ := gtk.SeparatorMenuItemNew()
		menu.Attach(separator, 0, 4, 1, 2)
		close, _ := gtk.MenuItemNewWithLabel("Programm schließen")
		close.Connect("activate", func() { taskbarCommand(id, "kill") })
		menu.Attach(close, 0, 4, 2, 3)
        for _,pinned:=range loadDesktopLayout().Pins {
         if dockAppMatches(pinned,app,title){pinned:=pinned;item,_:=gtk.MenuItemNewWithLabel("Von der Leiste lösen");item.Connect("activate",func(){_ = pinApp(pinned,false)});menu.Attach(item,0,4,3,4);break}
        }
		_ = menu.SetProperty("attach-widget", button.Object)
		menu.ShowAll()
		menu.PopupAtPointer(event)
		menu.Connect("selection-done", func() { menu.Destroy() })
		return true
	})
	return button
}

// A parent blocked by its file picker should reveal that picker when activated.
func modalFocusTarget(tree desktopNode, id int64) int64 {
	var parent desktopNode
	var nodes []desktopNode
	var walk func(desktopNode)
	walk = func(n desktopNode) {
		nodes = append(nodes, n)
		if n.ID == id {
			parent = n
		}
		for _, c := range n.Nodes {
			walk(c)
		}
		for _, c := range n.FloatingNodes {
			walk(c)
		}
	}
	walk(tree)
	if parent.PID <= 0 || parent.Name != "Einstellungen — MiracleOS" {
		return id
	}
	for _, n := range nodes {
		if n.PID == parent.PID && n.Name == "Dateiauswahl — MiracleOS" {
			return n.ID
		}
	}
	return id
}
