package main

import (
	"strings"
	"testing"
)

func TestSharedContextMenuValidation(t *testing.T) {
	valid := `{"version":1,"title":"Bild","items":[{"id":"copy","label":"Kopieren"},{"separator":true},{"id":"paste","label":"Einfügen","enabled":false}]}`
	request, err := decodeSharedContextMenu(strings.NewReader(valid))
	if err != nil {
		t.Fatal(err)
	}
	if request.Items[0].Enabled != nil || request.Items[2].Enabled == nil || *request.Items[2].Enabled {
		t.Fatal("enabled defaults/state lost")
	}
	invalid := []string{
		`{"version":2,"items":[{"id":"x","label":"X"}]}`,
		`{"version":1,"items":[{"id":"x","label":"X","command":"echo bad"}]}`,
		`{"version":1,"items":[{"id":"x","label":"X"},{"id":"x","label":"Y"}]}`,
		`{"version":1,"items":[{"separator":true}]}`,
		`{"version":1,"items":[{"id":"x","label":"X\nY"}]}`,
		valid + `{}`,
		strings.Repeat(" ", sharedContextMenuLimit+1),
	}
	for _, input := range invalid {
		if _, err := decodeSharedContextMenu(strings.NewReader(input)); err == nil {
			t.Fatalf("accepted invalid request: %.100s", input)
		}
	}
}
