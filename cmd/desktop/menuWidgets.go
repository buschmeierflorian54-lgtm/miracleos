package main

import (
	"embed"
	"fmt"
	"os/exec"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

//go:embed images/menu-icons/*.png
var menuIconFiles embed.FS

var activeMainMenu *gtk.Window

func hideStartMenu() {
	if activeMainMenu != nil {
		activeMainMenu.Hide()
	}
}

type menuApp struct {
	id, name, caption, description string
}

var menuApps = map[string]menuApp{
	"river":    {"river", "River", "River", "Dein Browser für neue Ideen. Die River-Oberfläche ist als Vorschau vorbereitet."},
	"store":    {"store", "Miracle Store", "Miracle\nStore", "Ein Platz zum Entdecken neuer Apps. Diese Ansicht ist eine Vorschau; es werden keine Apps installiert."},
	"media":    {"media", "Multimedia Player", "Multimedia\nPlayer", "Musik, Filme und Lieblingsmomente an einem Ort. Die Medienbibliothek ist als Vorschau vorbereitet."},
	"chrome":   {"chrome", "Chrome", "Chrome", "Diese Chrome-Kachel ist eine Vorschau. Zum Surfen kannst du den bereits eingerichteten Browser River öffnen."},
	"mail":     {"mail", "Mail", "Mail", "Dein Posteingang an einem Ort. Diese Vorschau ist mit keinem E-Mail-Konto verbunden."},
	"calendar": {"calendar", "Calendar", "Calendar", "Dein Kalender für heute und die nächsten Tage."},
	"news":     {"news", "News", "News", "Nachrichten, die dich interessieren. Der Nachrichtenbereich ist noch eine Vorschau ohne Live-Meldungen."},
	"weather":  {"weather", "Weather", "Weather", "Berlin · 27 °C · leicht bewölkt\n\nDie Wetterkarte zeigt Beispieldaten. Ein Wetterdienst ist noch nicht verbunden."},
	"conchat":  {"conchat", "Conchat", "Conchat", "Ein gemeinsamer Platz für Gespräche. Diese Vorschau sendet keine Nachrichten."},
	"meetnow":  {"meetnow", "Meetnow", "Meetnow", "Ein Klick zu deinem nächsten Treffen. Diese Vorschau startet keine Kamera und kein Gespräch."},
	"clock":    {"clock", "Clock", "Clock", "Die aktuelle Zeit auf deinem Desktop."},
	"notes":    {"notes", "Xtranote", "Xtranote", "Platz für Notizen und spontane Ideen."},
}

func menuClass(widget gtk.IWidget, name string) {
	context, _ := widget.ToWidget().GetStyleContext()
	context.AddClass(name)
}

func menuLabel(text, class string) *gtk.Label {
	label, _ := gtk.LabelNew(text)
	label.SetXAlign(0)
	if class != "" {
		menuClass(label, class)
	}
	return label
}

func menuIcon(id string, size int) *gtk.Image {
	data, err := menuIconFiles.ReadFile("images/menu-icons/" + id + ".png")
	if err == nil {
		pixbuf, loadErr := gdk.PixbufNewFromBytesOnly(data)
		if loadErr == nil {
			scaled, scaleErr := pixbuf.ScaleSimple(size, size, gdk.INTERP_BILINEAR)
			if scaleErr == nil {
				image, _ := gtk.ImageNewFromPixbuf(scaled)
				return image
			}
		}
	}
	image, _ := gtk.ImageNewFromIconName("application-x-executable", gtk.ICON_SIZE_DIALOG)
	return image
}

func activateMenuApp(id string) {
	hideStartMenu()
	command := ""
	if id == "notes" {
		command = "mousepad"
	}
	if command != "" {
		if err := exec.Command(command).Start(); err == nil {
			return
		}
	}
	showMenuApp(id)
}

func showMenuApp(id string) {
	hideStartMenu()
	app := menuApps[id]
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle(app.name)
	win.SetDefaultSize(380, 340)
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 14)
	menuClass(box, "demo-window")
	icon := menuIcon(id, 84)
	icon.SetHAlign(gtk.ALIGN_CENTER)
	box.PackStart(icon, false, false, 0)
	heading := menuLabel(app.name, "section-heading")
	heading.SetHAlign(gtk.ALIGN_CENTER)
	box.PackStart(heading, false, false, 0)
	if id == "calendar" {
		calendar, _ := gtk.CalendarNew()
		box.PackStart(calendar, true, true, 0)
	} else if id == "clock" {
		clock := menuLabel(time.Now().Format("15:04:05"), "weather-temperature")
		clock.SetHAlign(gtk.ALIGN_CENTER)
		box.PackStart(clock, true, true, 0)
		alive := true
		win.Connect("destroy", func() { alive = false })
		glib.TimeoutAdd(1000, func() bool {
			if !alive {
				return false
			}
			clock.SetText(time.Now().Format("15:04:05"))
			return true
		})
	} else {
		body := menuLabel(app.description, "demo-description")
		body.SetLineWrap(true)
		body.SetMaxWidthChars(34)
		box.PackStart(body, true, true, 0)
		box.PackStart(menuLabel("App-Vorschau", "muted-caption"), false, false, 0)
	}
	close, _ := gtk.ButtonNewWithLabel("Schließen")
	menuClass(close, "soft-button")
	close.Connect("clicked", func() { win.Destroy() })
	box.PackEnd(close, false, false, 0)
	win.Add(box)
	win.ShowAll()
}

