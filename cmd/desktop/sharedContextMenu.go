package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const sharedContextMenuLimit = 1024 * 1024

type sharedContextMenuItem struct {
	ID        string `json:"id,omitempty"`
	Label     string `json:"label,omitempty"`
	Enabled   *bool  `json:"enabled,omitempty"`
	Separator bool   `json:"separator,omitempty"`
}
type sharedContextMenuRequest struct {
	Version int                     `json:"version"`
	Title   string                  `json:"title,omitempty"`
	Items   []sharedContextMenuItem `json:"items"`
}
type sharedContextMenuResponse struct {
	Version int    `json:"version"`
	Status  string `json:"status"`
	ID      string `json:"id,omitempty"`
	Error   string `json:"error,omitempty"`
}

func decodeSharedContextMenu(input io.Reader) (sharedContextMenuRequest, error) {
	var request sharedContextMenuRequest
	data, err := io.ReadAll(io.LimitReader(input, sharedContextMenuLimit+1))
	if err != nil {
		return request, err
	}
	if len(data) > sharedContextMenuLimit {
		return request, errors.New("request exceeds 1 MiB")
	}
	if !utf8.Valid(data) {
		return request, errors.New("request is not UTF-8")
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&request); err != nil {
		return request, err
	}
	var extra interface{}
	if err = decoder.Decode(&extra); err != io.EOF {
		return request, errors.New("expected one JSON request")
	}
	if request.Version != 1 {
		return request, errors.New("unsupported protocol version")
	}
	if len(request.Items) == 0 || len(request.Items) > 100 {
		return request, errors.New("items must contain between 1 and 100 entries")
	}
	validText := func(text string, max int) bool {
		return len(text) <= max && strings.IndexFunc(text, unicode.IsControl) < 0
	}
	if !validText(request.Title, 160) {
		return request, errors.New("invalid title")
	}
	seen := map[string]bool{}
	actions := 0
	for _, item := range request.Items {
		if item.Separator {
			if item.ID != "" || item.Label != "" || item.Enabled != nil {
				return request, errors.New("separator must not contain action fields")
			}
			continue
		}
		if strings.TrimSpace(item.ID) == "" || !validText(item.ID, 128) || seen[item.ID] {
			return request, errors.New("action IDs must be unique non-empty text")
		}
		if strings.TrimSpace(item.Label) == "" || !validText(item.Label, 240) {
			return request, errors.New("invalid action label")
		}
		seen[item.ID] = true
		actions++
	}
	if actions == 0 {
		return request, errors.New("menu requires at least one action")
	}
	return request, nil
}

// The caller owns all actions. This process only presents labels and returns an
// opaque ID; it never interprets IDs as paths, commands, or URLs.
func runSharedContextMenu() {
	request, err := decodeSharedContextMenu(os.Stdin)
	if err != nil {
		json.NewEncoder(os.Stdout).Encode(sharedContextMenuResponse{Version: 1, Status: "error", Error: err.Error()})
		return
	}
	gtk.Init(nil)
	loadCSS()
	provider, _ := gtk.CssProviderNew()
	if err := provider.LoadFromData(systemThemeCSS(readSystemPreferences())); err == nil {
		if screen, err := gdk.ScreenGetDefault(); err == nil {
			gtk.AddProviderForScreen(screen, provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION+1)
		}
	}
	anchor, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	anchor.SetTitle("MiracleOS Kontextmenü")
	anchor.SetDecorated(false)
	anchor.SetResizable(false)
	anchor.SetDefaultSize(1, 1)
	anchor.SetSkipTaskbarHint(true)
	anchor.SetSkipPagerHint(true)
	anchor.SetPosition(gtk.WIN_POS_MOUSE)
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		// A process adapter cannot inherit the caller's Wayland surface/input serial.
		// Use an unobtrusive centered layer surface instead of a decorated host window.
		layershell.InitForWindow(anchor)
		layershell.SetNamespace(anchor, "miracleos-context-menu")
		layershell.SetLayer(anchor, layershell.LAYER_SHELL_LAYER_OVERLAY)
		layershell.SetExclusiveZone(anchor, 0)
		layershell.SetKeyboardMode(anchor, layershell.LAYER_SHELL_KEYBOARD_MODE_EXCLUSIVE)
		anchor.SetSizeRequest(1, 1)
	}
	var menu *gtk.Menu
	finished := false
	finish := func(response sharedContextMenuResponse) {
		if finished {
			return
		}
		finished = true
		if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		// Destroy after GTK finishes dispatching the activation signal.
		glib.IdleAdd(func() { menu.Destroy(); anchor.Destroy(); gtk.MainQuit() })
	}
	actions := []contextAction{}
	if request.Title != "" {
		actions = append(actions, contextAction{label: request.Title, disabled: true}, contextAction{})
	}
	for _, item := range request.Items {
		if item.Separator {
			actions = append(actions, contextAction{})
			continue
		}
		id := item.ID
		actions = append(actions, contextAction{label: item.Label, disabled: item.Enabled != nil && !*item.Enabled, run: func() { finish(sharedContextMenuResponse{Version: 1, Status: "selected", ID: id}) }})
	}
	actions = append(actions, contextAction{}, contextAction{label: "Abbrechen", run: func() { finish(sharedContextMenuResponse{Version: 1, Status: "cancelled"}) }})
	menu = buildSharedMenu(actions)
	_ = menu.SetProperty("attach-widget", anchor.Object)
	cancel := func() { finish(sharedContextMenuResponse{Version: 1, Status: "cancelled"}) }
	menu.Connect("selection-done", cancel)
	anchor.Connect("delete-event", func() bool { cancel(); return true })
	anchor.Connect("map-event", func() bool {
		glib.IdleAdd(func() {
			menu.ShowAll()
			menu.PopupAtWidget(anchor, gdk.GDK_GRAVITY_SOUTH_WEST, gdk.GDK_GRAVITY_NORTH_WEST, nil)
		})
		return false
	})
	anchor.Connect("key-press-event", func(_ *gtk.Window, event *gdk.Event) bool {
		if gdk.EventKeyNewFromEvent(event).KeyVal() == gdk.KEY_Escape {
			cancel()
			return true
		}
		return false
	})
	anchor.ShowAll()
	gtk.Main()
}

// Internal applications and opt-in external adapters share the same renderer.
func buildSharedMenu(actions []contextAction) *gtk.Menu {
	menu, _ := gtk.MenuNew()
	menuClass(menu, "glass-context-menu")
	for _, action := range actions {
		if action.label == "" {
			separator, _ := gtk.SeparatorMenuItemNew()
			menu.Append(separator)
			continue
		}
		row, _ := gtk.MenuItemNewWithLabel(action.label)
		row.SetSensitive(!action.disabled)
		run := action.run
		row.Connect("activate", func() {
			if run != nil {
				run()
			}
		})
		menu.Append(row)
	}
	return menu
}
