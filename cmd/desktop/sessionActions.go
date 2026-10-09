package main

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"os"
	"os/exec"
	"path/filepath"
)

func sessionAction(action string) {
	hideStartMenu()
	go func() {
		var command *exec.Cmd
		if action == "lock" {
			binary, _ := os.Executable()
			command = exec.Command("python3", filepath.Join(filepath.Dir(binary), "runtime", "lockScreen.py"))
		} else {
			command = exec.Command("systemctl", action)
		}
		if err := command.Run(); err != nil {
			glib.IdleAdd(func() {
				dialog := gtk.MessageDialogNew(nil, 0, gtk.MESSAGE_ERROR, gtk.BUTTONS_CLOSE, "Die Aktion konnte nicht ausgeführt werden: %s", err)
				dialog.Connect("response", func() { dialog.Destroy() })
				dialog.ShowAll()
			})
		}
	}()
}
