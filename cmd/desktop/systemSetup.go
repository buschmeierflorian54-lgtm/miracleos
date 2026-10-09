package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"os"
	"os/exec"
	"strings"
	"time"
)

func systemHelper(action string, data interface{}, password string) (string, error) {
	raw, _ := json.Marshal(data)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sudo", "-S", "-k", "-p", "", "/usr/local/lib/miracleos/system-helper.py", action)
	input := string(raw)
	if password != "" {
		input = password + "\n" + input
	}
	cmd.Stdin = strings.NewReader(input)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(output.String()))
	}
	return strings.TrimSpace(output.String()), nil
}
func systemSetupPending() bool {
	if _, err := os.Stat("/run/live/medium/live/filesystem.squashfs"); err == nil {
		return false
	}
	_, err := os.Stat("/var/lib/miracleos/setup-pending")
	return err == nil
}
func launchSpotify() { _ = startSetupSpotify() }

func showSystemSetup(preview bool) {
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Ersteinrichtung — MiracleOS")
	win.SetDefaultSize(1100, 760)
	win.SetDecorated(false)
	win.Fullscreen()
	root, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	menuClass(root, "setup-background")
	card, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 18)
	card.SetHAlign(gtk.ALIGN_CENTER)
	card.SetVAlign(gtk.ALIGN_CENTER)
	card.SetSizeRequest(540, -1)
	menuClass(card, "setup-card")
	root.PackStart(card, true, true, 0)
	brand := menuLabel("MiracleOS · Schritt 1 von 4", "setup-brand")
	card.PackStart(brand, false, false, 0)
	heading := menuLabel("Willkommen. Wie heißt du?", "setup-heading")
	heading.SetLineWrap(true)
	heading.SetMaxWidthChars(28)
	heading.SetXAlign(0)
	card.PackStart(heading, false, false, 0)
	hint := settingsText("So begrüßen wir dich auf diesem Gerät.")
	card.PackStart(hint, false, false, 0)
	name, _ := gtk.EntryNew()
	name.SetPlaceholderText("Dein Name")
	if raw, err := os.ReadFile("/var/lib/miracleos/setup-account.json"); err == nil {
		var pending struct { Name string `json:"name"` }
		if json.Unmarshal(raw, &pending) == nil { name.SetText(pending.Name) }
	}
	menuClass(name, "setup-entry")
	card.PackStart(name, false, false, 0)
	password, _ := gtk.EntryNew()
	password.SetVisibility(false)
	password.SetInputPurpose(gtk.INPUT_PURPOSE_PASSWORD)
	password.SetPlaceholderText("Passwort (mindestens 8 Zeichen)")
	menuClass(password, "setup-entry")
	card.PackStart(password, false, false, 0)
	repeat, _ := gtk.EntryNew()
	repeat.SetVisibility(false)
	repeat.SetInputPurpose(gtk.INPUT_PURPOSE_PASSWORD)
	repeat.SetPlaceholderText("Passwort wiederholen")
	menuClass(repeat, "setup-entry")
	card.PackStart(repeat, false, false, 0)
	current, _ := gtk.EntryNew()
	current.SetVisibility(false)
	current.SetPlaceholderText("Aktuelles Passwort für erneute Einrichtung")
	menuClass(current, "setup-entry")
	card.PackStart(current, false, false, 0)
	music, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 14)
	card.PackStart(music, false, false, 0)
	music.PackStart(settingsText("Mit Alt + F öffnest du deine Musiksteuerung über dem Desktop: pausieren, weiterschalten und die Lautstärke ändern, ohne deine Arbeit zu verlassen. Im Spotify-Widget siehst du Titel und Cover. Über das Dock erreichst du Spotify jederzeit."), false, false, 0)
	music.PackStart(settingsText("Möchtest du Spotify jetzt einrichten? Melde dich in der Spotify-App an und probiere einen Song aus."), false, false, 0)
	player, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	addCanvasMusic(player)
	card.PackStart(player, false, false, 0)
	spotifyPlayed := false
	wifi, hasWiFi := newSetupWiFi()
	card.PackStart(wifi, false, false, 0)
	status := settingsText("")
	card.PackStart(status, false, false, 0)
	nav, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 12)
	card.PackStart(nav, false, false, 0)
	step, busy := 0, false
	if hasWiFi { step = -1 }
	if preview && len(os.Args) > 2 && os.Args[2] == "spotify" {
		step = 2
	}
	alive := true
	win.Connect("destroy", func() { alive = false })
	var next, back, skip *gtk.Button
	var render func()
	render = func() {
		number,total := step+1,4
		if hasWiFi { number,total=step+2,5 }
		brand.SetText(fmt.Sprintf("MiracleOS · Schritt %d von %d", number,total))
		for _, w := range []gtk.IWidget{name, password, repeat, current, music, player, wifi, back, skip} {
			w.ToWidget().SetNoShowAll(true)
			w.ToWidget().Hide()
		}
		show := func(w gtk.IWidget) {
			w.ToWidget().SetNoShowAll(false)
			w.ToWidget().ShowAll()
			w.ToWidget().SetNoShowAll(true)
		}
		next.SetLabel("Weiter")
		next.SetSensitive(!busy)
		if (step > 0 || (hasWiFi && step == 0)) && step < 3 {
			show(back)
		}
		switch step {
		case -1:
			heading.SetText("Verbinde dich mit deinem WLAN.")
			hint.SetText("Wähle dein Netzwerk. Du kannst WLAN auch später einrichten.")
			show(wifi)
		case 0:
			heading.SetText("Willkommen. Wie heißt du?")
			hint.SetText("So begrüßen wir dich auf diesem Gerät.")
			show(name)
		case 1:
			heading.SetText("Wähle dein Passwort.")
			hint.SetText("Damit meldest du dich an und bestätigst Änderungen am System.")
			show(password)
			show(repeat)
			if !systemSetupPending() && !preview {
				show(current)
			}
		case 2:
			heading.SetText("Deine Musik. Überall dabei.")
			hint.SetText("Spotify ist in MiracleOS bereits installiert.")
			show(music)
			show(skip)
			next.SetLabel("Ja, jetzt einrichten")
		case 3:
			heading.SetText("Alles bereit.")
			hint.SetText("Deine Musik steuerst du mit Alt + F. Willkommen bei MiracleOS.")
			next.SetLabel("MiracleOS starten")
			if spotifyPlayed { show(player) }
		}
	}
	back = settingsButton("Zurück", func() {
		if busy {
			return
		}
		step--
		status.SetText("")
		render()
	})
	nav.PackStart(back, false, false, 0)
	skip = settingsButton("Später", func() {
		if busy {
			return
		}
		step = 3
		render()
	})
	nav.PackStart(skip, false, false, 0)
	next = settingsButton("Weiter", func() {
		if busy {
			return
		}
		status.SetText("")
		if step == 2 {
			pw, _ := password.GetText()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			keyring := exec.CommandContext(ctx, "gnome-keyring-daemon", "--unlock", "--components=secrets")
			keyring.Stdin = strings.NewReader(pw)
			if !preview {
				if err := keyring.Run(); err != nil {
					status.SetText("Der Schlüsselbund konnte nicht geöffnet werden. Bitte gehe zurück und prüfe dein Passwort.")
					return
				}
			}
			showSetupSpotify(win, func(played bool) {
				spotifyPlayed = played
				step = 3
				render()
			})
			return
		}
		if step == 0 {
			n, _ := name.GetText()
			if strings.TrimSpace(n) == "" {
				status.SetText("Bitte deinen Namen eingeben.")
				return
			}
		}
		if step == 1 {
			pw, _ := password.GetText()
			again, _ := repeat.GetText()
			if len([]rune(pw)) < 8 || pw != again || strings.ContainsAny(pw, "\r\n\x00") {
				status.SetText("Mindestens 8 Zeichen; beide Passwörter müssen übereinstimmen.")
				return
			}
		}
		if step < 3 {
			step++
			render()
			return
		}
		if preview {
			win.Destroy()
			gtk.MainQuit()
			return
		}
		busy = true
		render()
		status.SetText("Dein Konto wird eingerichtet …")
		n, _ := name.GetText()
		pw, _ := password.GetText()
		old, _ := current.GetText()
		go func() {
			_, err := systemHelper("setup", map[string]string{"name": n, "password": pw}, old)
			glib.IdleAdd(func() {
				if !alive {
					return
				}
				busy = false
				if err != nil {
					status.SetText("Einrichtung fehlgeschlagen: " + err.Error())
					render()
					next.SetLabel("Erneut versuchen")
					return
				}
				password.SetText("")
				repeat.SetText("")
				current.SetText("")
				status.SetText("Eingerichtet. Du wirst jetzt zur Anmeldung weitergeleitet.")
				next.SetSensitive(false)
				_ = exec.Command("swaymsg", "exit").Start()
			})
		}()
	})
	nav.PackEnd(next, false, false, 0)
	win.Add(root)
	win.ShowAll()
	render()
	if step == 0 { name.GrabFocus() }
}

