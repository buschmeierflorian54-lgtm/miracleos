package main

import (
	"github.com/gotk3/gotk3/gtk"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type pickerRule struct {
	Kind  uint32
	Value string
}
type pickerFilter struct {
	Name  string
	Rules []pickerRule
}
type filePickerOptions struct {
	Title, Folder, Name, AcceptLabel string
	Save, Directory, Multiple        bool
	Filters                          []pickerFilter
	Active                           int
	Parent                           *gtk.Window
	Done                             func([]string)
	once                             sync.Once
}

func (p *filePickerOptions) finish(paths []string) {
	p.once.Do(func() {
		if p.Done != nil {
			p.Done(paths)
		}
	})
}
func (p *filePickerOptions) accepts(path string) bool {
	if len(p.Filters) == 0 {
		return true
	}
	if p.Active < 0 || p.Active >= len(p.Filters) {
		return false
	}
	rules := p.Filters[p.Active].Rules
	if len(rules) == 0 {
		return true
	}
	for _, rule := range rules {
		if rule.Kind == 0 {
			match, _ := filepath.Match(strings.ToLower(rule.Value), strings.ToLower(filepath.Base(path)))
			if match {
				return true
			}
		}
		if rule.Kind == 1 {
			ext := strings.ToLower(filepath.Ext(path))
			typ := strings.Split(mime.TypeByExtension(ext), ";")[0]
			if ext == ".deb" {
				typ = "application/vnd.debian.binary-package"
			}
			if typ == "" {
				if file, err := os.Open(path); err == nil {
					buf := make([]byte, 512)
					n, _ := file.Read(buf)
					file.Close()
					typ = strings.Split(http.DetectContentType(buf[:n]), ";")[0]
				}
			}
			expected := strings.ToLower(rule.Value)
			if expected == typ || expected == "*/*" || (strings.HasSuffix(expected, "/*") && strings.HasPrefix(typ, strings.TrimSuffix(expected, "*"))) {
				return true
			}
		}
	}
	return false
}
func showFilePicker(p *filePickerOptions) *glassFileManager {
	if p.Folder == "" {
		p.Folder, _ = os.UserHomeDir()
	}
	fm := newGlassFileManager(p.Folder, p)
	if p.Parent != nil {
		fm.window.SetTransientFor(p.Parent)
		fm.window.SetModal(true)
		fm.window.Present()
	}
	return fm
}
func (fm *glassFileManager) addPickerControls(root *gtk.Box) {
	p := fm.picker
	area, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 9)
	menuClass(area, "picker-controls")
	if p.Title != "" {
		area.PackStart(menuLabel(p.Title, "fm-card-title"), false, false, 0)
	}
	row, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if p.Save {
		fm.pickerName, _ = gtk.EntryNew()
		fm.pickerName.SetText(p.Name)
		fm.pickerName.SetPlaceholderText("Dateiname")
		menuClass(fm.pickerName, "fm-path")
		fm.pickerName.Connect("activate", fm.submitPicker)
		row.PackStart(fm.pickerName, true, true, 0)
	}
	if len(p.Filters) > 0 {
		filters, _ := gtk.ComboBoxTextNew()
		for _, filter := range p.Filters {
			filters.AppendText(filter.Name)
		}
		filters.SetActive(p.Active)
		filters.Connect("changed", func() { p.Active = filters.GetActive(); fm.render() })
		row.PackStart(filters, true, true, 0)
	} else if !p.Save {
		row.PackStart(menuLabel("Alle Dateien", "fm-muted"), true, true, 0)
	}
	cancel, _ := gtk.ButtonNewWithLabel("Abbrechen")
	menuClass(cancel, "installer-secondary")
	cancel.Connect("clicked", func() { fm.window.Destroy() })
	row.PackEnd(cancel, false, false, 0)
	label := p.AcceptLabel
	if label == "" {
		label = "Auswählen"
		if p.Save {
			label = "Speichern"
		}
	}
	accept, _ := gtk.ButtonNewWithLabel(strings.ReplaceAll(label, "_", ""))
	menuClass(accept, "installer-primary")
	accept.Connect("clicked", fm.submitPicker)
	row.PackEnd(accept, false, false, 0)
	area.PackStart(row, false, false, 0)
	root.PackEnd(area, false, false, 0)
}
func (fm *glassFileManager) submitPicker() {
	p := fm.picker
	paths := append([]string(nil), fm.pickerSelection...)
	if p.Save {
		name, _ := fm.pickerName.GetText()
		name = strings.TrimSpace(name)
		if name == "" || filepath.Base(name) != name || name == "." || name == ".." {
			fm.status.SetText("Bitte einen gültigen Dateinamen eingeben.")
			return
		}
		if filepath.Ext(name) == "" && len(p.Filters) > 0 && p.Active >= 0 && p.Active < len(p.Filters) {
			for _, rule := range p.Filters[p.Active].Rules {
				if rule.Kind == 0 && strings.HasPrefix(rule.Value, "*.") && !strings.ContainsAny(rule.Value[2:], "*?[]") {
					name += rule.Value[1:]
					break
				}
			}
		}
		paths = []string{filepath.Join(fm.current, name)}
		if !p.accepts(paths[0]) {
			fm.status.SetText("Dieses Dateiformat wird nicht akzeptiert.")
			return
		}
		if info, err := os.Stat(paths[0]); err == nil {
			if info.IsDir() {
				fm.status.SetText("Dieser Name gehört zu einem Ordner.")
				return
			}
			dialog := gtk.MessageDialogNew(fm.window, gtk.DIALOG_MODAL, gtk.MESSAGE_QUESTION, gtk.BUTTONS_YES_NO, "Die Datei existiert bereits. Ersetzen?")
			menuClass(dialog, "fm-dialog")
			response := dialog.Run()
			dialog.Destroy()
			if response != gtk.RESPONSE_YES {
				return
			}
		}
	} else {
		if p.Directory && len(paths) == 0 {
			paths = []string{fm.current}
		}
		if len(paths) == 0 {
			fm.status.SetText("Bitte eine Datei auswählen.")
			return
		}
		for _, path := range paths {
			info, err := os.Stat(path)
			if err != nil {
				fm.status.SetText("Die Auswahl ist nicht mehr verfügbar.")
				return
			}
			if info.IsDir() != p.Directory {
				if info.IsDir() {
					fm.navigate(path, true)
				}
				return
			}
			if !info.IsDir() && !p.accepts(path) {
				fm.status.SetText("Dieses Dateiformat wird nicht akzeptiert.")
				return
			}
		}
	}
	p.finish(paths)
	fm.window.Destroy()
}
