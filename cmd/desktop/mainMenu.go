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
	"fmt"

	"os/user"
	"sort"

	"github.com/MiracleOS-Team/libxdg-go/desktopFiles"
	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func createAppGroup(apps []desktopFiles.DesktopFile) *gtk.Box {
	group, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
	for _, app := range apps {
		buttonBox, _ := gtk.ButtonNew()
		appBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
		sc, _ := appBox.GetStyleContext()
		sc.AddClass("mm_applist_app")

		pixbuf, err := gdk.PixbufNewFromFile(app.Icon)
		if err == nil {

			maxWidth := 16
			maxHeight := 16

			scaledPixbuf := scalePixbuf(pixbuf, maxWidth, maxHeight)

			icon, _ := gtk.ImageNewFromPixbuf(scaledPixbuf)
			appBox.PackStart(icon, false, false, 5)
		}

		label, _ := gtk.LabelNew(app.Name)

		appBox.PackStart(label, false, false, 5)
		buttonBox.Add(appBox)
		buttonBox.Connect("button-press-event", func() {
			fmt.Println("Clicked on", app.Name)
			hideStartMenu()
			go desktopFiles.ExecuteDesktopFile(app, []string{}, "")
		})
		group.PackStart(buttonBox, false, false, 5)

	}
	return group
}

func createAppList(apps1 []desktopFiles.DesktopFile) *gtk.ScrolledWindow {
	scroll, _ := gtk.ScrolledWindowNew(nil, nil)
	scroll.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	vbox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)

	apps := []desktopFiles.DesktopFile{}

	for _, app := range apps1 {
		if app.NoDisplay {
			continue
		} else {
			apps = append(apps, app)
		}
	}

	categories := map[string][]desktopFiles.DesktopFile{}

	for _, app := range apps {
		if _, ok := categories[firstN(app.Name, 1)]; !ok {
			categories[firstN(app.Name, 1)] = []desktopFiles.DesktopFile{}
		}
		categories[firstN(app.Name, 1)] = append(categories[firstN(app.Name, 1)], app)
	}

	// Sort categories alphabetically
	sortedCategories := make([]string, 0, len(categories))
	for category := range categories {
		sortedCategories = append(sortedCategories, category)
	}
	sort.Strings(sortedCategories)

	for _, category := range sortedCategories {
		label, _ := gtk.LabelNew(category)
		label.SetMarkup("<b>" + category + "</b>")
		label.SetXAlign(0)
		vbox.PackStart(label, false, false, 5)

		// Sort applications within each category
		appList := categories[category]
		sortedAppNames := make([]string, 0, len(appList))
		for _, app := range appList {
			sortedAppNames = append(sortedAppNames, app.Name)
		}
		sort.Strings(sortedAppNames)

		sortedApps := make([]desktopFiles.DesktopFile, 0, len(appList))
		for _, appName := range sortedAppNames {
			for _, app := range appList {
				if app.Name == appName {
					sortedApps = append(sortedApps, app)
				}
			}
		}

		vbox.PackStart(createAppGroup(sortedApps), false, false, 5)
	}

	scroll.Add(vbox)
	return scroll
}

func createUserInfo() *gtk.Box {
	userObject, err := user.Current()
	if err != nil {
		panic(err)
	}

	subtitle := "Standard User"

	groups, err := userObject.GroupIds()
	if err == nil {
		for _, groupId := range groups {
			gp, err := user.LookupGroupId(groupId)
			if err != nil {
				continue
			}
			if gp.Name == "sudo" {
				subtitle = "Administrator"
			}
			if gp.Name == "wheel" {
				subtitle = "Administrator"
			}

		}

	}

	userBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	avatar, _ := gdk.PixbufNewFromBytesOnly(ppData)
	userImage, _ := gtk.ImageNewFromPixbuf(scalePixbuf(avatar, 40, 40))
	userNameBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)

	userNameLabel, _ := gtk.LabelNew(cases.Title(language.English, cases.Compact).String(userObject.Username))
	userNameLabel.SetXAlign(0)
	sc, _ := userNameLabel.GetStyleContext()
	sc.AddClass("h2")

	userSubtitleLabel, _ := gtk.LabelNew(subtitle)
	userSubtitleLabel.SetXAlign(0)

	userNameBox.PackStart(userNameLabel, false, false, 0)
	userNameBox.PackEnd(userSubtitleLabel, false, false, 0)
	userBox.PackStart(userImage, false, false, 5)
	userBox.PackStart(userNameBox, false, false, 5)
	return userBox
}

