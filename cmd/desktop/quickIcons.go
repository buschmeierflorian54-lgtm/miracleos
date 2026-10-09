package main

import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

var quickIconPaths = map[string]string{
	"wifi":      `<path d="M3 9.5a13 13 0 0 1 18 0M6.5 13a8 8 0 0 1 11 0M9.8 16.3a3.3 3.3 0 0 1 4.4 0"/><circle cx="12" cy="19" r=".6"/>`,
	"bluetooth": `<path d="M7 8l10 8-5 4V4l5 4-10 8"/>`,
	"bell":      `<path d="M6 16V11a6 6 0 0 1 12 0v5l1.5 2h-15zM10 20.5a2.2 2.2 0 0 0 4 0"/>`,
	"night":     `<path d="M20 14.5A8 8 0 0 1 9.5 4a8 8 0 1 0 10.5 10.5z"/>`,
	"dark":      `<circle cx="12" cy="12" r="8"/><path d="M12 4v16a8 8 0 0 0 0-16z" fill="currentColor"/>`,
	"airplane":  `<path d="M12 3v7l8 4v3l-8-2v4l3 2H9l3-2v-4l-8 2v-3l8-4z"/>`,
	"sound":     `<path d="M4 10v4h3.5l4.5 4V6l-4.5 4zM15.5 9.5a4 4 0 0 1 0 5M18 7a7.5 7.5 0 0 1 0 10"/>`,
	"mute":      `<path d="M4 10v4h3.5l4.5 4V6l-4.5 4zM15.5 9.5l5 5m0-5-5 5"/>`,
	"sun":       `<circle cx="12" cy="12" r="3"/><path d="M12 3v3m0 12v3M3 12h3m12 0h3M6 6l2 2m8 8 2 2M6 18l2-2m8-8 2-2"/>`,
	"settings":  `<path d="M4 7h9m4 0h3M4 17h3m4 0h9"/><circle cx="15" cy="7" r="2"/><circle cx="9" cy="17" r="2"/>`,
	"lock":      `<rect x="5" y="10.5" width="14" height="10" rx="3"/><path d="M8.5 10.5V8a3.5 3.5 0 0 1 7 0v2.5"/>`,
	"power":     `<path d="M12 4v8m-4.5-5a7 7 0 1 0 9 0"/>`,
	"music":     `<path d="M9 18V6l10-2v12M9 8l10-2"/><circle cx="6" cy="18" r="3"/><circle cx="16" cy="16" r="3"/>`,
	"play":      `<path d="M8 5v14l11-7z" fill="currentColor"/>`, "pause": `<path d="M9 6v12M15 6v12"/>`,
	"next": `<path d="M6 7l9 5-9 5z" fill="currentColor"/><path d="M18 7v10"/>`, "previous": `<path d="M18 7l-9 5 9 5z" fill="currentColor"/><path d="M6 7v10"/>`,
	"wired": `<rect x="4" y="3" width="16" height="11" rx="2"/><path d="M12 14v4m-7 0h14M5 18v3m7-3v3m7-3v3"/>`,
}

func quickIcon(name, color string, size int) *gtk.Image {
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 24 24" fill="none" stroke="%s" color="%s" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">%s</svg>`, size, size, color, color, quickIconPaths[name])
	p, e := gdk.PixbufNewFromBytesOnly([]byte(svg))
	if e != nil {
		im, _ := gtk.ImageNewFromIconName("image-missing", gtk.ICON_SIZE_BUTTON)
		return im
	}
	im, _ := gtk.ImageNewFromPixbuf(p)
	return im
}
func quickTextColor() string {
	if readSystemPreferences().Dark {
		return "#fff0fb"
	}
	return "#3a2233"
}
