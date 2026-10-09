package main

import (
	"context"
	"fmt"
	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/godbus/dbus/v5"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
	"os/exec"
	"time"
)

func spotifySetProperty(s mediaSnapshot, name string, value interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	c, err := dbus.ConnectSessionBus()
	if err != nil {
		return err
	}
	defer c.Close()
	return c.Object(s.Bus, "/org/mpris/MediaPlayer2").CallWithContext(ctx, "org.freedesktop.DBus.Properties.Set", 0, "org.mpris.MediaPlayer2.Player", name, dbus.MakeVariant(value)).Err
}
func mediaTime(us int64) string {
	if us < 0 {
		us = 0
	}
	seconds := us / 1000000
	return fmt.Sprintf("%d:%02d", seconds/60, seconds%60)
}

func overlayButtonIcon(button *gtk.Button, icon, tooltip string) {
	button.SetLabel("")
	button.SetImage(fmIcon(icon, gtk.ICON_SIZE_BUTTON))
	button.SetAlwaysShowImage(true)
	button.SetTooltipText(tooltip)
	button.SetVAlign(gtk.ALIGN_CENTER)
	menuClass(button, "spotify-icon-button")
}

func showSpotifyOverlay() {
	// One overlay per session: another invocation of Alt+F toggles the window.
	bus, err := dbus.ConnectSessionBus()
	if err != nil {
		return
	}
	defer bus.Close()
	reply, err := bus.RequestName("org.miracleos.SpotifyOverlay", dbus.NameFlagDoNotQueue)
	if err != nil {
		return
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		_ = bus.Object("org.miracleos.SpotifyOverlay", "/org/miracleos/SpotifyOverlay").Call("org.miracleos.SpotifyOverlay.Toggle", 0).Err
		return
	}
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Spotify — MiracleOS")
	win.SetDecorated(false)
	win.SetAppPaintable(true)
	if screen := win.GetScreen(); screen != nil {
		if visual, e := screen.GetRGBAVisual(); e == nil {
			win.SetVisual(visual)
		}
	}
	layershell.InitForWindow(win)
	layershell.SetNamespace(win, "miracleos-spotify")
	layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_OVERLAY)
	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_BOTTOM, true)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_BOTTOM, 96)
	layershell.SetKeyboardMode(win, layershell.LAYER_SHELL_KEYBOARD_MODE_ON_DEMAND)
	toggle := &spotifyOverlayToggle{win: win}
	bus.Export(toggle, "/org/miracleos/SpotifyOverlay", "org.miracleos.SpotifyOverlay")
	root, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 22)
	menuClass(root, "spotify-overlay")
	root.SetSizeRequest(620, -1)
	art, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	art.PackStart(quickIcon("music", "#ffd9ec", 46), true, true, 0)
	updateArt := mediaArtworkUpdater(art, 150)
	root.PackStart(art, false, false, 0)
	body, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	body.SetHExpand(true)
	root.PackStart(body, true, true, 0)
	top, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 12)
	body.PackStart(top, false, false, 0)
	text, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 4)
	top.PackStart(text, true, true, 0)
	status := menuLabel("Spotify öffnen", "spotify-overlay-status")
	status.SetXAlign(0)
	text.PackStart(status, false, false, 0)
	title := menuLabel("Keine Wiedergabe", "spotify-overlay-title")
	title.SetXAlign(0)
	title.SetEllipsize(pango.ELLIPSIZE_END)
	title.SetMaxWidthChars(26)
	text.PackStart(title, false, false, 0)
	artist := menuLabel("Melde dich in Spotify an.", "canvas-artist")
	artist.SetXAlign(0)
	artist.SetEllipsize(pango.ELLIPSIZE_END)
	artist.SetMaxWidthChars(40)
	text.PackStart(artist, false, false, 0)
	close := settingsButton("×", func() { win.Hide() })
	close.SetTooltipText("Overlay schließen · Alt + F")
	overlayButtonIcon(close, "window-close-symbolic", "Overlay schließen · Alt + F")
	top.PackEnd(close, false, false, 0)
	timeline, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 8)
	body.PackStart(timeline, false, false, 0)
	position := menuLabel("0:00", "canvas-artist")
	length := menuLabel("0:00", "canvas-artist")
	progress, _ := gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL, 0, 100, 1)
	progress.SetDrawValue(false)
	timeline.PackStart(position, false, false, 0)
	timeline.PackStart(progress, true, true, 0)
	timeline.PackStart(length, false, false, 0)
	controls, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	body.PackStart(controls, false, false, 0)
	var media mediaSnapshot
	updating, fetching, alive := false, false, true
	invoke := func(method string) {
		snap := media
		if snap.Bus != "" {
			go func() { _ = spotifyMediaCommand(snap, method) }()
		}
	}
	shuffle := settingsButton("⤨", func() {
		snap := media
		if snap.Bus != "" {
			go func() { _ = spotifySetProperty(snap, "Shuffle", !snap.Shuffle) }()
		}
	})
	shuffle.SetTooltipText("Zufallswiedergabe")
	controls.PackStart(shuffle, false, false, 0)
	previous := settingsButton("⏮", func() { invoke("Previous") })
	controls.PackStart(previous, false, false, 0)
	play := settingsButton("▶", func() { invoke("PlayPause") })
	menuClass(play, "canvas-player-main")
	controls.PackStart(play, false, false, 0)
	next := settingsButton("⏭", func() { invoke("Next") })
	controls.PackStart(next, false, false, 0)
	repeat := settingsButton("↻", func() {
		snap := media
		value := "Playlist"
		if snap.Loop == "Playlist" {
			value = "Track"
		} else if snap.Loop == "Track" {
			value = "None"
		}
		if snap.Bus != "" {
			go func() { _ = spotifySetProperty(snap, "LoopStatus", value) }()
		}
	})
	controls.PackStart(repeat, false, false, 0)
	overlayButtonIcon(shuffle, "media-playlist-shuffle-symbolic", "Zufallswiedergabe")
	overlayButtonIcon(previous, "media-skip-backward-symbolic", "Vorheriger Titel")
	overlayButtonIcon(play, "media-playback-start-symbolic", "Abspielen")
	overlayButtonIcon(next, "media-skip-forward-symbolic", "Nächster Titel")
	overlayButtonIcon(repeat, "media-playlist-repeat-symbolic", "Wiederholen")
	volume, _ := gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL, 0, 100, 1)
	volume.SetDrawValue(false)
	volume.SetSizeRequest(100, -1)
	volume.SetTooltipText("Spotify-Lautstärke")
	controls.PackEnd(volume, true, true, 0)
	controls.PackEnd(menuLabel("♫", "canvas-artist"), false, false, 0)
	body.PackStart(settingsButton("Spotify öffnen", launchSpotify), false, false, 0)
	progress.Connect("change-value", func(_ *gtk.Scale, _ gtk.ScrollType, value float64) bool {
		if !updating && media.Seek && media.Length > 0 {
			snap := media
			go func() {
				_ = spotifyMediaCommand(snap, "SetPosition", snap.Track, int64(value*float64(snap.Length)/100))
			}()
		}
		return false
	})
	volume.Connect("change-value", func(_ *gtk.Scale, _ gtk.ScrollType, value float64) bool {
		if !updating && media.Control && media.HasVolume {
			snap := media
			go func() { _ = spotifySetProperty(snap, "Volume", value/100) }()
		}
		return false
	})
	refresh := func() bool {
		if !alive {
			return false
		}
		if fetching {
			return true
		}
		fetching = true
		go func() {
			snap := readSpotifyMedia()
			glib.IdleAdd(func() {
				fetching = false
				if !alive {
					return
				}
				media = snap
				title.SetText(snap.Title)
				details := snap.Artist
				if snap.Album != "" {
					details += " · " + snap.Album
				}
				artist.SetText(details)
				if snap.Bus == "" {
					title.SetText("Keine Wiedergabe")
					artist.SetText("Spotify öffnen und anmelden")
					status.SetText("Spotify ist bereit")
				} else if snap.Playing {
					status.SetText("Läuft gerade auf Spotify")
				} else {
					status.SetText("Pausiert")
				}
				updateArt(snap.ArtURL)
				play.SetSensitive(snap.Control)
				previous.SetSensitive(snap.Control && snap.Previous)
				next.SetSensitive(snap.Control && snap.Next)
				shuffle.SetSensitive(snap.Control && snap.HasShuffle)
				repeat.SetSensitive(snap.Control && snap.HasLoop)
				volume.SetSensitive(snap.Control && snap.HasVolume)
				progress.SetSensitive(snap.Seek && snap.Length > 0)
				if snap.Playing {
					overlayButtonIcon(play, "media-playback-pause-symbolic", "Pausieren")
				} else {
					overlayButtonIcon(play, "media-playback-start-symbolic", "Abspielen")
				}
				if snap.Shuffle {
					shuffle.SetStateFlags(gtk.STATE_FLAG_CHECKED, false)
				} else {
					shuffle.UnsetStateFlags(gtk.STATE_FLAG_CHECKED)
				}
				if snap.Loop == "Track" {
					overlayButtonIcon(repeat, "media-playlist-repeat-song-symbolic", "Diesen Titel wiederholen")
				} else if snap.Loop == "Playlist" {
					overlayButtonIcon(repeat, "media-playlist-repeat-symbolic", "Playlist wiederholen")
				} else {
					overlayButtonIcon(repeat, "media-playlist-repeat-symbolic", "Wiederholen aus")
				}
				position.SetText(mediaTime(snap.Position))
				length.SetText(mediaTime(snap.Length))
				updating = true
				v := 0.0
				if snap.Length > 0 {
					v = 100 * float64(snap.Position) / float64(snap.Length)
				}
				progress.SetValue(v)
				volume.SetValue(100 * snap.Volume)
				updating = false
			})
		}()
		return true
	}
	win.Connect("destroy", func() { alive = false; gtk.MainQuit() })
	win.Connect("key-press-event", func(_ *gtk.Window, e *gdk.Event) bool {
		if gdk.EventKeyNewFromEvent(e).KeyVal() == gdk.KEY_Escape {
			win.Hide()
			return true
		}
		return false
	})
	win.Add(root)
	win.ShowAll()
	refresh()
	glib.TimeoutAdd(1000, refresh)
	gtk.Main()
}

type spotifyOverlayToggle struct{ win *gtk.Window }

func (t *spotifyOverlayToggle) Toggle() *dbus.Error {
	glib.IdleAdd(func() {
		if t.win.GetVisible() {
			t.win.Hide()
		} else {
			t.win.ShowAll()
			t.win.Present()
		}
	})
	return nil
}

func launchSpotifyOverlay() {
	exe, err := exec.LookPath("miracleos-desktop")
	if err == nil {
		_ = exec.Command(exe, "--spotify-overlay").Start()
	}
}
