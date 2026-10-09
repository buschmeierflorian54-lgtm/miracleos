package main

import (
	"encoding/json"
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	"os"
	"path/filepath"
)

var tagPalette = []string{"#ef5865", "#329df0", "#f4bf30", "#36ba77", "#ba56e5", "#f28b35", "#e958ac", "#43b8b3"}

func tagColorFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "miracleos", "tag-colors.json")
}
func readTagColors() map[string]string {
	colors := map[string]string{}
	data, _ := os.ReadFile(tagColorFile())
	_ = json.Unmarshal(data, &colors)
	return colors
}
func (fm *glassFileManager) isTag(path string) bool {
	return filepath.Dir(path) == filepath.Join(fm.home, "Documents", "Tags")
}
func (fm *glassFileManager) styleTagDot(dot *gtk.Label, path string) {
	if fm.tagDots == nil {
		fm.tagDots = map[*gtk.Label]string{}
	}
	if _, ok := fm.tagDots[dot]; !ok {
		dot.Connect("destroy", func() { delete(fm.tagDots, dot) })
	}
	fm.tagDots[dot] = path
	color := readTagColors()[path]
	if color == "" {
		for i, name := range []string{"Work", "Games", "Files", "Important", "Movies"} {
			if filepath.Base(path) == name {
				color = tagPalette[i]
			}
		}
	}
	valid := false
	for _, value := range tagPalette {
		if color == value {
			valid = true
		}
	}
	if !valid {
		color = tagPalette[0]
	}
	provider, _ := gtk.CssProviderNew()
	_ = provider.LoadFromData(fmt.Sprintf("label { color: %s; font-size: 16px; }", color))
	style, _ := dot.GetStyleContext()
	style.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER+1)
}
func (fm *glassFileManager) chooseTagColor(path string) {
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Tag-Farbe")
	win.SetTransientFor(fm.window)
	win.SetModal(true)
	root, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	menuClass(root, "fm-shell")
	menuClass(root, "glass-dialog")
	root.PackStart(glassDialogHeader(win, "Tag-Farbe"), false, false, 0)
	root.PackStart(menuLabel("Farbe für "+filepath.Base(path), "fm-card-title"), false, false, 10)
	row, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 8)
	row.SetMarginStart(20)
	row.SetMarginEnd(20)
	row.SetMarginBottom(20)
	for _, color := range tagPalette {
		color := color
		b, _ := gtk.ButtonNewWithLabel("●")
		b.SetTooltipText(color)
		provider, _ := gtk.CssProviderNew()
		_ = provider.LoadFromData("button, label {color: " + color + ";font-size:26px;}")
		style, _ := b.GetStyleContext()
		style.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER+1)
		child, _ := b.GetChild()
		cs, _ := child.ToWidget().GetStyleContext()
		cs.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER+1)
		b.Connect("clicked", func() {
			colors := readTagColors()
			colors[path] = color
			data, _ := json.Marshal(colors)
			_ = os.MkdirAll(filepath.Dir(tagColorFile()), 0755)
			if err := os.WriteFile(tagColorFile(), data, 0600); err != nil {
				fm.status.SetText(err.Error())
				return
			}
			for dot, p := range fm.tagDots {
				if p == path {
					fm.styleTagDot(dot, path)
				}
			}
			win.Destroy()
		})
		row.PackStart(b, false, false, 0)
	}
	root.PackStart(row, false, false, 0)
	win.Add(root)
	win.ShowAll()
}
