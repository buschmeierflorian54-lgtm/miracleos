package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

type loginReply struct {
	Type        string `json:"type"`
	Kind        string `json:"auth_message_type"`
	Message     string `json:"auth_message"`
	Description string `json:"description"`
}

func loginExchange(conn net.Conn, request map[string]interface{}) (loginReply, error) {
	var reply loginReply
	data, err := json.Marshal(request)
	if err != nil {
		return reply, err
	}
	conn.SetDeadline(time.Now().Add(45 * time.Second))
	if err = binary.Write(conn, binary.NativeEndian, uint32(len(data))); err != nil {
		return reply, err
	}
	if _, err = conn.Write(data); err != nil {
		return reply, err
	}
	var size uint32
	if err = binary.Read(conn, binary.NativeEndian, &size); err != nil {
		return reply, err
	}
	if size > 1024*1024 {
		return reply, fmt.Errorf("invalid login response")
	}
	data = make([]byte, size)
	if _, err = io.ReadFull(conn, data); err != nil {
		return reply, err
	}
	err = json.Unmarshal(data, &reply)
	return reply, err
}
func showLoginScreen() {
	user := os.Getenv("MIRACLEOS_LOGIN_USER")
 if user==""{data,err:=os.ReadFile("/etc/miracleos-account.json");if err==nil{var account struct{Username string `json:"username"`};if json.Unmarshal(data,&account)==nil{user=account.Username}}}
	if user == "" {
		user = "test"
	}
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetDecorated(false)
	win.SetTitle("MiracleOS Anmeldung")
	layershell.InitForWindow(win)
	layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_OVERLAY)
	layershell.SetKeyboardMode(win, layershell.LAYER_SHELL_KEYBOARD_MODE_EXCLUSIVE)
	for _, edge := range []layershell.LayerShellEdgeFlags{layershell.LAYER_SHELL_EDGE_TOP, layershell.LAYER_SHELL_EDGE_BOTTOM, layershell.LAYER_SHELL_EDGE_LEFT, layershell.LAYER_SHELL_EDGE_RIGHT} {
		layershell.SetAnchor(win, edge, true)
	}
	menuClass(win, "miracleos-login")
	provider, _ := gtk.CssProviderNew()
	provider.LoadFromData(`
 .miracleos-login {background-color:#e9d7e9;background-image:linear-gradient(rgba(250,240,250,.45),rgba(250,240,250,.65)),url('/usr/share/backgrounds/miracleos/light.png');background-size:cover;background-position:center;color:#3a2233;}
 .miracleos-login label {color:#3a2233;}
 .login-clock {font-size:36px;font-weight:bold;}
 .login-welcome {font-size:24px;font-weight:bold;}
 .login-status {font-size:13px;color:#62485c;}
 .miracleos-login entry {min-width:300px;min-height:28px;padding:10px 18px;border-radius:26px;background:rgba(255,255,255,.8);border:1px solid #b18aa6;color:#3a2233;caret-color:#3a2233;box-shadow:inset 0 1px rgba(255,255,255,.7);}
 .miracleos-login entry:focus {border-color:#a13c78;box-shadow:0 0 0 3px rgba(161,60,120,.15);}
 .miracleos-login button {min-height:30px;padding:9px 22px;border-radius:24px;background-image:linear-gradient(135deg,rgba(230,100,163,.65),rgba(137,45,109,.65));color:white;border:1px solid rgba(255,255,255,.35);box-shadow:inset 0 1px rgba(255,255,255,.2);}
 .miracleos-login button:hover {background-color:rgba(255,255,255,.15);}
 `)
	screen := win.GetScreen()
	gtk.AddProviderForScreen(screen, provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION+5)
	appearance, _ := gtk.CssProviderNew()
	gtk.AddProviderForScreen(screen, appearance, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION+6)
	applyAppearance := func() bool {
		p := systemPreferences{Accent:"#bf69a4"}
		if data, err := os.ReadFile(filepath.Join("/var/lib/miracleos/appearance", filepath.Base(user), "settings.json")); err == nil { _ = json.Unmarshal(data, &p) }
		if !regexp.MustCompile(`^#[0-9a-fA-F]{6}$`).MatchString(p.Accent) { p.Accent="#bf69a4" }
		mode, foreground, surface, veil := "light", "#3a2233", "#f8eff7", "rgba(250,240,250,.55)"
		if p.Dark { mode,foreground,surface,veil="dark","#fff0fb","#352337","rgba(25,12,32,.45)" }
		_ = appearance.LoadFromData(fmt.Sprintf(`
 .miracleos-login {background-color:%s;background-image:linear-gradient(%s,%s),url('/usr/share/backgrounds/miracleos/%s.png');color:%s;}
 .miracleos-login label,.miracleos-login image {color:%s;}
 .miracleos-login entry {background:%s;color:%s;caret-color:%s;border-color:%s;}
 .miracleos-login entry:focus {border-color:%s;}
 .miracleos-login button {background:%s;color:white;}
 .miracleos-login button label {color:white;}
 `,surface,veil,veil,mode,foreground,foreground,surface,foreground,foreground,p.Accent,p.Accent,p.Accent))
		return true
	}
	applyAppearance()
	glib.TimeoutAdd(1000, applyAppearance)
	root, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 18)
	root.SetHAlign(gtk.ALIGN_CENTER)
	root.SetVAlign(gtk.ALIGN_CENTER)
	clock := menuLabel(time.Now().Format("15:04"), "login-clock")
	clock.SetHAlign(gtk.ALIGN_CENTER)
	clock.SetXAlign(0.5)
	clock.SetMarginBottom(46)
	root.PackStart(clock, false, false, 0)
	glib.TimeoutAdd(1000, func() bool { clock.SetText(time.Now().Format("15:04")); return true })
	avatar := fmIcon("avatar-default-symbolic", gtk.ICON_SIZE_DIALOG)
	avatar.SetPixelSize(64)
	root.PackStart(avatar, false, false, 0)
	root.PackStart(menuLabel("Willkommen, "+user+"!", "login-welcome"), false, false, 0)
	entry, _ := gtk.EntryNew()
	entry.SetVisibility(false)
	entry.SetInputPurpose(gtk.INPUT_PURPOSE_PASSWORD)
	entry.SetPlaceholderText("Bitte Passwort eingeben")
	entry.SetSensitive(false)
	root.PackStart(entry, false, false, 0)
	submit, _ := gtk.ButtonNewWithLabel("Anmelden →")
	submit.SetSensitive(false)
	root.PackStart(submit, false, false, 0)
	status := menuLabel("Anmeldung wird vorbereitet …", "login-status")
	status.SetMaxWidthChars(45)
	status.SetLineWrap(true)
	root.PackStart(status, false, false, 0)
	win.Add(root)
	win.ShowAll()
	var conn net.Conn
	var send func(map[string]interface{})
	var process func(loginReply, error)
	send = func(request map[string]interface{}) {
		entry.SetSensitive(false)
		submit.SetSensitive(false)
		go func() { reply, err := loginExchange(conn, request); glib.IdleAdd(func() { process(reply, err) }) }()
	}
	start := func() { send(map[string]interface{}{"type": "create_session", "username": user}) }
	process = func(reply loginReply, err error) {
		if err != nil {
			status.SetText("Anmeldung konnte nicht vorbereitet werden. Bitte neu starten.")
			return
		}
		switch reply.Type {
		case "auth_message":
			if reply.Kind == "info" || reply.Kind == "error" {
				if reply.Kind == "error" {
					status.SetText("Bitte Passwort prüfen.")
				}
				send(map[string]interface{}{"type": "post_auth_message_response", "response": nil})
				return
			}
			entry.SetVisibility(reply.Kind == "visible")
			entry.SetSensitive(true)
			submit.SetSensitive(true)
			entry.GrabFocus()
			if reply.Kind == "secret" {
				entry.SetPlaceholderText("Bitte Passwort eingeben")
			} else {
				entry.SetPlaceholderText(reply.Message)
			}
			if statusText, _ := status.GetText(); statusText == "Anmeldung wird vorbereitet …" {
				status.SetText("Mit deinem MiracleOS-Passwort anmelden")
			}
		case "success":
			status.SetText("Dein Desktop wird gestartet …")
			go func() {
				reply, err := loginExchange(conn, map[string]interface{}{"type": "start_session", "cmd": []string{"/usr/local/bin/miracleos-session"}, "env": []string{}})
				glib.IdleAdd(func() {
					if err == nil && reply.Type == "success" {
						gtk.MainQuit()
					} else {
						status.SetText("Desktop konnte nicht gestartet werden.")
					}
				})
			}()
		case "error":
			status.SetText("Passwort nicht akzeptiert. Bitte erneut versuchen.")
			entry.SetText("")
			go func() {
				_, err := loginExchange(conn, map[string]interface{}{"type": "cancel_session"})
				glib.IdleAdd(func() {
					if err == nil {
						start()
					} else {
						status.SetText("Bitte die Anmeldung neu starten.")
					}
				})
			}()
		}
	}
	authenticate := func() {
		if !submit.GetSensitive() {
			return
		}
		password, _ := entry.GetText()
		if password == "" {
			return
		}
		entry.SetText("")
		status.SetText("Passwort wird geprüft …")
		send(map[string]interface{}{"type": "post_auth_message_response", "response": password})
	}
	submit.Connect("clicked", authenticate)
	entry.Connect("activate", authenticate)
	go func() {
		var err error
		conn, err = net.Dial("unix", os.Getenv("GREETD_SOCK"))
		glib.IdleAdd(func() {
			if err != nil {
				status.SetText("Anmeldedienst nicht erreichbar.")
				return
			}
			start()
		})
	}()
	gtk.Main()
	if conn != nil {
		conn.Close()
	}
}
