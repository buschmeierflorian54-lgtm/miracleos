package main

import (
	"encoding/json"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"os/exec"
)

type installDisk struct {
	Path  string `json:"path"`
	Label string `json:"label"`
}

func showSystemInstaller() {
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("MiracleOS installieren")
	win.SetDefaultSize(720, 520)
	root, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 18)
	menuClass(root, "setup-card")
	root.PackStart(menuLabel("MiracleOS installieren", "setup-heading"), false, false, 0)
	root.PackStart(settingsText("Installiere MiracleOS auf diesem Gerät. Alle Partitionen und Daten des gewählten Laufwerks werden gelöscht. Andere Laufwerke bleiben erhalten."), false, false, 0)
	disks, _ := gtk.ComboBoxTextNew()
	root.PackStart(disks, false, false, 0)
	status := settingsText("Laufwerke werden gesucht …")
	root.PackStart(status, false, false, 0)
	confirm, _ := gtk.CheckButtonNewWithLabel("Ich möchte das gesamte gewählte Laufwerk löschen.")
	root.PackStart(confirm, false, false, 0)
	var choices []installDisk
	busy, alive := false, true
	win.Connect("delete-event",func()bool{return busy})
	win.Connect("destroy", func() { alive = false; gtk.MainQuit() })
	var install *gtk.Button
	install = settingsButton("Installieren", func() {
		if busy {
			return
		}
		i := disks.GetActive()
		if i < 0 || i >= len(choices) || !confirm.GetActive() {
			status.SetText("Wähle ein Laufwerk und bestätige das Löschen.")
			return
		}
		disk := choices[i]
		dialog := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_WARNING, gtk.BUTTONS_YES_NO, "Letzte Bestätigung: %s vollständig löschen und MiracleOS installieren?", disk.Label)
		menuClass(dialog,"setup-confirmation")
		answer := dialog.Run()
		dialog.Destroy()
		if answer != gtk.RESPONSE_YES {
			return
		}
		busy = true
		install.SetSensitive(false)
		disks.SetSensitive(false)
		confirm.SetSensitive(false)
		status.SetText("MiracleOS wird installiert. Das kann einige Minuten dauern. Bitte nicht ausschalten.")
		go func() {
			_, err := systemHelper("install", map[string]string{"disk": disk.Path, "confirmation": "ERASE"}, "")
			glib.IdleAdd(func() {
				if !alive {
					return
				}
				if err != nil {
					busy = false
					install.SetSensitive(true)
					disks.SetSensitive(true)
					confirm.SetSensitive(true)
					status.SetText("Installation fehlgeschlagen: " + err.Error())
					return
				}
				busy=false
				status.SetText("Installation abgeschlossen. Entferne die ISO und starte neu. Danach beginnt die Ersteinrichtung.")
				root.PackStart(settingsButton("Neu starten", func() { _ = exec.Command("systemctl", "reboot").Start() }), false, false, 0)
				root.ShowAll()
			})
		}()
	})
	install.SetSensitive(false)
	root.PackStart(install, false, false, 0)
	root.PackStart(settingsButton("Spotify im Live-System ausprobieren", launchSpotify), false, false, 0)
	win.Add(root)
	win.ShowAll()
	go func() {
		raw, err := systemHelper("disks", nil, "")
		glib.IdleAdd(func() {
			if !alive {
				return
			}
			if err != nil {
				status.SetText("Laufwerke nicht verfügbar: " + err.Error())
				return
			}
			if json.Unmarshal([]byte(raw), &choices) != nil {
				status.SetText("Ungültige Laufwerksliste.")
				return
			}
			for _, d := range choices {
				disks.AppendText(d.Label)
			}
			if len(choices) > 0 {
				disks.SetActive(0)
				install.SetSensitive(true)
				status.SetText("Wähle das Ziellaufwerk.")
			} else {
				status.SetText("Kein geeignetes Installationslaufwerk gefunden.")
			}
		})
	}()
}
