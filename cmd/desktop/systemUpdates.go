package main

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type systemUpdateStatus struct {
	Installed string `json:"installed"`
	Available string `json:"available"`
	Message   string `json:"message"`
	Checked   string `json:"checked"`
	Automatic bool   `json:"automatic"`
}

func (ui *settingsUI) updates(content *gtk.Box) {
	card := settingsCard("MiracleOS-Updates")
	card.PackStart(settingsText("Neue Versionen werden täglich aus dem offiziellen GitHub-Repository geladen und vor der Installation auf ihre Signatur geprüft. Laufende Apps bleiben geöffnet. Melde dich nach einem Update erneut an, um die neue Version zu nutzen."), false, false, 0)
	label := settingsText("")
	card.PackStart(label, false, false, 0)
	content.PackStart(card, false, false, 0)
	view := ui.view
	refresh := func() bool {
		if !ui.alive || ui.view != view {
			return false
		}
		raw, err := os.ReadFile("/var/lib/miracleos/updates/status.json")
		var state systemUpdateStatus
		if err != nil || json.Unmarshal(raw, &state) != nil {
			label.SetText("Die Update-Funktion ist auf dieser Installation noch nicht eingerichtet.")
			return true
		}
		automatic := "Pausiert"
		if state.Automatic {
			automatic = "Aktiviert"
		}
		text := "Installierte Version: " + state.Installed + "\nAutomatische Updates: " + automatic + "\n\n" + state.Message
		if state.Available != "" {
			text += "\nVersion: " + state.Available
		}
		if state.Checked != "" {
			text += "\nLetzte Prüfung: " + state.Checked
		}
		label.SetText(text)
		return true
	}
	refresh()
	glib.TimeoutAdd(2000, refresh)
	for _, item := range []struct{ label, action string }{
		{"Jetzt prüfen und aktualisieren …", "apply"},
		{"Automatische Updates aktivieren …", "enable"},
		{"Automatische Updates pausieren …", "disable"},
		{"Vorherige Version wiederherstellen …", "rollback"},
	} {
		action, title := item.action, item.label
		card.PackStart(settingsButton(title, func() { ui.updateActionDialog(action) }), false, false, 0)
	}
	content.PackStart(settingsText("MiracleOS aktualisiert hier Desktop, Browser, Fenstersystem und zugehörige Laufzeitdateien. Debian-Paketupdates werden weiterhin über die Debian-Paketverwaltung gepflegt. Persönliche Dateien bleiben erhalten."), false, false, 0)
}

func (ui *settingsUI) updateActionDialog(action string) {
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Systemupdates")
	win.SetTransientFor(ui.win)
	win.SetModal(true)
	win.SetDefaultSize(440, 280)
	root, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	menuClass(root, "fm-shell")
	menuClass(root, "glass-dialog")
	root.PackStart(glassDialogHeader(win, "Systemupdates"), false, false, 0)
	body, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	menuClass(body, "fm-dialog")
	body.PackStart(settingsText("Bestätige diese Änderung mit deinem MiracleOS-Passwort."), false, false, 0)
	password, _ := gtk.EntryNew()
	password.SetVisibility(false)
	password.SetInputPurpose(gtk.INPUT_PURPOSE_PASSWORD)
	password.SetPlaceholderText("MiracleOS-Passwort")
	body.PackStart(password, false, false, 0)
	result := settingsText("")
	body.PackStart(result, false, false, 0)
	alive, busy := true, false
	win.Connect("destroy", func() { alive = false })
	button, _ := gtk.ButtonNewWithLabel("Bestätigen")
	menuClass(button, "settings-button")
	button.Connect("clicked", func() {
		if busy {
			return
		}
		secret, _ := password.GetText()
		if secret == "" || strings.ContainsAny(secret, "\r\n\x00") {
			result.SetText("Bitte dein Passwort eingeben.")
			return
		}
		busy = true
		password.SetText("")
		password.SetSensitive(false)
		button.SetSensitive(false)
		result.SetText("Änderung wird ausgeführt …")
		go func() {
			_, err := systemHelper("update", map[string]string{"action": action}, secret)
			glib.IdleAdd(func() {
				if !alive {
					return
				}
				if err != nil {
					result.SetText("Änderung fehlgeschlagen. Prüfe dein Passwort und den Update-Status.")
					busy = false
					password.SetSensitive(true)
					button.SetSensitive(true)
				} else {
					win.Destroy()
				}
			})
		}()
	})
	body.PackStart(button, false, false, 0)
	root.PackStart(body, true, true, 0)
	win.Add(root)
	win.ShowAll()
}
