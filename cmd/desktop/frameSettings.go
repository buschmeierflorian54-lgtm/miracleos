package main

import (
	"fmt"
	"github.com/gotk3/gotk3/gtk"
)

func (ui *settingsUI) frameColors(content *gtk.Box) {
	card := settingsCard("Welche Farbe soll der Fensterrand haben?")
	card.PackStart(settingsText("Gilt für die Fensterrahmen aller Apps. Dateien und Einstellungen passen ihre Außenfläche an."), false, false, 0)
	row, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	colors := []struct{ hex, name string }{
		{"#f4ddeb", "Rosa"}, {"#c8e2e1", "Türkis"}, {"#dedff5", "Lavendel"}, {"#e8e8ec", "Silber"}, {"#000000", "Schwarz"},
	}
	buttons := []*gtk.Button{}
	for _, color := range colors {
		color := color
		b, _ := gtk.ButtonNewWithLabel(" ")
		menuClass(b, "settings-swatch")
		b.SetVAlign(gtk.ALIGN_CENTER)
		b.SetHAlign(gtk.ALIGN_CENTER)
		b.SetSizeRequest(38, 38)
		b.SetTooltipText(frameColorName(color.name, readSystemPreferences().Dark))
		swatch := frameSwatch(color.hex, readSystemPreferences().Dark)
		provider, _ := gtk.CssProviderNew()
		_ = provider.LoadFromData(fmt.Sprintf("button {background-image:none; background-color:%s; color:%s; border-radius:50%%;} button:hover {background-color:%s;}", swatch, map[bool]string{true: "#ffffff", false: "#352737"}[readSystemPreferences().Dark || color.hex == "#000000"], swatch))
		ctx, _ := b.GetStyleContext()
		ctx.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER)
		if readSystemPreferences().FrameColor == color.hex {
			b.SetLabel("✓")
		}
		b.Connect("clicked", func() {
			p := readSystemPreferences()
			p.FrameColor = color.hex
			ui.save(p)
			for _, other := range buttons {
				other.SetLabel(" ")
			}
			b.SetLabel("✓")
		})
		buttons = append(buttons, b)
		row.PackStart(b, false, false, 0)
	}
	row.PackStart(settingsButton("Systemfarbe verwenden", func() {
		p := readSystemPreferences()
		p.FrameColor = ""
		ui.save(p)
		for _, b := range buttons {
			b.SetLabel(" ")
		}
	}), false, false, 0)
	card.PackStart(row, false, false, 5)
	content.PackStart(card, false, false, 0)
}
