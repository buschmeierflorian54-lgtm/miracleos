package main

import (
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type appFramePreference struct {
	Color string `json:"color"`
	AppID string `json:"app_id,omitempty"`
	Class string `json:"class,omitempty"`
	Title string `json:"title_prefix,omitempty"`
}

var framePalette = []struct{ label, color string }{
	{"Globale Farbe", ""}, {"Rosa", "#f4ddeb"}, {"Türkis", "#c8e2e1"}, {"Lavendel", "#dedff5"}, {"Silber", "#e8e8ec"}, {"Schwarz", "#000000"},
}

func frameSwatch(color string, dark bool) string {
	if !dark {
		return color
	}
	switch color {
	case "#f4ddeb":
		return "#513044"
	case "#c8e2e1":
		return "#224b4a"
	case "#dedff5":
		return "#373756"
	case "#e8e8ec":
		return "#35383f"
	}
	return color
}
func frameColorName(label string, dark bool) string {
	if !dark {
		return label
	}
	switch label {
	case "Rosa":
		return "Pflaume"
	case "Türkis":
		return "Petrol"
	case "Lavendel":
		return "Nachtviolett"
	case "Silber":
		return "Graphit"
	}
	return label
}
func appFrameRule(app installedApp, color string) appFramePreference {
	r := appFramePreference{Color: color, AppID: strings.TrimSuffix(app.ID, ".desktop"), Class: app.WMClass}
	switch app.Name {
	case "Dateien":
		r.Title = "Dateien — MiracleOS"
	case "Einstellungen":
		r.Title = "Einstellungen — MiracleOS"
	case "Terminal":
		if strings.Contains(strings.ToLower(app.ID), "miracleos") {
			r.Title = "Terminal — MiracleOS"
		}
	}
	if r.Title != "" {
		r.AppID = ""
		r.Class = ""
	}
	return r
}
func appFramePicker(app installedApp, ui *settingsUI) *gtk.Box {
	row, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	row.SetVAlign(gtk.ALIGN_CENTER)
	selected := readSystemPreferences().AppFrames[app.ID].Color
	buttons := []*gtk.Button{}
	alive := true
	row.Connect("destroy", func() { alive = false })
	for _, choice := range framePalette {
		choice := choice
		b, _ := gtk.ButtonNewWithLabel(" ")
		b.SetVAlign(gtk.ALIGN_CENTER)
		b.SetHAlign(gtk.ALIGN_CENTER)
		b.SetSizeRequest(28, 28)
		b.SetTooltipText(app.Name + ": " + frameColorName(choice.label, readSystemPreferences().Dark))
		label := " "
		if choice.color == selected {
			label = "✓"
		}
		if choice.color == "" {
			label = "↺"
		}
		b.SetLabel(label)
		color := frameSwatch(choice.color, readSystemPreferences().Dark)
		if color == "" {
			color = "#f3edf2"
		}
		text := "#352737"
		if choice.color != "" && (readSystemPreferences().Dark || color == "#000000") {
			text = "#ffffff"
		}
		provider, _ := gtk.CssProviderNew()
		_ = provider.LoadFromData(fmt.Sprintf("button {background-image:none;background-color:%s;color:%s;border-radius:50%%;min-width:26px;min-height:26px;padding:0;border:1px solid alpha(%s,.3);} button label {color:%s;}", color, text, text, text))
		ctx, _ := b.GetStyleContext()
		ctx.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER)
		b.Connect("clicked", func() {
			if !alive || !ui.alive {
				return
			}
			p := readSystemPreferences()
			if p.AppFrames == nil {
				p.AppFrames = map[string]appFramePreference{}
			}
			if choice.color == "" {
				delete(p.AppFrames, app.ID)
			} else {
				p.AppFrames[app.ID] = appFrameRule(app, choice.color)
			}
			ui.save(p)
			for i, other := range buttons {
				label := " "
				if i == 0 {
					label = "↺"
				}
				other.SetLabel(label)
			}
			if choice.color != "" {
				b.SetLabel("✓")
			}
		})
		buttons = append(buttons, b)
		row.PackStart(b, false, false, 0)
	}
	return row
}
func writeAppFrames(p systemPreferences) error {
	var out strings.Builder
	keys := make([]string, 0, len(p.AppFrames))
	for key := range p.AppFrames {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		r := p.AppFrames[key]
		valid := false
		for _, c := range framePalette[1:] {
			if r.Color == c.color {
				valid = true
			}
		}
		if !valid {
			continue
		}
		local := p
		local.FrameColor = r.Color
		color := windowFrameSurface(local)
		text := "#352737"
		if p.Dark || r.Color == "#000000" {
			text = "#fff0fb"
		}
		for _, m := range []struct{ kind, value string }{{"title", r.Title}, {"app", r.AppID}, {"class", r.Class}, {"class", r.AppID}} {
			if m.value == "" || strings.ContainsAny(m.value, "\t\r\n") {
				continue
			}
			fmt.Fprintf(&out, "%s\t%s\t%s\t%s\n", m.kind, m.value, color, text)
		}
	}
	path := filepath.Join(filepath.Dir(preferencesPath()), "app-frames.tsv")
	if b, e := os.ReadFile(path); e == nil && string(b) == out.String() {
		return nil
	}
	if e := os.WriteFile(path+".tmp", []byte(out.String()), 0600); e != nil {
		return e
	}
	return os.Rename(path+".tmp", path)
}
func appClientPreferences(p systemPreferences) systemPreferences {
	title := ""
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--files":
			title = "Dateien — MiracleOS"
		case "--settings":
			title = "Einstellungen — MiracleOS"
		case "--terminal":
			title = "Terminal — MiracleOS"
		}
	}
	if title != "" {
		for _, r := range p.AppFrames {
			if r.Title == title {
				p.FrameColor = r.Color
				break
			}
		}
	}
	return p
}
