/*
	MiracleOS's desktop - A taskbar, with a main menu and a notification daemon
    Copyright (C) 2025 MiracleOS Contributors

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

*/

package main

import (
	_ "embed"
	"fmt"
	"os/exec"
	"time"

	"context"
	"github.com/MiracleOS-Team/desktoplib/batteryHandler"
	"github.com/MiracleOS-Team/desktoplib/foreignToplevel"
	"github.com/MiracleOS-Team/libxdg-go/desktopFiles"
	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"regexp"
	"strconv"
	"strings"
)

//go:embed images/pp.png
var ppData []byte

func getDateInfo() (string, string) {
	hours, minutes, _ := time.Now().Clock()
	curTimeInString := fmt.Sprintf("%d:%02d", hours, minutes)

	curDay := time.Now().Day()
	curDayName := firstN(time.Now().Weekday().String(), 3)
	var curDayCal string

	if curDay <= 5 {
		curMonth := time.Now().Month()
		curDayCal = fmt.Sprintf("%s. %02d %s", curDayName, curDay, curMonth)
	} else {
		curDayCal = fmt.Sprintf("%s. %02d", curDayName, curDay)
	}
	return curDayCal, curTimeInString
}

func createSidestuff() *gtk.Box {
	side, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	side.SetHAlign(gtk.ALIGN_END)
	menuClass(side, "sidestuff")
	statusButton, _ := gtk.ButtonNew()
	quickTray = statusButton
	menuClass(statusButton, "bar-quick-tray")
	statusButton.SetTooltipText("Schnelleinstellungen")
	icons, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 12)
	volume, _ := gtk.ImageNewFromIconName("audio-volume-medium-symbolic", gtk.ICON_SIZE_BUTTON)
	volume.SetPixelSize(16)
	icons.PackStart(volume, false, false, 0)
	pollStatusIcon(volume, readAudioIcon)
	network, _ := gtk.ImageNewFromIconName("network-wired-symbolic", gtk.ICON_SIZE_BUTTON)
	network.SetPixelSize(16)
	icons.PackStart(network, false, false, 0)
	pollStatusIcon(network, func() (string, error) {
		out, err := quickCommand("nmcli", "-t", "-f", "TYPE,STATE", "device")
		if err != nil {
			return "network-offline-symbolic", err
		}
		if strings.Contains(out, "wifi:connected") {
			return "network-wireless-signal-excellent-symbolic", nil
		}
		if strings.Contains(out, "ethernet:connected") {
			return "network-wired-symbolic", nil
		}
		return "network-offline-symbolic", nil
	})
	if batteryHandler.IsBattery() {
		im, _ := gtk.ImageNewFromIconName(batteryHandler.GetBatteryIcon(), gtk.ICON_SIZE_BUTTON)
		im.SetPixelSize(16)
		icons.PackStart(im, false, false, 0)
	}
	statusButton.Add(icons)
	statusButton.Connect("clicked", showQuickSettings)
	side.PackStart(statusButton, false, false, 0)
	dateButton, _ := gtk.ButtonNew()
	menuClass(dateButton, "bar-date-pill")
	dateButton.SetTooltipText("Datum und Uhrzeit")
	dates, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 8)
	clock, _ := gtk.LabelNew(time.Now().Format("15:04"))
	menuClass(clock, "bar-clock")
	date, _ := gtk.LabelNew("")
	menuClass(date, "bar-date")
	dates.PackStart(clock, false, false, 0)
	dates.PackStart(date, false, false, 0)
	dateButton.Add(dates)
	updateDate := func() {
		clock.SetText(time.Now().Format("15:04"))
		days := []string{"So.", "Mo.", "Di.", "Mi.", "Do.", "Fr.", "Sa."}
		date.SetText(days[int(time.Now().Weekday())] + " " + time.Now().Format("02.01."))
	}
	updateDate()
	glib.TimeoutAdd(1000, func() bool { updateDate(); return true })
	dateButton.Connect("clicked", func() { launchSettings() })
	side.PackStart(dateButton, false, false, 0)
	brightness, _ := gtk.ButtonNew()
	menuClass(brightness, "bar-brightness-pill")
	brightness.Add(quickIcon("sun", quickTextColor(), 20))
	brightness.SetTooltipText("Helligkeit und Schnelleinstellungen")
	brightness.Connect("clicked", showQuickSettings)
	side.PackStart(brightness, false, false, 0)
	return side
}

