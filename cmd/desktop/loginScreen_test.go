package main

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"testing"
)

func TestLoginExchange(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	go func() {
		defer server.Close()
		var size uint32
		binary.Read(server, binary.NativeEndian, &size)
		data := make([]byte, size)
		io.ReadFull(server, data)
		var request map[string]interface{}
		json.Unmarshal(data, &request)
		if request["type"] != "create_session" {
			return
		}
		reply := []byte(`{"type":"auth_message","auth_message_type":"secret","auth_message":"Password:"}`)
		binary.Write(server, binary.NativeEndian, uint32(len(reply)))
		server.Write(reply)
	}()
	reply, err := loginExchange(client, map[string]interface{}{"type": "create_session", "username": "test"})
	if err != nil || reply.Type != "auth_message" || reply.Kind != "secret" {
		t.Fatalf("%+v %v", reply, err)
	}
}
