package main

import (
	"context"
	"github.com/godbus/dbus/v5"
	"strings"
	"time"
)

type mediaSnapshot struct {
 Album, Loop string
 Shuffle, HasShuffle, HasLoop, HasVolume bool
 Volume float64
 ArtURL string
	Bus, Title, Artist                     string
	Playing, Control, Next, Previous, Seek bool
	Position, Length                       int64
	Track                                  dbus.ObjectPath
}

func readSpotifyMedia() mediaSnapshot {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return mediaSnapshot{}
	}
	defer conn.Close()
	var names []string
	if conn.BusObject().CallWithContext(ctx, "org.freedesktop.DBus.ListNames", 0).Store(&names) != nil {
		return mediaSnapshot{}
	}
	for _, name := range names {
		if !strings.HasPrefix(name, "org.mpris.MediaPlayer2.spotify") {
			continue
		}
		var props map[string]dbus.Variant
		if conn.Object(name, "/org/mpris/MediaPlayer2").CallWithContext(ctx, "org.freedesktop.DBus.Properties.GetAll", 0, "org.mpris.MediaPlayer2.Player").Store(&props) != nil {
			continue
		}
		return decodeSpotifyMedia(name, props)
	}
	return mediaSnapshot{}
}
func decodeSpotifyMedia(name string, p map[string]dbus.Variant) mediaSnapshot {
	s := mediaSnapshot{Bus: name, Title: "Noch keine Wiedergabe", Artist: "Spotify"}
 s.Shuffle,s.HasShuffle=p["Shuffle"].Value().(bool);s.Loop,s.HasLoop=p["LoopStatus"].Value().(string);s.Volume,s.HasVolume=p["Volume"].Value().(float64)
	s.Playing = p["PlaybackStatus"].Value() == "Playing"
	s.Control, _ = p["CanControl"].Value().(bool)
	s.Next, _ = p["CanGoNext"].Value().(bool)
	s.Previous, _ = p["CanGoPrevious"].Value().(bool)
	s.Seek, _ = p["CanSeek"].Value().(bool)
	s.Position, _ = p["Position"].Value().(int64)
	if m, ok := p["Metadata"].Value().(map[string]dbus.Variant); ok {
 s.Album,_=m["xesam:album"].Value().(string)
 s.ArtURL,_=m["mpris:artUrl"].Value().(string)
		if title, ok := m["xesam:title"].Value().(string); ok && title != "" {
			s.Title = title
		}
		if artists, ok := m["xesam:artist"].Value().([]string); ok && len(artists) > 0 {
			s.Artist = strings.Join(artists, ", ")
		}
		s.Length, _ = m["mpris:length"].Value().(int64)
		if n, ok := m["mpris:length"].Value().(uint64); ok && n <= uint64(^uint64(0)>>1) {
			s.Length = int64(n)
		}
		s.Track, _ = m["mpris:trackid"].Value().(dbus.ObjectPath)
	}
	return s
}
func spotifyMediaCommand(s mediaSnapshot, method string, args ...interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	c, e := dbus.ConnectSessionBus()
	if e != nil {
		return e
	}
	defer c.Close()
	return c.Object(s.Bus, "/org/mpris/MediaPlayer2").CallWithContext(ctx, "org.mpris.MediaPlayer2.Player."+method, 0, args...).Err
}
