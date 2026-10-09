package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func debField(path, field string) string {
	data, err := exec.Command("dpkg-deb", "--field", path, field).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func showPackageInstaller(path string) {
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Paketinstaller — MiracleOS")
	win.SetDefaultSize(660, 590)
	win.Connect("destroy", gtk.MainQuit)
	busy, alive := false, true
	var status *gtk.Label
	canClose := func() bool {
		if busy {
			status.SetText("Die Installation läuft noch. Bitte auf den Abschluss warten.")
			return false
		}
		return true
	}
	win.Connect("delete-event", func() bool { return !canClose() })
	root, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	menuClass(root, "fm-shell")
	root.PackStart(glassWindowHeader(win, "Paketinstaller — MiracleOS", canClose), false, false, 0)
	content, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 16)
	menuClass(content, "installer-content")
	icon := menuIcon("store", 92)
	icon.SetHAlign(gtk.ALIGN_CENTER)
	content.PackStart(icon, false, false, 0)
	heading := menuLabel("Paket öffnen", "installer-title")
	heading.SetXAlign(0.5)
	content.PackStart(heading, false, false, 0)
	subtitle := menuLabel("Debian-Programme für deinen Desktop", "installer-subtitle")
	subtitle.SetXAlign(0.5)
	subtitle.SetLineWrap(true)
	subtitle.SetMaxWidthChars(60)
	content.PackStart(subtitle, false, false, 0)
	card, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	menuClass(card, "installer-card")
	metadata := menuLabel("Wähle eine heruntergeladene .deb-Datei.", "installer-metadata")
	card.PackStart(metadata, false, false, 0)
	description := menuLabel("", "installer-description")
	description.SetLineWrap(true)
	description.SetMaxWidthChars(66)
	card.PackStart(description, false, false, 0)
	content.PackStart(card, false, false, 0)
	status = menuLabel("Bereit", "installer-status")
	status.SetLineWrap(true)
	status.SetMaxWidthChars(65)
	content.PackStart(status, false, false, 0)
	progress, _ := gtk.ProgressBarNew()
	menuClass(progress, "installer-progress")
	content.PackStart(progress, false, false, 0)
	buttons, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	open, _ := gtk.ButtonNewWithLabel("Paket auswählen …")
	install, _ := gtk.ButtonNewWithLabel("Installieren")
	menuClass(open, "installer-secondary")
	menuClass(install, "installer-primary")
	install.SetSensitive(false)
	buttons.PackStart(open, false, false, 0)
	buttons.PackEnd(install, false, false, 0)
	content.PackEnd(buttons, false, false, 0)
	root.PackStart(content, true, true, 0)
	win.Add(root)
	selected := ""
	packageName := ""
	win.Connect("destroy", func() { alive = false })
	load := func(file string) {
		absolute, err := filepath.Abs(file)
		if err != nil {
			status.SetText("Die Datei konnte nicht geöffnet werden.")
			return
		}
		info, err := os.Stat(absolute)
		if err != nil || info.IsDir() || !strings.EqualFold(filepath.Ext(absolute), ".deb") {
			status.SetText("Bitte eine vorhandene Debian-Paketdatei (.deb) auswählen.")
			return
		}
		name := debField(absolute, "Package")
		if name == "" {
			status.SetText("Diese Datei ist kein gültiges Debian-Paket.")
			install.SetSensitive(false)
			selected = ""
			return
		}
		selected, packageName = absolute, name
		heading.SetText(name)
		version, arch := debField(absolute, "Version"), debField(absolute, "Architecture")
		metadata.SetText(fmt.Sprintf("Version %s    ·    %s    ·    %.1f MB", version, arch, float64(info.Size())/(1024*1024)))
		text := debField(absolute, "Description")
		lines := strings.Split(text, "\n")
		if len(lines) > 7 {
			lines = lines[:7]
		}
		description.SetText(strings.Join(lines, "\n"))
		subtitle.SetText(filepath.Base(absolute))
		status.SetText("Bereit zur Installation. Benötigte Pakete werden mitinstalliert.")
		progress.SetFraction(0)
		install.SetSensitive(true)
		install.SetLabel("Installieren")
	}
	open.Connect("clicked", func() {
		showFilePicker(&filePickerOptions{Title: "Debian-Paket auswählen", Parent: win, Filters: []pickerFilter{{Name: "Debian-Pakete (*.deb)", Rules: []pickerRule{{Kind: 0, Value: "*.deb"}}}}, Done: func(paths []string) {
			if len(paths) > 0 && alive {
				load(paths[0])
			}
		}})
	})
	install.Connect("clicked", func() {
		if selected == "" || busy {
			return
		}
		busy = true
		file, name := selected, packageName
		install.SetSensitive(false)
		open.SetSensitive(false)
		status.SetText("Bitte die Installation mit deinem Passwort freigeben …")
		glib.TimeoutAdd(200, func() bool {
			if !alive || !busy {
				return false
			}
			progress.Pulse()
			return true
		})
		go func() {
			command := exec.Command("pkexec", "/usr/bin/gdebi", "--non-interactive", file)
			pipe, err := command.StdoutPipe()
			if err == nil {
				command.Stderr = command.Stdout
				err = command.Start()
			}
			last := ""
			if err == nil {
				scanner := bufio.NewScanner(pipe)
				for scanner.Scan() {
					line := strings.TrimSpace(scanner.Text())
					if line == "" {
						continue
					}
					last = line
					message := line
					glib.IdleAdd(func() {
						if alive {
							status.SetText(message)
						}
					})
				}
				err = command.Wait()
			}
			result := "Installation abgeschlossen."
			if err != nil {
				result = "Installation abgebrochen oder fehlgeschlagen."
				if last != "" {
					result += "\n" + last
				}
			} else {
				installed, checkErr := exec.Command("dpkg-query", "-W", "-f=${Status}", name).Output()
				if checkErr != nil || !strings.Contains(string(installed), "install ok installed") {
					result = "Das Paket wurde noch nicht vollständig eingerichtet."
				}
			}
			glib.IdleAdd(func() {
				busy = false
				if !alive {
					return
				}
				status.SetText(result)
				progress.SetFraction(1)
				open.SetSensitive(true)
				install.SetSensitive(true)
				if err == nil {
					install.SetLabel("Erneut installieren")
				} else {
					install.SetLabel("Erneut versuchen")
					progress.SetFraction(0)
				}
			})
		}()
	})
	if path != "" {
		load(path)
	}
	win.ShowAll()
}
