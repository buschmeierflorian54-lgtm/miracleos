package main

import (
	"encoding/json"
	"os/exec"
	"testing"
)

func TestTerminalQuotedPaths(t *testing.T) {
	paths := []string{"/tmp/a b", "/tmp/a'b", "/tmp/$(touch should-not-run)", "/tmp/a\nb"}
	command := "python3 -c 'import sys,json;print(json.dumps(sys.argv[1:]))' " + terminalQuotedPaths(paths)
	out, err := exec.Command("sh", "-c", command).Output()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	if err = json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(paths) {
		t.Fatal(got)
	}
	for i := range paths {
		if paths[i] != got[i] {
			t.Fatalf("round trip mismatch %q %q", paths[i], got[i])
		}
	}
}
