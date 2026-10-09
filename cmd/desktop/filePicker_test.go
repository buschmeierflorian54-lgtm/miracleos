package main

import (
	"github.com/godbus/dbus/v5"
	"os"
	"path/filepath"
	"testing"
)

func TestPickerFilters(t *testing.T) {
	cases := []struct {
		rule pickerRule
		name string
		want bool
	}{{pickerRule{0, "*.PNG"}, "photo.png", true}, {pickerRule{0, "*.png"}, "notes.pdf", false}, {pickerRule{1, "image/*"}, "photo.jpg", true}, {pickerRule{1, "application/pdf"}, "report.pdf", true}, {pickerRule{1, "image/png"}, "report.pdf", false}, {pickerRule{0, "*.deb"}, "steam.deb", true}}
	for _, c := range cases {
		p := filePickerOptions{Filters: []pickerFilter{{Name: "Test", Rules: []pickerRule{c.rule}}}}
		if got := p.accepts(c.name); got != c.want {
			t.Errorf("%+v %s: %v", c.rule, c.name, got)
		}
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "unknown")
	os.WriteFile(path, []byte("plain text"), 0600)
	p := filePickerOptions{Filters: []pickerFilter{{Name: "Text", Rules: []pickerRule{{1, "text/plain"}}}}}
	if !p.accepts(path) {
		t.Error("MIME detection without extension failed")
	}
}
func TestPortalFilterDecoding(t *testing.T) {
	raw := []interface{}{[]interface{}{"Pictures", []interface{}{[]interface{}{uint32(1), "image/*"}, []interface{}{uint32(0), "*.webp"}}}}
	var filters []pickerFilter
	options := map[string]dbus.Variant{"filters": dbus.MakeVariant(raw)}
	portalValue(options, "filters", &filters)
	if len(filters) != 1 || len(filters[0].Rules) != 2 || filters[0].Rules[1].Value != "*.webp" {
		t.Fatalf("Portal filters lost: %#v", filters)
	}
}
