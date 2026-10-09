package main

import (
	"encoding/json"
	"net"
	"testing"
)

func TestWindowDropAcknowledgement(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	listener, err := net.Listen("unix", windowDropSocket(4242))
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	received := make(chan windowDropRequest, 1)
	go func() {
		conn, _ := listener.Accept()
		defer conn.Close()
		var request windowDropRequest
		json.NewDecoder(conn).Decode(&request)
		received <- request
		json.NewEncoder(conn).Encode(windowDropReply{OK: true})
	}()
	if err = sendWindowDrop(4242, windowDropRequest{Paths: []string{"/tmp/example.png"}, Move: true}); err != nil {
		t.Fatal(err)
	}
	request := <-received
	if len(request.Paths) != 1 || !request.Move {
		t.Fatal("drop data changed")
	}
	go func() {
		conn, _ := listener.Accept()
		defer conn.Close()
		var request windowDropRequest
		json.NewDecoder(conn).Decode(&request)
		json.NewEncoder(conn).Encode(windowDropReply{OK: false})
	}()
	if sendWindowDrop(4242, windowDropRequest{Paths: []string{"/tmp/rejected.png"}}) == nil {
		t.Fatal("rejected transfer reported success")
	}
}