func createPowerButtons() *gtk.Box {
	powerBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	settingsBtn, _ := gtk.ButtonNewWithLabel("⚙")
	settingsBtn.SetTooltipText("Einstellungen")
	settingsBtn.Connect("clicked", launchSettings)
	powerBox.PackStart(settingsBtn, false, false, 5)
	shutdownBtn, _ := gtk.ButtonNew()
	shutdownBtn.Add(fmIcon("system-shutdown-symbolic", gtk.ICON_SIZE_BUTTON))
	shutdownBtn.SetTooltipText("Herunterfahren, Neustarten oder Sperren")
	shutdownBtn.Connect("clicked", func() {
		menu := buildSharedMenu([]contextAction{
			{"Herunterfahren", func() { sessionAction("poweroff") }, false},
			{"Neustarten", func() { sessionAction("reboot") }, false},
			{"Sperren", func() { sessionAction("lock") }, false},
		})
		menu.ShowAll()
		menu.PopupAtPointer(nil)
		menu.Connect("selection-done", func() { menu.Destroy() })
	})
	powerBox.PackEnd(shutdownBtn, false, false, 5)
	return powerBox
}

func createMainMenu(apps []desktopFiles.DesktopFile) *gtk.Window {return createFreeStart()}
func createLegacyMainMenu(apps []desktopFiles.DesktopFile) *gtk.Window {
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	activeMainMenu = win
	dismissPopupOutside(win, func() { win.Hide() })
	win.SetTitle("Main Menu")

	win.SetDecorated(false)
	win.SetResizable(false)
	win.SetTypeHint(gdk.WINDOW_TYPE_HINT_DOCK)

	layershell.InitForWindow(win)
	layershell.SetNamespace(win, "miracleos")
	layershell.SetKeyboardMode(win, layershell.LAYER_SHELL_KEYBOARD_MODE_EXCLUSIVE)

	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_BOTTOM, true)

	layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_OVERLAY)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_TOP, 0)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_LEFT, 0)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_RIGHT, 0)

	disp, _ := gdk.DisplayGetDefault()
	mon, _ := disp.GetMonitor(0)
	layershell.SetMonitor(win, mon)

	mainBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	sc, _ := mainBox.GetStyleContext()
	sc.AddClass("mm_menu_m2")

	topBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	sc, _ = topBox.GetStyleContext()
	sc.AddClass("mm_toppart")
	topBox.PackStart(createQuickAppsHeader(), false, false, 0)

	searchEntry, _ := gtk.EntryNew()
	searchEntry.SetPlaceholderText("Apps durchsuchen …")
	sc, _ = searchEntry.GetStyleContext()
	sc.AddClass("mos-input")
	searchEntry.SetHAlign(gtk.ALIGN_CENTER)
	searchEntry.SetSizeRequest(100, -1)
	searchEntry.SetWidthChars(24)
	topBox.PackStart(searchEntry, true, true, 5)
	topBox.PackEnd(menuLabel("MiracleOS", "menu-wordmark"), false, false, 0)

	contentBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)

	appList := installedAppsView(win, searchEntry)
	sc, _ = appList.GetStyleContext()
	sc.AddClass("mm_tab")

	fastApps := createWidgetPanel()
	sc, _ = fastApps.GetStyleContext()
	sc.AddClass("mm_tab")

	otherTab := createAppPanel()
	sc, _ = otherTab.GetStyleContext()
	sc.AddClass("mm_tab")

	appList.SetSizeRequest(250, 420)
	fastApps.SetSizeRequest(280, 420)
	otherTab.SetSizeRequest(280, 420)

	contentBox.PackStart(appList, false, false, 0)
	contentBox.PackStart(fastApps, false, false, 0)
	contentBox.PackStart(otherTab, false, false, 0)

	mainBox.PackStart(topBox, true, true, 10)
	mainBox.PackStart(contentBox, true, true, 10)

	userInfo := createUserInfo()
	sc, _ = userInfo.GetStyleContext()
	sc.AddClass("mm_profileinfo")

	powerButtons := createPowerButtons()
	sc, _ = userInfo.GetStyleContext()
	sc.AddClass("mm_managingicons")

	bottomBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	bottomBox.PackStart(userInfo, false, false, 10)
	bottomBox.PackEnd(powerButtons, false, false, 10)
	sc, _ = bottomBox.GetStyleContext()
	sc.AddClass("mm_bottompart")
	mainBox.PackStart(bottomBox, false, false, 10)

	win.Add(mainBox)
	return win
}

func createPlaceholder(name string) *gtk.Box {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	label, _ := gtk.LabelNew(name + "\nThis functionality isn't available for now")
	box.PackStart(label, false, false, 10)
	return box
}
