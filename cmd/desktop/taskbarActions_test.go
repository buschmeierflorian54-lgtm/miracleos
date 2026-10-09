package main

import "testing"

func TestTaskbarDistinctWindows(t *testing.T) {
	tree := desktopNode{Nodes: []desktopNode{{ID: 10, PID: 100, Name: "Same", AppID: "app"}, {ID: 11, PID: 100, Name: "Same", AppID: "app"}, {Name: "__i3_scratch", FloatingNodes: []desktopNode{{ID: 12, PID: 101, Name: "Hidden", AppID: "app"}}}}}
	snapshot, err := readTaskbarSnapshot(tree, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Windows) != 2 || snapshot.Windows[0].ID != 10 || snapshot.Windows[1].ID != 11 {
		t.Fatalf("incorrect window identities: %+v", snapshot.Windows)
	}
	if len(snapshot.Minimized) != 1 || snapshot.Minimized[0].ID != 12 {
		t.Fatal("minimized window missing or duplicated")
	}
}
