package main

import "testing"

func TestWindowColorLifecycle(t *testing.T) {
	windowContextColors = map[int64]string{10: "#329df0"}
	defer func() { windowContextColors = map[int64]string{} }()
	snapshot := taskbarSnapshot{Windows: []taskbarEntry{{ID: 10}, {ID: 11}}}
	pruneWindowColors(snapshot)
	if windowContextColors[10] == "" || windowContextColors[11] != "" {
		t.Fatal("second window inherited context color")
	}
	pruneWindowColors(taskbarSnapshot{Windows: []taskbarEntry{{ID: 11}}, Minimized: []desktopNode{{ID: 10}}})
	if windowContextColors[10] == "" {
		t.Fatal("minimizing lost color")
	}
	pruneWindowColors(taskbarSnapshot{Windows: []taskbarEntry{{ID: 11}}})
	if windowContextColors[10] != "" {
		t.Fatal("closed window retained color")
	}
	pruneWindowColors(taskbarSnapshot{Windows: []taskbarEntry{{ID: 11}, {ID: 12}}})
	if windowContextColors[12] != "" {
		t.Fatal("reopened app inherited color")
	}
}