func menuAppButton(id string) *gtk.Button {
	app := menuApps[id]
	button, _ := gtk.ButtonNew()
	button.SetTooltipText(app.name)
	menuClass(button, "menu-app-tile")
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 2)
	icon := menuIcon(id, 52)
	icon.SetHAlign(gtk.ALIGN_CENTER)
	box.PackStart(icon, false, false, 0)
	label := menuLabel(app.caption, "app-caption")
	label.SetXAlign(0.5)
	label.SetHAlign(gtk.ALIGN_CENTER)
	box.PackStart(label, false, false, 0)
	button.Add(box)
	button.SetSizeRequest(56, 76)
	button.Connect("clicked", func() { activateMenuApp(id) })
	return button
}

func menuAppGrid(ids []string, columns int) *gtk.Grid {
	grid, _ := gtk.GridNew()
	grid.SetColumnSpacing(4)
	grid.SetRowSpacing(6)
	grid.SetColumnHomogeneous(true)
	for index, id := range ids {
		grid.Attach(menuAppButton(id), index%columns, index/columns, 1, 1)
	}
	if len(ids) < columns {
		spacer, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
		spacer.SetSizeRequest(56, 76)
		grid.Attach(spacer, columns-1, 0, 1, 1)
	}
	return grid
}

func createAppPanel() *gtk.Box {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 8)
	menuClass(box, "app-panel")
	box.PackStart(menuLabel("Most Used", "section-heading"), false, false, 0)
	box.PackStart(menuAppGrid([]string{"river", "store", "media"}, 3), false, false, 0)
	box.PackStart(menuLabel("Recently added", "section-heading"), false, false, 0)
	box.PackStart(menuAppGrid([]string{"chrome"}, 4), false, false, 0)
	box.PackStart(menuLabel("Productivity", "section-heading"), false, false, 0)
	box.PackStart(menuAppGrid([]string{"mail", "calendar", "news", "weather", "conchat", "meetnow", "clock", "notes"}, 4), false, false, 0)
	return box
}

func createQuickAppsHeader() *gtk.Box {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	for _, id := range []string{"river", "store", "media"} {
		id := id
		button, _ := gtk.ButtonNew()
		button.SetTooltipText(menuApps[id].name)
		menuClass(button, "quick-app-button")
		button.Add(menuIcon(id, 34))
		button.Connect("clicked", func() { activateMenuApp(id) })
		box.PackStart(button, false, false, 0)
	}
	return box
}

func createWeatherWidget() *gtk.Button {
	button, _ := gtk.ButtonNew()
	menuClass(button, "weather-card")
	button.SetTooltipText("Wettervorschau mit Beispieldaten")
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	top, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	text, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	text.PackStart(menuLabel("27°", "weather-temperature"), false, false, 0)
	text.PackStart(menuLabel("Berlin · Beispieldaten", "muted-caption"), false, false, 0)
	top.PackStart(text, true, true, 0)
	top.PackEnd(menuIcon("weather", 64), false, false, 0)
	box.PackStart(top, false, false, 0)
	forecast, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	for index, day := range []string{"Mo", "Di", "Mi", "Do", "Fr", "Sa", "So"} {
		cell, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
		label := menuLabel(day, "forecast-caption")
		label.SetXAlign(0.5)
		cell.PackStart(label, false, false, 0)
		if index == 2 || index == 3 || index == 6 {
			cell.PackStart(menuIcon("weather", 22), false, false, 0)
		} else {
			sun := menuLabel("☀", "forecast-sun")
			sun.SetXAlign(0.5)
			sun.SetSizeRequest(-1, 22)
			cell.PackStart(sun, false, false, 0)
		}
		degree := menuLabel([]string{"24°", "29°", "26°", "27°", "28°", "31°", "28°"}[index], "forecast-caption")
		degree.SetXAlign(0.5)
		cell.PackStart(degree, false, false, 0)
		forecast.PackStart(cell, true, true, 0)
	}
	box.PackStart(forecast, false, false, 0)
	button.Add(box)
	button.Connect("clicked", func() { showMenuApp("weather") })
	return button
}

