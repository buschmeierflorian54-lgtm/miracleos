package main

import (
	"github.com/godbus/dbus/v5"
	"testing"
)

func TestSpotifyOverlayProperties(t *testing.T) {
	s := decodeSpotifyMedia("org.mpris.MediaPlayer2.spotify", map[string]dbus.Variant{
		"Shuffle": dbus.MakeVariant(true), "LoopStatus": dbus.MakeVariant("Track"), "Volume": dbus.MakeVariant(.6),
		"Metadata": dbus.MakeVariant(map[string]dbus.Variant{"xesam:album": dbus.MakeVariant("Album")}),
	})
	if !s.Shuffle || !s.HasShuffle || !s.HasLoop || !s.HasVolume || s.Volume != .6 || s.Loop != "Track" || s.Album != "Album" {
		t.Fatalf("properties lost: %+v", s)
	}
	missing := decodeSpotifyMedia("spotify", map[string]dbus.Variant{})
	if missing.HasVolume || missing.HasShuffle || missing.HasLoop {
		t.Fatal("unsupported controls enabled")
	}
}
func TestMediaTime(t *testing.T) {
	for _, tc := range []struct {
		us   int64
		want string
	}{{-1, "0:00"}, {0, "0:00"}, {42000000, "0:42"}, {210000000, "3:30"}} {
		if got := mediaTime(tc.us); got != tc.want {
			t.Fatalf("%d: %s", tc.us, got)
		}
	}
}
