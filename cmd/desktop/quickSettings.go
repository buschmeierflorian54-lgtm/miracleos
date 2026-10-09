package main

import (
	"context"
	"fmt"
	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func quickCommand(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	b, e := cmd.CombinedOutput()
	return strings.TrimSpace(string(b)), e
}

type quickHardware struct {
	wifi, bluetooth, audio bool
	brightness             string
	volume                 float64
}

func detectQuickHardware() quickHardware {
	h := quickHardware{volume: 100}
	devices, _ := quickCommand("nmcli", "-t", "-f", "TYPE", "device")
	h.wifi = strings.Contains(devices, "wifi")
	bt, e := quickCommand("bluetoothctl", "list")
	h.bluetooth = e == nil && strings.Contains(bt, "Controller ")
	v, e := quickCommand("pactl", "get-sink-volume", "@DEFAULT_SINK@")
	h.audio = e == nil
	for _, s := range strings.Fields(v) {
		if strings.HasSuffix(s, "%") {
			h.volume, _ = strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
			break
		}
	}
	paths, _ := filepath.Glob("/sys/class/backlight/*/brightness")
	for _, p := range paths {
		f, e := os.OpenFile(p, os.O_WRONLY, 0)
		if e == nil {
			f.Close()
			h.brightness = p
			break
		}
	}
	return h
}

var activeQuickSettings *gtk.Window
var quickTray *gtk.Button

type quickTile struct {
	button   *gtk.Button
	subtitle *gtk.Label
	icon     *gtk.Box
	name     string
	on       bool
}

func (t *quickTile) update(enabled, on bool) {
	t.on = on
	t.button.SetSensitive(enabled)
	if enabled {
		t.button.SetTooltipText("")
	}
	sc, _ := t.button.GetStyleContext()
	sc.RemoveClass("quick-active")
	color := quickTextColor()
	text := "Aus"
	if !enabled {
		text = "Nicht verfügbar"
		t.button.SetTooltipText("Auf diesem Rechner nicht verfügbar")
	} else if on {
		text = "An"
		color = "white"
		sc.AddClass("quick-active")
	}
	t.subtitle.SetText(text)
	children := t.icon.GetChildren()
	children.Foreach(func(v interface{}) { v.(*gtk.Widget).Destroy() })
	t.icon.PackStart(quickIcon(t.name, color, 20), false, false, 0)
	t.icon.ShowAll()
}
func showQuickSettings() {
	if activeQuickSettings != nil {
		activeQuickSettings.Destroy()
		return
	}
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Schnelleinstellungen")
	win.SetDecorated(false)
	win.SetResizable(false)
	win.SetAppPaintable(true)
	menuClass(win, "quick-window")
	if screen := win.GetScreen(); screen != nil {
		if visual, e := screen.GetRGBAVisual(); e == nil {
			win.SetVisual(visual)
		}
	}
	layershell.InitForWindow(win)
	layershell.SetNamespace(win, "miracleos-quick-settings")
	layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_OVERLAY)
	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_RIGHT, true)
	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_BOTTOM, true)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_RIGHT, 16)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_BOTTOM, 70)
	layershell.SetKeyboardMode(win, layershell.LAYER_SHELL_KEYBOARD_MODE_EXCLUSIVE)
	activeQuickSettings = win
	dismissPopupOutside(win, func() { win.Destroy() })
	if quickTray != nil {
		menuClass(quickTray, "quick-tray-active")
	}
	root, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 16)
	menuClass(root, "quick-panel")
	root.SetSizeRequest(380, -1)
	win.Add(root)
	alive := true
	stop := make(chan struct{})
	win.Connect("destroy", func() {
		alive = false
		close(stop)
		activeQuickSettings = nil
		if quickTray != nil {
			sc, _ := quickTray.GetStyleContext()
			sc.RemoveClass("quick-tray-active")
		}
	})
	win.Connect("key-press-event", func(_ *gtk.Window, e *gdk.Event) bool {
		if gdk.EventKeyNewFromEvent(e).KeyVal() == gdk.KEY_Escape {
			win.Destroy()
			return true
		}
		return false
	})
	header, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	header.SetSizeRequest(-1, 36)
	title, _ := gtk.LabelNew("Schnelleinstellungen")
	title.SetXAlign(0)
	menuClass(title, "quick-title")
	header.PackStart(title, true, true, 0)
	iconButton := func(name, hint string, run func()) *gtk.Button {
		b, _ := gtk.ButtonNew()
		menuClass(b, "quick-icon-button")
		b.Add(quickIcon(name, quickTextColor(), 19))
		b.SetTooltipText(hint)
		b.Connect("clicked", run)
		return b
	}
	header.PackEnd(iconButton("settings", "Einstellungen öffnen", func() { win.Destroy(); launchSettings() }), false, false, 0)
	root.PackStart(header, false, false, 0)
	status, _ := gtk.LabelNew("")
	status.SetLineWrap(true)
	status.SetNoShowAll(true)
	menuClass(status, "quick-error")
	showError := func(e error) {
		if e != nil {
			status.SetText("Änderung fehlgeschlagen. Bitte erneut versuchen.")
			status.Show()
		}
	}
	// The media card exists only while Spotify exposes its real MPRIS player.
	mediaBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	menuClass(mediaBox, "quick-media")
	mediaBox.SetNoShowAll(true)
	root.PackStart(mediaBox, false, false, 0)
	mediaTop, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 12)
	art, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	menuClass(art, "quick-media-art")
	art.PackStart(quickIcon("music", "#f5d8e9", 22), true, true, 0)
	updateArtwork := mediaArtworkUpdater(art,52)
	mediaTop.PackStart(art, false, false, 0)
	trackBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 2)
	track, _ := gtk.LabelNew("")
	track.SetXAlign(0)
	track.SetMaxWidthChars(23)
	track.SetEllipsize(3)
	menuClass(track, "quick-track")
	artist, _ := gtk.LabelNew("")
	artist.SetXAlign(0)
	artist.SetMaxWidthChars(23)
	artist.SetEllipsize(3)
	trackBox.PackStart(track, false, false, 0)
	trackBox.PackStart(artist, false, false, 0)
	mediaTop.PackStart(trackBox, true, true, 0)
	badge, _ := gtk.LabelNew("Spotify")
	menuClass(badge, "quick-media-badge")
	badge.SetVAlign(gtk.ALIGN_START)
	mediaTop.PackEnd(badge, false, false, 0)
	mediaBox.PackStart(mediaTop, false, false, 0)
	progress, _ := gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL, 0, 100, 1)
	progress.SetDrawValue(false)
	menuClass(progress, "quick-media-progress")
	mediaBox.PackStart(progress, false, false, 0)
	controls, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 20)
	controls.SetHAlign(gtk.ALIGN_CENTER)
	var media mediaSnapshot
	mediaCommand := func(method string, args ...interface{}) {
		snap := media
		go func() {
			e := spotifyMediaCommand(snap, method, args...)
			glib.IdleAdd(func() {
				if alive {
					showError(e)
				}
			})
		}()
	}
	previous := iconButton("previous", "Vorheriger Titel", func() { mediaCommand("Previous") })
	play := iconButton("play", "Wiedergabe / Pause", func() { mediaCommand("PlayPause") })
	next := iconButton("next", "Nächster Titel", func() { mediaCommand("Next") })
	for _, b := range []*gtk.Button{previous, next} {
		c, _ := b.GetChild()
		c.ToWidget().Destroy()
	}
	previous.Add(quickIcon("previous", "#ffdaef", 19))
	next.Add(quickIcon("next", "#ffdaef", 19))
	menuClass(play, "quick-media-play")
	controls.PackStart(previous, false, false, 0)
	controls.PackStart(play, false, false, 0)
	controls.PackStart(next, false, false, 0)
	mediaBox.PackStart(controls, false, false, 0)
	progress.Connect("button-release-event", func(_ interface{}, _ *gdk.Event) bool {
		if media.Seek && media.Length > 0 && media.Track != "" {
			mediaCommand("SetPosition", media.Track, int64(progress.GetValue()*float64(media.Length)/100))
		}
		return false
	})
	grid, _ := gtk.GridNew()
	grid.SetRowSpacing(8)
	grid.SetColumnSpacing(8)
	grid.SetColumnHomogeneous(true)
	root.PackStart(grid, false, false, 0)
	pending := false
	run := func(tile *quickTile, job func() error) {
		if pending {
			return
		}
		pending = true
		tile.button.SetSensitive(false)
		go func() {
			e := job()
			glib.IdleAdd(func() {
				if alive {
					pending = false
					tile.button.SetSensitive(true)
					if e == nil && tile.name == "dark" {
						win.Destroy()
						showQuickSettings()
						return
					}
					showError(e)
				}
			})
		}()
	}
	command := func(name string, args ...string) error { _, e := quickCommand(name, args...); return e }
	tiles := make([]*quickTile, 6)
	names := []string{"WLAN", "Bluetooth", "Nicht stören", "Nachtlicht", "Dunkelmodus", "Flugmodus"}
	icons := []string{"wifi", "bluetooth", "bell", "night", "dark", "airplane"}
	for i := range names {
		t := &quickTile{name: icons[i]}
		t.button, _ = gtk.ButtonNew()
		menuClass(t.button, "quick-tile")
		row, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
		t.icon, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
		t.icon.SetVAlign(gtk.ALIGN_CENTER)
		row.PackStart(t.icon, false, false, 0)
		labels, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)
		labels.SetVAlign(gtk.ALIGN_CENTER)
		l, _ := gtk.LabelNew(names[i])
		l.SetXAlign(0)
		menuClass(l, "quick-tile-title")
		t.subtitle, _ = gtk.LabelNew("Prüfen …")
		t.subtitle.SetXAlign(0)
		menuClass(t.subtitle, "quick-tile-state")
		labels.PackStart(l, false, false, 0)
		labels.PackStart(t.subtitle, false, false, 0)
		row.PackStart(labels, true, true, 0)
		t.button.Add(row)
		t.update(false, false)
		tiles[i] = t
		grid.Attach(t.button, i%2, i/2, 1, 1)
		index := i
		t.button.Connect("clicked", func() {
			on := t.on
			run(t, func() error {
				switch index {
				case 0:
					v := "on"
					if on {
						v = "off"
					}
					return command("nmcli", "radio", "wifi", v)
				case 1:
					v := "on"
					if on {
						v = "off"
					}
					return command("bluetoothctl", "power", v)
				case 4:
					p := readSystemPreferences()
					p.Dark = !p.Dark
					p = modeWallpaper(p)
					if e := saveSystemPreferences(p); e != nil {
						return e
					}
					return applyWallpaper(p.Wallpaper)
				case 5:
					v := "off"
					if on {
						v = "on"
					}
					return command("nmcli", "radio", "all", v)
				}
				return nil
			})
		})
	}
	sliders, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	menuClass(sliders, "quick-sliders")
	root.PackStart(sliders, false, false, 0)
	type sliderUI struct {
		scale   *gtk.Scale
		value   *gtk.Label
		setting bool
		timer   glib.SourceHandle
	}
	var hardware quickHardware
	slider := func(label, icon string, change func(float64) error) *sliderUI {
		u := &sliderUI{}
		box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
		head, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 8)
		l, _ := gtk.LabelNew(label)
		l.SetXAlign(0)
		menuClass(l, "quick-slider-title")
		head.PackStart(l, true, true, 0)
		u.value, _ = gtk.LabelNew("—")
		menuClass(u.value, "quick-slider-value")
		head.PackEnd(u.value, false, false, 0)
		box.PackStart(head, false, false, 0)
		row, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
		row.PackStart(quickIcon(icon, quickTextColor(), 18), false, false, 0)
		u.scale, _ = gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL, 0, 100, 1)
		u.scale.SetDrawValue(false)
		u.scale.SetSensitive(false)
		row.PackStart(u.scale, true, true, 0)
		box.PackStart(row, false, false, 0)
		sliders.PackStart(box, false, false, 0)
		u.scale.Connect("value-changed", func() {
			if u.setting {
				return
			}
			v := u.scale.GetValue()
			u.value.SetText(fmt.Sprintf("%.0f %%", v))
			if u.timer != 0 {
				glib.SourceRemove(u.timer)
			}
			u.timer = glib.TimeoutAdd(160, func() bool {
				u.timer = 0
				go func() {
					e := change(v)
					glib.IdleAdd(func() {
						if alive {
							showError(e)
						}
					})
				}()
				return false
			})
		})
		win.Connect("destroy", func() {
			if u.timer != 0 {
				glib.SourceRemove(u.timer)
			}
		})
		return u
	}
	volume := slider("Lautstärke", "sound", func(v float64) error {
		return command("pactl", "set-sink-volume", "@DEFAULT_SINK@", fmt.Sprintf("%.0f%%", v))
	})
	brightness := slider("Helligkeit", "sun", func(v float64) error {
		if hardware.brightness == "" {
			return nil
		}
		b, e := os.ReadFile(filepath.Join(filepath.Dir(hardware.brightness), "max_brightness"))
		if e != nil {
			return e
		}
		max, e := strconv.Atoi(strings.TrimSpace(string(b)))
		if e != nil {
			return e
		}
		n := int(v * float64(max) / 100)
		if n < 1 {
			n = 1
		}
		return os.WriteFile(hardware.brightness, []byte(strconv.Itoa(n)), 0644)
	})
	footer, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 8)
	footer.SetSizeRequest(-1, 44)
	avatar, _ := gtk.LabelNew("t")
	menuClass(avatar, "quick-avatar")
	avatar.SetVAlign(gtk.ALIGN_CENTER)
	footer.PackStart(avatar, false, false, 0)
	username := os.Getenv("USER")
	if u, e := user.Current(); e == nil {
		username = u.Username
	}
	account, _ := gtk.LabelNew(username)
	account.SetXAlign(0)
	footer.PackStart(account, true, true, 0)
	footer.PackEnd(iconButton("power", "Ausschalten oder neu starten", func() {
		menu, _ := gtk.MenuNew()
		menuClass(menu, "glass-context-menu")
		for _, entry := range []struct{ name, action string }{{"Herunterfahren", "poweroff"}, {"Neustarten", "reboot"}} {
			entry := entry
			item, _ := gtk.MenuItemNewWithLabel(entry.name)
			item.Connect("activate", func() { win.Destroy(); sessionAction(entry.action) })
			menu.Append(item)
		}
		menu.ShowAll()
		menu.PopupAtPointer(nil)
		menu.Connect("selection-done", menu.Destroy)
	}), false, false, 0)
	footer.PackEnd(iconButton("lock", "Sperren", func() { win.Destroy(); sessionAction("lock") }), false, false, 0)
	root.PackStart(footer, false, false, 0)
	root.PackStart(status, false, false, 0)
	updateSlider := func(u *sliderUI, enabled bool, v float64) {
		u.scale.SetSensitive(enabled)
		if !enabled {
			u.value.SetText("—")
			u.scale.SetTooltipText("Auf diesem Rechner nicht verfügbar")
			return
		}
		if u.timer != 0 {
			return
		}
		u.setting = true
		u.scale.SetValue(v)
		u.setting = false
		u.value.SetText(fmt.Sprintf("%.0f %%", v))
	}
	win.ShowAll()
	mediaBox.Hide()
	go func() {
		for {
			h := detectQuickHardware()
			wifi, _ := quickCommand("nmcli", "radio", "wifi")
			bt, _ := quickCommand("bluetoothctl", "show")
			snap := readSpotifyMedia()
			p := readSystemPreferences()
			bri := 50.0
			if h.brightness != "" {
				m, _ := os.ReadFile(filepath.Join(filepath.Dir(h.brightness), "max_brightness"))
				max, _ := strconv.Atoi(strings.TrimSpace(string(m)))
				b, _ := os.ReadFile(h.brightness)
				n, _ := strconv.Atoi(strings.TrimSpace(string(b)))
				if max > 0 {
					bri = float64(n) * 100 / float64(max)
				}
			}
			glib.IdleAdd(func() {
				if !alive {
					return
				}
				hardware = h
				if !pending {
					tiles[0].update(h.wifi, wifi == "enabled")
					tiles[1].update(h.bluetooth, strings.Contains(bt, "Powered: yes"))
					tiles[2].update(false, false)
					tiles[3].update(false, false)
					tiles[4].update(true, p.Dark)
					tiles[5].update(h.wifi, wifi == "disabled")
				}
				updateSlider(volume, h.audio, h.volume)
				updateSlider(brightness, h.brightness != "", bri)
				media = snap
				updateArtwork(snap.ArtURL)
				if snap.Bus == "" {
					mediaBox.Hide()
				} else {
					track.SetText(snap.Title)
					artist.SetText(snap.Artist)
					previous.SetSensitive(snap.Control && snap.Previous)
					next.SetSensitive(snap.Control && snap.Next)
					play.SetSensitive(snap.Control)
					child, _ := play.GetChild()
					child.ToWidget().Destroy()
					icon := "play"
					if snap.Playing {
						icon = "pause"
					}
					play.Add(quickIcon(icon, "#3a2233", 20))
					play.ShowAll()
					progress.SetSensitive(snap.Seek && snap.Length > 0)
					if snap.Length > 0 {
						progress.SetValue(float64(snap.Position) * 100 / float64(snap.Length))
					} else {
						progress.SetValue(0)
					}
					mediaBox.SetNoShowAll(false)
					mediaBox.ShowAll()
					mediaBox.SetNoShowAll(true)
				}
			})
			select {
			case <-stop:
				return
			case <-time.After(1500 * time.Millisecond):
			}
		}
	}()
}