type taskbarEntry struct {
	ID       int64
	Window   foreignToplevel.Toplevel
	IconPath string
}
type taskbarSnapshot struct {
	Windows        []taskbarEntry
	Minimized      []desktopNode
	MinimizedIcons map[int64]string
}

// All compositor queries and icon lookup happen in the worker, never GTK's loop.
func readTaskbarSnapshot(tree desktopNode, apps []desktopFiles.DesktopFile) (taskbarSnapshot, error) {
	result := taskbarSnapshot{Minimized: minimizedWindowsFromTree(tree), MinimizedIcons: map[int64]string{}}
	catalog := installedApplications()
	for _, node := range result.Minimized {
		id := node.AppID
		if id == "" {
			id = node.WindowProperties.Class
		}
		result.MinimizedIcons[node.ID] = taskbarAppIcon(id, node.Name, catalog)
	}
	var walk func(desktopNode)
	walk = func(node desktopNode) {
		if node.Name == "__i3_scratch" {
			return
		}
		if node.PID > 0 && node.Name != "" {
			id := node.AppID
			if id == "" {
				id = node.WindowProperties.Class
			}
			window := foreignToplevel.Toplevel{AppID: id, Title: node.Name}
			result.Windows = append(result.Windows, taskbarEntry{ID: node.ID, Window: window, IconPath: taskbarAppIcon(id, node.Name, catalog)})
		}
		for _, child := range append(node.Nodes, node.FloatingNodes...) {
			walk(child)
		}
	}
	walk(tree)
	return result, nil
}

func createWorkspacesFromSnapshot(snapshot taskbarSnapshot) *gtk.Box {
	pruneWindowColors(snapshot)
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	box.SetHAlign(gtk.ALIGN_START)
	menuClass(box, "workspaces")
	for _, entry := range snapshot.Windows {
		box.PackStart(taskbarWindowButton(entry.ID, entry.Window.AppID, entry.Window.Title, entry.IconPath, false), false, false, 0)
	}
	for _, node := range snapshot.Minimized {
		box.PackStart(taskbarWindowButton(node.ID, node.AppID, node.Name, snapshot.MinimizedIcons[node.ID], true), false, false, 0)
	}
	return box
}
func createMainIcons(apps []desktopFiles.DesktopFile) *gtk.Box {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	box.SetHAlign(gtk.ALIGN_CENTER)

	desktopImage, _ := gtk.ImageNewFromIconName("preferences-system-windows-symbolic", gtk.ICON_SIZE_LARGE_TOOLBAR)
	searchImage, _ := gtk.ImageNewFromIconName("system-search-symbolic", gtk.ICON_SIZE_LARGE_TOOLBAR)

	ppPbuff, _ := gdk.PixbufNewFromBytesOnly(ppData)

	customIcon, _ := gtk.ImageNewFromPixbuf(ppPbuff)
	customButton, _ := gtk.ButtonNew()
	customButton.Add(customIcon)

	mm := createMainMenu(apps)

	customButton.Connect("clicked", func() {
		if mm.IsVisible() {
			mm.Hide()
		} else {
			mm.ShowAll()
		}
		//box.Hide()
	})

	fileButton, _ := gtk.ButtonNew()
	menuClass(fileButton, "bar-launcher")
	fileButton.SetTooltipText("Dateien — MiracleOS")
	fileButton.Add(menuIcon("files", 38))
	fileButton.Connect("clicked", func() { launchFileManager("") })
	_ = desktopImage
	box.PackStart(fileButton, false, false, 0)
	box.PackStart(customButton, false, false, 0)
	searchButton, _ := gtk.ButtonNew()
	menuClass(searchButton, "bar-launcher")
	searchButton.Add(searchImage)
	searchButton.SetTooltipText("Startmenü öffnen")
	searchButton.Connect("clicked", func() { mm.ShowAll() })
	box.PackStart(searchButton, false, false, 0)
	settingsButton, _ := gtk.ButtonNew()
	menuClass(settingsButton, "bar-launcher")
	settingsButton.Add(settingIcon("preferences-system", 30))
	settingsButton.SetTooltipText("Einstellungen — MiracleOS")
	settingsButton.Connect("clicked", launchSettings)
	box.PackStart(settingsButton, false, false, 0)

	return box
}

