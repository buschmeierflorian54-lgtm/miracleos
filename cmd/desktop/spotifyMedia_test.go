package main

import (
	"github.com/godbus/dbus/v5"
	"testing"
)

func TestSpotifyMetadataAndCapabilities(t *testing.T) {
	s := decodeSpotifyMedia("org.mpris.MediaPlayer2.spotify", map[string]dbus.Variant{"PlaybackStatus": dbus.MakeVariant("Playing"), "CanControl": dbus.MakeVariant(true), "CanSeek": dbus.MakeVariant(false), "Metadata": dbus.MakeVariant(map[string]dbus.Variant{"xesam:title": dbus.MakeVariant("Song"), "xesam:artist": dbus.MakeVariant([]string{"Artist", "Guest"}), "mpris:length": dbus.MakeVariant(int64(123000000)), "mpris:trackid": dbus.MakeVariant(dbus.ObjectPath("/track/1"))})})
	if s.Title != "Song" || s.Artist != "Artist, Guest" || !s.Playing || !s.Control || s.Seek || s.Length != 123000000 || s.Track != "/track/1" {
		t.Fatalf("unexpected player snapshot: %+v", s)
	}
}
func TestSpotifyMissingMetadata(t *testing.T) {
	s := decodeSpotifyMedia("spotify", map[string]dbus.Variant{})
	if s.Title != "Noch keine Wiedergabe" || s.Control || s.Seek || s.Length != 0 {
		t.Fatalf("invalid missing-metadata state: %+v", s)
	}
}
func TestSpotifyUnsignedLength(t *testing.T) {
	s := decodeSpotifyMedia("spotify", map[string]dbus.Variant{"Metadata": dbus.MakeVariant(map[string]dbus.Variant{"mpris:length": dbus.MakeVariant(uint64(171996000))})})
	if s.Length != 171996000 {
		t.Fatalf("Spotify unsigned length not decoded: %d", s.Length)
	}
}
