package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (ui *settingsUI) passwordDialog() {
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Passwort ändern")
	win.SetTransientFor(ui.win)
	win.SetModal(true)
	win.SetDefaultSize(420, 360)
	root, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	menuClass(root, "fm-shell")
	menuClass(root, "glass-dialog")
	root.PackStart(glassDialogHeader(win, "Passwort ändern"), false, false, 0)
	body, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	menuClass(body, "fm-dialog")
	body.PackStart(menuLabel("Dein Passwort ändern", "dialog-heading"), false, false, 0)
	entries := []*gtk.Entry{}
	for _, name := range []string{"Aktuelles Passwort", "Neues Passwort", "Neues Passwort bestätigen"} {
		e, _ := gtk.EntryNew()
		e.SetVisibility(false)
		e.SetInputPurpose(gtk.INPUT_PURPOSE_PASSWORD)
		e.SetPlaceholderText(name)
		menuClass(e, "fm-path")
		body.PackStart(e, false, false, 0)
		entries = append(entries, e)
	}
	status := settingsText("")
	body.PackStart(status, false, false, 0)
	alive, busy := true, false
	win.Connect("destroy", func() { alive = false })
	var save *gtk.Button
	save = dialogAction("Passwort ändern", "dialog-password", "", true, func() {
		if busy {
			return
		}
		current, _ := entries[0].GetText()
		next, _ := entries[1].GetText()
		repeat, _ := entries[2].GetText()
		if current == "" || next == "" {
			status.SetText("Bitte beide Passwörter eingeben.")
			return
		}
		if next != repeat {
			status.SetText("Die neuen Passwörter stimmen nicht überein.")
			return
		}
		if current == next {
			status.SetText("Bitte ein anderes neues Passwort wählen.")
			return
		}
		if strings.ContainsAny(current+next, "\r\n\x00") {
			status.SetText("Dieses Passwort enthält ungültige Zeichen.")
			return
		}
		busy = true
		save.SetSensitive(false)
		status.SetText("Passwort wird geändert …")
		for _, e := range entries {
			e.SetText("")
			e.SetSensitive(false)
		}
		go func() {
			data, _ := json.Marshal(map[string]string{"current": current, "new": next})
			binary, _ := os.Executable()
			ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "python3", filepath.Join(filepath.Dir(binary), "runtime", "changePassword.py"))
			cmd.Stdin = bytes.NewReader(data)
			err := cmd.Run()
			glib.IdleAdd(func() {
				if !alive {
					return
				}
				busy = false
				for _, e := range entries {
					e.SetSensitive(true)
				}
				if err != nil {
					save.SetSensitive(true)
					status.SetText("Ändern fehlgeschlagen. Prüfe dein aktuelles Passwort und wähle ein ausreichend starkes neues Passwort.")
					entries[0].GrabFocus()
				} else {
					status.SetText("Passwort geändert. Es gilt jetzt auch für Anmeldung und Sperrbildschirm.")
					save.SetSensitive(false)
				}
			})
		}()
	})
	body.PackStart(save, false, false, 0)
	body.PackStart(dialogAction("Schließen", "", "Esc", false, func() { win.Destroy() }), false, false, 0)
	root.PackStart(body, true, true, 0)
	win.Add(root)
	dialogEscape(win)
	win.ShowAll()
	entries[0].GrabFocus()
}