func createBar(apps []desktopFiles.DesktopFile) *gtk.Window {
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Main Bar")
	win.SetDecorated(false)
	win.SetResizable(false)
	win.SetKeepAbove(true)
	win.SetSkipTaskbarHint(true)
	win.SetSkipPagerHint(true)
	win.SetTypeHint(gdk.WINDOW_TYPE_HINT_DOCK)
	win.SetAppPaintable(true)

	screen, _ := gdk.ScreenGetDefault()
	root_window, _ := screen.GetRootWindow()
	sc_height := root_window.WindowGetHeight()
	sc_width := root_window.WindowGetWidth()
	width := sc_width
	height := 100
	win.SetSizeRequest(width, 0)
	win.Move((sc_width-width)/2, sc_height-height)

	layershell.InitForWindow(win)
	layershell.SetNamespace(win, "miracleos")
	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_LEFT, true)
	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_BOTTOM, true)
	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_RIGHT, true)

	layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_TOP)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_TOP, 0)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_LEFT, 0)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_RIGHT, 0)

	layershell.SetExclusiveZone(win, 50)
	layershell.SetKeyboardMode(win, layershell.LAYER_SHELL_KEYBOARD_MODE_NONE)
	disp, _ := gdk.DisplayGetDefault()
	mon, _ := disp.GetMonitor(0)
	layershell.SetMonitor(win, mon)

	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	sc, _ := box.GetStyleContext()
	sc.AddClass("bar")
	box.SetCenterWidget(createCenterDock(apps))
 box.PackEnd(createSidestuff(), false, false, 0)

	win.Add(box)
	return win
}

func pollStatusIcon(image *gtk.Image, query func() (string, error)) {
	go func() {
		last := ""
		for {
			if icon, err := query(); err == nil && icon != last {
				last = icon
				glib.IdleAdd(func() { image.SetFromIconName(icon, gtk.ICON_SIZE_BUTTON) })
			}
			time.Sleep(2 * time.Second)
		}
	}()
}
func readAudioIcon() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	mute, _ := quickCommand("pactl", "get-sink-mute", "@DEFAULT_SINK@")
	if strings.Contains(mute, "yes") {
		return "audio-volume-muted-symbolic", nil
	}
	data, err := exec.CommandContext(ctx, "pactl", "get-sink-volume", "@DEFAULT_SINK@").Output()
	if err != nil {
		return "", err
	}
	match := regexp.MustCompile(`(\d+)%`).FindStringSubmatch(string(data))
	if len(match) < 2 {
		return "audio-card-symbolic", nil
	}
	volume, _ := strconv.Atoi(match[1])
	if volume == 0 {
		return "audio-volume-muted-symbolic", nil
	}
	if volume >= 85 {
		return "audio-volume-high-symbolic", nil
	}
	if volume >= 25 {
		return "audio-volume-medium-symbolic", nil
	}
	return "audio-volume-low-symbolic", nil
}
