package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type windowDropRequest struct {
	Paths []string `json:"paths"`
	Move  bool     `json:"move"`
}
type windowDropReply struct {
	OK bool `json:"ok"`
}

func windowDropSocket(pid int) string {
	return filepath.Join(os.Getenv("XDG_RUNTIME_DIR"), fmt.Sprintf("miracleos-drop-%d.sock", pid))
}
func registerWindowDrop(win *gtk.Window, receive func(windowDropRequest, func(error))) {
	path := windowDropSocket(os.Getpid())
	listener, err := net.Listen("unix", path)
	if err != nil {
		return
	}
	os.Chmod(path, 0600)
	alive := true
	win.Connect("destroy", func() { alive = false; listener.Close(); os.Remove(path) })
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.SetDeadline(time.Now().Add(5 * time.Second))
			var request windowDropRequest
			err = json.NewDecoder(io.LimitReader(conn, 1024*1024)).Decode(&request)
			if err != nil || len(request.Paths) == 0 || len(request.Paths) > 256 {
				conn.Close()
				continue
			}
			valid := true
			for _, path := range request.Paths {
				if !filepath.IsAbs(path) {
					valid = false
				}
			}
			if !valid {
				conn.Close()
				continue
			}
			complete := func(err error) {
				go func() {
					defer conn.Close()
					conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
					json.NewEncoder(conn).Encode(windowDropReply{OK: err == nil})
				}()
			}
			glib.IdleAdd(func() {
				if alive {
					receive(request, complete)
				} else {
					complete(fmt.Errorf("window closed"))
				}
			})
		}
	}()
}
func sendWindowDrop(pid int, request windowDropRequest) error {
	conn, err := net.DialTimeout("unix", windowDropSocket(pid), 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(60 * time.Second))
	if err = json.NewEncoder(conn).Encode(request); err != nil {
		return err
	}
	var reply windowDropReply
	if err = json.NewDecoder(conn).Decode(&reply); err != nil {
		return err
	}
	if !reply.OK {
		return fmt.Errorf("window rejected drop")
	}
	return nil
}
func (fm *glassFileManager) receiveWindowDrop(request windowDropRequest, complete func(error)) {
	if fm.picker != nil {
		if fm.picker.Save {
			fm.navigate(filepath.Dir(request.Paths[0]), true)
			if fm.pickerName != nil {
				fm.pickerName.SetText(filepath.Base(request.Paths[0]))
			}
			complete(nil)
			return
		}
		if !fm.picker.Multiple && len(request.Paths) != 1 {
			complete(fmt.Errorf("only one file accepted"))
			return
		}
		for _, path := range request.Paths {
			info, err := os.Stat(path)
			if err != nil || info.IsDir() != fm.picker.Directory || (!info.IsDir() && !fm.picker.accepts(path)) {
				complete(fmt.Errorf("unsupported file format"))
				return
			}
		}
		complete(nil)
		fm.picker.finish(request.Paths)
		fm.window.Destroy()
		return
	}
	destination := fm.current
	fm.status.SetText("Dateien werden übertragen …")
	go func() {
		_, err := pasteFilePaths(request.Paths, request.Move, destination)
		complete(err)
		glib.IdleAdd(func() {
			if !fm.alive.Load() {
				return
			}
			fm.render()
			if err != nil {
				fm.status.SetText("Übertragen fehlgeschlagen: " + err.Error())
			}
		})
	}()
}

func focusDropWindow(id int64) { go focusDropWindowSync(id) }
func focusDropWindowSync(id int64) {
	action := "focus"
	if tree, err := desktopTree(); err == nil {
		for _, node := range minimizedWindowsFromTree(tree) {
			if node.ID == id {
				action = "scratchpad show, focus"
			}
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	exec.CommandContext(ctx, "swaymsg", fmt.Sprintf("[con_id=%d] %s", id, action)).Run()
}
