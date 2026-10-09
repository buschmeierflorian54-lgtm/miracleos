package main

import "testing"

func TestSettingsFocusRevealsPicker(t *testing.T) {
	tree := desktopNode{Nodes: []desktopNode{{ID: 9, PID: 123, Name: "Einstellungen — MiracleOS"}, {ID: 10, PID: 123, Name: "Dateiauswahl — MiracleOS"}, {ID: 11, PID: 124, Name: "Einstellungen — MiracleOS"}}}
	if got := modalFocusTarget(tree, 9); got != 10 {
		t.Fatalf("blocked parent focused instead of picker: %d", got)
	}
	if got := modalFocusTarget(tree, 11); got != 11 {
		t.Fatalf("unrelated instance redirected: %d", got)
	}
	tree.Nodes = tree.Nodes[:1]
	if got := modalFocusTarget(tree, 9); got != 9 {
		t.Fatalf("closed picker still targeted: %d", got)
	}
}
