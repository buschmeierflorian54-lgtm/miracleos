package main

import (
	"fmt"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
	"github.com/gotk3/gotk3/glib"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const pickerInterface = "org.freedesktop.impl.portal.FileChooser"

type filePortal struct{ conn *dbus.Conn }
type portalRequest struct {
	mu     sync.Mutex
	closed bool
	cancel func()
}

func (r *portalRequest) Close() *dbus.Error {
	r.mu.Lock()
	r.closed = true
	cancel := r.cancel
	r.mu.Unlock()
	if cancel != nil {
		glib.IdleAdd(cancel)
	}
	return nil
}
func portalTuple(value interface{}) []interface{} {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil
	}
	for v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	result := []interface{}{}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			result = append(result, v.Index(i).Interface())
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			result = append(result, v.Field(i).Interface())
		}
	}
	return result
}
func decodePickerFilter(value interface{}) (pickerFilter, bool) {
	fields := portalTuple(value)
	if len(fields) != 2 {
		return pickerFilter{}, false
	}
	name, ok := fields[0].(string)
	if !ok {
		return pickerFilter{}, false
	}
	filter := pickerFilter{Name: name}
	for _, raw := range portalTuple(fields[1]) {
		rule := portalTuple(raw)
		if len(rule) != 2 {
			return pickerFilter{}, false
		}
		kind, err := strconv.ParseUint(fmt.Sprint(rule[0]), 10, 32)
		text, ok := rule[1].(string)
		if err != nil || !ok || kind > 1 {
			return pickerFilter{}, false
		}
		filter.Rules = append(filter.Rules, pickerRule{uint32(kind), text})
	}
	return filter, true
}
func portalValue(options map[string]dbus.Variant, key string, target interface{}) {
	if value, ok := options[key]; ok {
		switch dest := target.(type) {
		case *[]pickerFilter:
			for _, raw := range portalTuple(value.Value()) {
				if f, ok := decodePickerFilter(raw); ok {
					*dest = append(*dest, f)
				}
			}
			return
		case *pickerFilter:
			if f, ok := decodePickerFilter(value.Value()); ok {
				*dest = f
			}
			return
		}
		_ = dbus.Store([]interface{}{value.Value()}, target)
	}
}
func portalFolder(options map[string]dbus.Variant, key string) string {
	var bytes []byte
	portalValue(options, key, &bytes)
	return strings.TrimRight(string(bytes), "\x00")
}
func (portal *filePortal) OpenFile(handle dbus.ObjectPath, app, parent, title string, options map[string]dbus.Variant) (uint32, map[string]dbus.Variant, *dbus.Error) {
	return portal.choose(handle, title, options, false, false)
}
func (portal *filePortal) SaveFile(handle dbus.ObjectPath, app, parent, title string, options map[string]dbus.Variant) (uint32, map[string]dbus.Variant, *dbus.Error) {
	return portal.choose(handle, title, options, true, false)
}
func (portal *filePortal) SaveFiles(handle dbus.ObjectPath, app, parent, title string, options map[string]dbus.Variant) (uint32, map[string]dbus.Variant, *dbus.Error) {
	return portal.choose(handle, title, options, false, true)
}
func (portal *filePortal) choose(handle dbus.ObjectPath, title string, options map[string]dbus.Variant, save, batch bool) (uint32, map[string]dbus.Variant, *dbus.Error) {
	result := make(chan []string, 1)
	req := &portalRequest{}
	if err := portal.conn.Export(req, handle, "org.freedesktop.impl.portal.Request"); err != nil {
		return 2, nil, dbus.MakeFailedError(err)
	}
	defer portal.conn.Export(nil, handle, "org.freedesktop.impl.portal.Request")
	p := &filePickerOptions{Title: title, Save: save, Folder: portalFolder(options, "current_folder"), Done: func(paths []string) { result <- paths }}
	portalValue(options, "multiple", &p.Multiple)
	portalValue(options, "directory", &p.Directory)
	portalValue(options, "current_name", &p.Name)
	portalValue(options, "accept_label", &p.AcceptLabel)
	portalValue(options, "filters", &p.Filters)
	var current pickerFilter
	portalValue(options, "current_filter", &current)
	for i, f := range p.Filters {
		if f.Name == current.Name {
			p.Active = i
			break
		}
	}
	if file := portalFolder(options, "current_file"); file != "" {
		p.Folder = filepath.Dir(file)
		p.Name = filepath.Base(file)
	}
	if batch {
		p.Directory = true
		p.Multiple = false
		p.Filters = nil
	}
	glib.IdleAdd(func() {
		req.mu.Lock()
		if req.closed {
			req.mu.Unlock()
			p.finish(nil)
			return
		}
		req.mu.Unlock()
		fm := showFilePicker(p)
		req.mu.Lock()
		req.cancel = func() { fm.window.Destroy() }
		closed := req.closed
		req.mu.Unlock()
		if closed {
			fm.window.Destroy()
		}
	})
	paths := <-result
	if len(paths) == 0 {
		return 1, map[string]dbus.Variant{}, nil
	}
	if batch {
		var files [][]byte
		portalValue(options, "files", &files)
		folder := paths[0]
		paths = nil
		for _, bytes := range files {
			name := strings.TrimRight(string(bytes), "\x00")
			if name == "" || name != filepath.Base(name) || name == "." || name == ".." {
				return 2, map[string]dbus.Variant{}, nil
			}
			path := filepath.Join(folder, name)
			stem := strings.TrimSuffix(name, filepath.Ext(name))
			for i := 1; ; i++ {
				if _, err := os.Lstat(path); os.IsNotExist(err) {
					break
				}
				path = filepath.Join(folder, fmt.Sprintf("%s (%d)%s", stem, i, filepath.Ext(name)))
			}
			paths = append(paths, path)
		}
	}
	uris := []string{}
	for _, path := range paths {
		uris = append(uris, (&url.URL{Scheme: "file", Path: path}).String())
	}
	results := map[string]dbus.Variant{"uris": dbus.MakeVariant(uris)}
	if len(p.Filters) > 0 && p.Active >= 0 && p.Active < len(p.Filters) {
		results["current_filter"] = dbus.MakeVariant(p.Filters[p.Active])
	}
	return 0, results, nil
}
func runFilePortal() {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		log.Fatal(err)
	}
	reply, err := conn.RequestName("org.freedesktop.impl.portal.desktop.miracleos", dbus.NameFlagDoNotQueue)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		log.Fatal("File portal is already running or unavailable: ", err)
	}
	backend := &filePortal{conn: conn}
	path := dbus.ObjectPath("/org/freedesktop/portal/desktop")
	if err = conn.Export(backend, path, pickerInterface); err != nil {
		log.Fatal(err)
	}
	_, err = prop.Export(conn, path, map[string]map[string]*prop.Prop{pickerInterface: {"version": {Value: uint32(3), Writable: false, Emit: prop.EmitConst}}})
	if err != nil {
		log.Fatal(err)
	}
	conn.Export(introspect.NewIntrospectable(&introspect.Node{Interfaces: []introspect.Interface{{Name: pickerInterface, Methods: introspect.Methods(backend)}, introspect.IntrospectData}}), path, "org.freedesktop.DBus.Introspectable")
}