func (ui *settingsUI) resetSystemDialog() {
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("MiracleOS zurücksetzen")
	win.SetTransientFor(ui.win)
	win.SetModal(true)
	win.SetDefaultSize(520, 340)
	root, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 14)
	menuClass(root, "setup-card")
	title := menuLabel("MiracleOS vollständig zurücksetzen?", "dialog-heading")
	title.SetLineWrap(true)
	title.SetMaxWidthChars(35)
	root.PackStart(title, false, false, 0)
	info := settingsText("Alle Dateien, Konten, Programme und Einstellungen auf dem Systemlaufwerk werden gelöscht. Lege vorher die MiracleOS-ISO ein. Danach wird MiracleOS neu installiert und startet mit der Ersteinrichtung.")
	root.PackStart(info, false, false, 0)
	entry, _ := gtk.EntryNew()
	entry.SetVisibility(false)
	entry.SetInputPurpose(gtk.INPUT_PURPOSE_PASSWORD)
	entry.SetPlaceholderText("Dein Passwort")
	entry.SetNoShowAll(true)
	root.PackStart(entry, false, false, 0)
	status := settingsText("")
	root.PackStart(status, false, false, 0)
	stage, busy := 0, false
	secret := ""
	alive := true
	win.Connect("destroy", func() { alive = false; secret = "" })
	var confirm *gtk.Button
	confirm = settingsButton("Ja, weiter", func() {
		if busy {
			return
		}
		if stage == 0 {
			stage = 1
			title.SetText("Bestätige mit deinem Passwort")
			entry.Show()
			entry.GrabFocus()
			confirm.SetLabel("Passwort prüfen")
			return
		}
		if stage == 1 {
			pw, _ := entry.GetText()
			if pw == "" || strings.ContainsAny(pw, "\r\n\x00") {
				status.SetText("Bitte dein Passwort eingeben.")
				return
			}
			busy = true
			confirm.SetSensitive(false)
			entry.SetText("")
			status.SetText("Passwort und Installationsmedium werden geprüft …")
			go func() {
				_, err := systemHelper("check-reset", nil, pw)
				glib.IdleAdd(func() {
					if !alive {
						return
					}
					busy = false
					confirm.SetSensitive(true)
					if err != nil {
						status.SetText("Zurücksetzen nicht möglich: " + err.Error())
						return
					}
					secret = pw
					stage = 2
					entry.Hide()
					title.SetText("Letzte Bestätigung: alles löschen?")
					status.SetText("Diese Aktion löscht das gesamte Systemlaufwerk. Die Neuinstallation beginnt jetzt.")
					confirm.SetLabel("Alles löschen und neu installieren")
				})
			}()
			return
		}
		busy = true
		confirm.SetSensitive(false)
		status.SetText("Installationsmedium wird gestartet …")
		pw := secret
		secret = ""
		go func() {
			_, err := systemHelper("reset", nil, pw)
			glib.IdleAdd(func() {
				if !alive {
					return
				}
				if err != nil {
					busy = false
					stage = 1
					entry.Show()
					confirm.SetLabel("Passwort prüfen")
					confirm.SetSensitive(true)
					status.SetText("Zurücksetzen fehlgeschlagen: " + err.Error())
				}
			})
		}()
	})
	root.PackStart(confirm, false, false, 0)
	root.PackStart(settingsButton("Abbrechen", func() {
		if !busy {
			win.Destroy()
		}
	}), false, false, 0)
	win.Add(root)
	dialogEscape(win)
	win.ShowAll()
}