func createCalendarWidget() *gtk.Button {
	button, _ := gtk.ButtonNew()
	menuClass(button, "calendar-card")
	button.SetTooltipText("Kalender öffnen · Termine sind Beispieldaten")
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 8)
	now := time.Now()
	row, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 8)
	date := menuLabel(fmt.Sprintf("%02d", now.Day()), "calendar-day")
	row.PackStart(date, false, false, 0)
	weekday := []string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"}[now.Weekday()]
	month := []string{"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"}[int(now.Month())-1]
	row.PackStart(menuLabel(weekday+"\n"+month, "calendar-caption"), false, false, 0)
	box.PackStart(row, false, false, 0)
	box.PackStart(menuLabel("●  Projektideen\n     10:00", "agenda-pink"), false, false, 0)
	box.PackStart(menuLabel("●  Zeit für Neues\n     15:30", "agenda-blue"), false, false, 0)
	box.PackEnd(menuLabel("Beispieltermine", "tiny-caption"), false, false, 0)
	button.Add(box)
	button.Connect("clicked", func() { showMenuApp("calendar") })
	return button
}

func createMusicWidget() *gtk.Box {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 8)
	menuClass(box, "music-card")
	art := menuIcon("media", 50)
	art.SetHAlign(gtk.ALIGN_CENTER)
	box.PackStart(art, false, false, 0)
	tracks := []string{"Hero", "Aurora", "Nightdrive"}
	track := 0
	title := menuLabel(tracks[track], "music-title")
	title.SetXAlign(0.5)
	box.PackStart(title, false, false, 0)
	caption := menuLabel("Musikvorschau", "tiny-caption")
	caption.SetXAlign(0.5)
	box.PackStart(caption, false, false, 0)
	progress, _ := gtk.ProgressBarNew()
	menuClass(progress, "music-progress")
	progress.SetFraction(0.28)
	box.PackStart(progress, false, false, 0)
	controls, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	controls.SetHAlign(gtk.ALIGN_CENTER)
	previous, _ := gtk.ButtonNewWithLabel("‹")
	play, _ := gtk.ButtonNewWithLabel("▶")
	next, _ := gtk.ButtonNewWithLabel("›")
	menuClass(play, "play-button")
	menuClass(previous, "music-control")
	menuClass(next, "music-control")
	play.SetTooltipText("Wiedergabevorschau · kein Audio verbunden")
	playing := false
	position := 0.28
	play.Connect("clicked", func() {
		playing = !playing
		if playing {
			play.SetLabel("Ⅱ")
		} else {
			play.SetLabel("▶")
		}
	})
	changeTrack := func(delta int) {
		track = (track + delta + len(tracks)) % len(tracks)
		title.SetText(tracks[track])
		position = 0
		progress.SetFraction(0)
	}
	previous.Connect("clicked", func() { changeTrack(-1) })
	next.Connect("clicked", func() { changeTrack(1) })
	controls.PackStart(previous, false, false, 0)
	controls.PackStart(play, false, false, 0)
	controls.PackStart(next, false, false, 0)
	box.PackStart(controls, false, false, 0)
	alive := true
	box.Connect("destroy", func() { alive = false })
	glib.TimeoutAdd(1000, func() bool {
		if !alive {
			return false
		}
		if playing {
			position += 1.0 / 180
			if position >= 1 {
				changeTrack(1)
			}
			progress.SetFraction(position)
		}
		return true
	})
	return box
}

func createWidgetPanel() *gtk.Box {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	menuClass(box, "widget-panel")
	box.PackStart(menuLabel("Dein Tag", "section-heading"), false, false, 0)
	weather := createWeatherWidget()
	box.PackStart(weather, false, false, 0)
	tiles, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 8)
	calendar := createCalendarWidget()
	music := createMusicWidget()
	calendar.SetSizeRequest(122, 170)
	music.SetSizeRequest(122, 170)
	tiles.PackStart(calendar, true, true, 0)
	tiles.PackStart(music, true, true, 0)
	box.PackStart(tiles, false, false, 0)
	add, _ := gtk.ButtonNewWithLabel("+")
	menuClass(add, "add-widget-button")
	add.SetHAlign(gtk.ALIGN_START)
	add.SetTooltipText("Widgets verwalten")
	add.Connect("clicked", func() {
		hideStartMenu()
		win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
		win.SetTitle("Widgets")
		settings, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
		menuClass(settings, "demo-window")
		settings.PackStart(menuLabel("Deine Widgets", "section-heading"), false, false, 0)
		widgets := []struct {
			name   string
			widget gtk.IWidget
		}{{"Wetter", weather}, {"Kalender", calendar}, {"Musik", music}}
		for _, item := range widgets {
			widget := item.widget
			toggle, _ := gtk.CheckButtonNewWithLabel(item.name)
			toggle.SetActive(widget.ToWidget().GetVisible())
			toggle.Connect("toggled", func() { widget.ToWidget().SetVisible(toggle.GetActive()) })
			settings.PackStart(toggle, false, false, 0)
		}
		win.Add(settings)
		win.ShowAll()
	})
	box.PackEnd(add, false, false, 0)
	return box
}
