package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func startSetupSpotify() error {
	if tree, err := desktopTree(); err == nil {
		if node, ok := setupSpotifyWindow(tree); ok {
			return exec.Command("swaymsg", fmt.Sprintf("[con_id=%d] focus", node.ID)).Run()
		}
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(cache, "miracleos")
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	log, err := os.OpenFile(filepath.Join(dir, "spotify-start.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	args := []string{}
	if exec.Command("systemd-detect-virt", "--quiet").Run() == nil {
		args = append(args, "--disable-gpu")
	}
	cmd := exec.Command("spotify", args...)
	cmd.Stdout, cmd.Stderr = log, log
	if err = cmd.Start(); err != nil {
		log.Close()
		return err
	}
	go func() { _ = cmd.Wait(); _ = log.Close() }()
	return nil
}

func setupSpotifyWindow(node desktopNode) (desktopNode, bool) {
	if strings.EqualFold(node.AppID, "spotify") || strings.EqualFold(node.WindowProperties.Class, "Spotify") {
		return node, true
	}
	for _, children := range [][]desktopNode{node.Nodes, node.FloatingNodes} {
		for _, child := range children {
			if found, ok := setupSpotifyWindow(child); ok {
				return found, true
			}
		}
	}
	return desktopNode{}, false
}

// The native Spotify client is borderless only during setup. Its temporary GTK
// header owns close confirmation and exposes no usable minimize control.
func showSetupSpotify(setup *gtk.Window, finished func(bool)) {
	width, height := setup.GetSize()
	spotifyW, spotifyH := min(960, width*62/100), min(650, height*66/100)
	x, y := width-spotifyW-36, (height-spotifyH)/2+18
	newLayer := func(namespace string, layer layershell.LayerShellLayerFlags) *gtk.Window {
		w, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
		layershell.InitForWindow(w)
		layershell.SetNamespace(w, namespace)
		layershell.SetLayer(w, layer)
		layershell.SetExclusiveZone(w, -1)
		w.SetAppPaintable(true)
		if screen := w.GetScreen(); screen != nil {
			if visual, e := screen.GetRGBAVisual(); e == nil {
				w.SetVisual(visual)
			}
		}
		return w
	}
	allEdges := func(w *gtk.Window) {
		layershell.SetAnchor(w, layershell.LAYER_SHELL_EDGE_TOP, true)
		layershell.SetAnchor(w, layershell.LAYER_SHELL_EDGE_BOTTOM, true)
		layershell.SetAnchor(w, layershell.LAYER_SHELL_EDGE_LEFT, true)
		layershell.SetAnchor(w, layershell.LAYER_SHELL_EDGE_RIGHT, true)
	}
	backdrop := newLayer("miracleos-setup-backdrop", layershell.LAYER_SHELL_LAYER_BOTTOM)
	allEdges(backdrop)
	canvas, _ := gtk.DrawingAreaNew()
	canvas.Connect("draw", func(_ *gtk.DrawingArea, cr *cairo.Context) bool {
		w, h := float64(canvas.GetAllocatedWidth()), float64(canvas.GetAllocatedHeight())
		cr.SetSourceRGB(.72, .55, .67)
		cr.Paint()
		for _, b := range [][5]float64{{.25, .28, .88, .20, .55}, {.69, .25, 1, .65, .20}, {.28, .74, .12, .76, .73}, {.79, .73, .48, .20, .94}} {
			for i := 60; i > 0; i-- {
				r := float64(i) / 60
				cr.SetSourceRGBA(b[2], b[3], b[4], .032*math.Exp(-r*r*2))
				cr.Arc(w*b[0], h*b[1], math.Min(w, h)*.43*r, 0, 2*math.Pi)
				cr.Fill()
			}
		}
		cr.SetSourceRGB(.99, .93, .97)
		cr.Rectangle(float64(x-5), float64(y-5), float64(spotifyW+10), float64(spotifyH+10))
		cr.Fill()
		return false
	})
	backdrop.Add(canvas)
	side := newLayer("miracleos-setup-sidebar", layershell.LAYER_SHELL_LAYER_OVERLAY)
	layershell.SetAnchor(side, layershell.LAYER_SHELL_EDGE_LEFT, true)
	layershell.SetMargin(side, layershell.LAYER_SHELL_EDGE_LEFT, 32)
	layershell.SetKeyboardMode(side, layershell.LAYER_SHELL_KEYBOARD_MODE_ON_DEMAND)
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 20)
	box.SetSizeRequest(max(180, x-76), -1)
	menuClass(box, "setup-spotify-sidebar")
	title := menuLabel("Spotify-Einrichtung", "setup-heading")
	title.SetLineWrap(true)
	title.SetMaxWidthChars(17)
	title.SetXAlign(0)
	box.PackStart(title, false, false, 0)
	text := settingsText("Melde dich in Spotify an. Spiele danach einen Song ab, um deine Musiksteuerung auszuprobieren.")
	text.SetMaxWidthChars(25)
	box.PackStart(text, false, false, 0)
	progress := settingsText("① Bei Spotify anmelden\n\n② Einen Song abspielen")
	box.PackStart(progress, false, false, 0)
	header := newLayer("miracleos-setup-spotify-header", layershell.LAYER_SHELL_LAYER_OVERLAY)
	qrHint := newLayer("miracleos-setup-spotify-qr-hint", layershell.LAYER_SHELL_LAYER_OVERLAY)
	layershell.SetAnchor(qrHint, layershell.LAYER_SHELL_EDGE_TOP, true)
	layershell.SetAnchor(qrHint, layershell.LAYER_SHELL_EDGE_RIGHT, true)
	layershell.SetMargin(qrHint, layershell.LAYER_SHELL_EDGE_TOP, y+spotifyH+18)
	layershell.SetMargin(qrHint, layershell.LAYER_SHELL_EDGE_RIGHT, 36)
	qrLabel := menuLabel("Bitte den Spotify-QR-Code scannen.", "setup-spotify-qr-hint")
	qrLabel.SetSizeRequest(spotifyW, -1)
	qrLabel.SetXAlign(.5)
	qrHint.Add(qrLabel)
	layershell.SetAnchor(header, layershell.LAYER_SHELL_EDGE_TOP, true)
	layershell.SetAnchor(header, layershell.LAYER_SHELL_EDGE_RIGHT, true)
	layershell.SetMargin(header, layershell.LAYER_SHELL_EDGE_TOP, y-38)
	layershell.SetMargin(header, layershell.LAYER_SHELL_EDGE_RIGHT, 31)
	head, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 12)
	head.SetSizeRequest(spotifyW+10, 38)
	menuClass(head, "setup-spotify-header")
	// Match compositor/miracleos-header.h: two rounded tabs with a lowered,
	// transparent middle, connected by the thin edge above the app content.
	head.Connect("draw", func(_ *gtk.Box, cr *cairo.Context) bool {
		w, h := float64(head.GetAllocatedWidth()), float64(head.GetAllocatedHeight())
		cr.SetSourceRGB(.99, .93, .97)
		drawSetupSpotifyTab(cr, 0, math.Min(300, w-130), h, true)
		cr.Fill()
		drawSetupSpotifyTab(cr, w-114, 114, h, false)
		cr.Fill()
		cr.Rectangle(0, h-3, w, 3)
		cr.Fill()
		return false
	})
	head.PackStart(menuLabel("●  Spotify", "setup-spotify-title"), false, false, 0)
	alive, confirmOpen, played := true, false, false
	stop := make(chan struct{})
	var confirmation *gtk.Window
	restore := func() {
		_ = exec.Command("swaymsg", `[class="(?i)spotify"] fullscreen disable, border normal, move scratchpad; [app_id="(?i)spotify"] fullscreen disable, border normal, move scratchpad`).Run()
	}
	finish := func() {
		if !alive {
			return
		}
		alive = false
		close(stop)
		if confirmation != nil {
			confirmation.Destroy()
		}
		header.Destroy()
		qrHint.Destroy()
		side.Destroy()
		backdrop.Destroy()
		restore()
		finished(played)
		setup.Show()
		setup.Fullscreen()
		setup.Present()
	}
	var askLater func()
	askLater = func() {
		if !alive || confirmOpen {
			return
		}
		confirmOpen = true
		confirmation = newLayer("miracleos-setup-spotify-confirm", layershell.LAYER_SHELL_LAYER_OVERLAY)
		allEdges(confirmation)
		layershell.SetKeyboardMode(confirmation, layershell.LAYER_SHELL_KEYBOARD_MODE_ON_DEMAND)
		shade, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
		menuClass(shade, "setup-confirm-shade")
		card, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 24)
		card.SetHAlign(gtk.ALIGN_CENTER)
		card.SetVAlign(gtk.ALIGN_CENTER)
		menuClass(card, "setup-card")
		menuClass(card, "setup-spotify-confirm-card")
		question := menuLabel("Willst du Spotify später einrichten?", "setup-heading")
		question.SetLineWrap(true)
		question.SetMaxWidthChars(28)
		card.PackStart(question, false, false, 0)
		buttons, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 16)
		buttons.PackStart(settingsButton("Jetzt einrichten", func() { confirmation.Destroy(); confirmation = nil; confirmOpen = false; _ = startSetupSpotify() }), false, false, 0)
		buttons.PackStart(settingsButton("Später einrichten", finish), false, false, 0)
		card.PackStart(buttons, false, false, 0)
		shade.PackStart(card, true, true, 0)
		confirmation.Add(shade)
		confirmation.ShowAll()
	}
	closeButton := settingsButton("×", askLater)
	closeButton.SetTooltipText("Spotify später einrichten")
	menuClass(closeButton, "setup-spotify-close")
	head.PackEnd(closeButton, false, false, 0)
	minimize := settingsButton("⌄", func() {})
	minimize.SetSensitive(false)
	minimize.SetTooltipText("Während der Einrichtung nicht verfügbar")
	head.PackEnd(minimize, false, false, 0)
	header.Add(head)
	retry := settingsButton("Spotify erneut öffnen", func() {
		if err := startSetupSpotify(); err != nil {
			progress.SetText("Spotify konnte nicht gestartet werden.")
		}
	})
	retry.SetNoShowAll(true)
	box.PackStart(retry, false, false, 0)
	box.PackStart(settingsButton("Später einrichten", askLater), false, false, 0)
	side.Add(box)
	setup.Hide()
	backdrop.ShowAll()
	side.ShowAll()
	header.ShowAll()
	qrHint.ShowAll()
	if err := startSetupSpotify(); err != nil {
		progress.SetText("Spotify konnte nicht gestartet werden.")
		retry.Show()
	}
	go func() {
		var placed int64
		attempts := 0
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
			}
			tree, err := desktopTree()
			if err != nil {
				continue
			}
			node, found := setupSpotifyWindow(tree)
			snapshot := readSpotifyMedia()
			attempts++
			missing := !found && attempts > 15
			glib.IdleAdd(func() {
				if !alive {
					return
				}
			// Serialize placement and teardown on the GTK thread: an old poll
			// must never bring Spotify back after the stage has been dismissed.
			if found && (placed != node.ID || node.FullscreenMode != 0 || node.Rect.X != x || node.Rect.Y != y || node.Rect.Width != spotifyW || node.Rect.Height != spotifyH) {
				command := fmt.Sprintf("[con_id=%d] floating enable, fullscreen disable, border none, resize set %d px %d px, move absolute position %d px %d px", node.ID, spotifyW, spotifyH, x, y)
				if placed != node.ID {
					command += ", focus"
				}
				_ = exec.Command("swaymsg", command).Run()
				placed = node.ID
			}
				if snapshot.Playing && found && !played {
					played = true
					_ = exec.Command("swaymsg", fmt.Sprintf("[con_id=%d] move scratchpad", node.ID)).Run()
					finish()
					return
				}
				if missing {
					progress.SetText("Spotify-Fenster geschlossen oder nicht gestartet.")
					retry.Show()
				} else if found {
					retry.Hide()
					if played {
						progress.SetText("✓ Angemeldet\n\n✓ Song ausprobiert")
					} else {
						progress.SetText("① Bei Spotify anmelden\n\n② Einen Song abspielen")
					}
				}
			})
		}
	}()
}

// Keep the radius and curved joins identical to the normal MiracleOS app tabs.
func drawSetupSpotifyTab(cr *cairo.Context, x, width, height float64, left bool) {
	const r, join = 15.0, 12.0
	cr.NewPath()
	cr.MoveTo(x+r, 0)
	cr.LineTo(x+width-r, 0)
	cr.CurveTo(x+width, 0, x+width, r/2, x+width, r)
	if left {
		cr.LineTo(x+width, height-join)
		cr.CurveTo(x+width, height, x+width+join/2, height, x+width+join, height)
		cr.LineTo(x, height)
	} else {
		cr.LineTo(x+width, height)
		cr.LineTo(x-join, height)
		cr.CurveTo(x-join/2, height, x, height, x, height-join)
	}
	cr.LineTo(x, r)
	cr.CurveTo(x, r/2, x, 0, x+r, 0)
	cr.ClosePath()
}
