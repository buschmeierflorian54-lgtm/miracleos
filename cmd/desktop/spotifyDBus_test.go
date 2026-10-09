package main

import (
 "os"
 "sync/atomic"
 "testing"
 "github.com/godbus/dbus/v5"
 "github.com/godbus/dbus/v5/prop"
)
type testSpotifyPlayer struct{calls atomic.Int32}
func(p *testSpotifyPlayer)PlayPause()*dbus.Error{p.calls.Add(1);return nil}
func TestSpotifyDBus(t *testing.T){
 if os.Getenv("MIRACLEOS_DBUS_TEST")!="1"{t.Skip("run under an isolated dbus-run-session")}
 conn,err:=dbus.ConnectSessionBus();if err!=nil{t.Fatal(err)};defer conn.Close()
 const bus="org.mpris.MediaPlayer2.spotify.codex_test";const path=dbus.ObjectPath("/org/mpris/MediaPlayer2")
 result,err:=conn.RequestName(bus,dbus.NameFlagDoNotQueue);if err!=nil||result!=dbus.RequestNameReplyPrimaryOwner{t.Fatal("mock player bus unavailable",err)}
 iface:="org.mpris.MediaPlayer2.Player"
 player:=&testSpotifyPlayer{};if err:=conn.Export(player,path,iface);err!=nil{t.Fatal(err)}
 props:=map[string]*prop.Prop{}
 for key,value:=range map[string]interface{}{
  "PlaybackStatus":"Playing","CanControl":true,"CanGoNext":true,"CanGoPrevious":true,"CanSeek":true,
  "Position":int64(42000000),"Shuffle":false,"LoopStatus":"None","Volume":.6,
  "Metadata":map[string]dbus.Variant{"xesam:title":dbus.MakeVariant("Test track"),"xesam:artist":dbus.MakeVariant([]string{"Test artist"}),"mpris:length":dbus.MakeVariant(int64(210000000))},
 } {writable:=key=="Volume"||key=="Shuffle"||key=="LoopStatus";props[key]=&prop.Prop{Value:value,Writable:writable,Emit:prop.EmitTrue}}
 if _,err:=prop.Export(conn,path,map[string]map[string]*prop.Prop{iface:props});err!=nil{t.Fatal(err)}
 snap:=readSpotifyMedia();if snap.Bus!=bus||snap.Title!="Test track"||!snap.Playing||snap.Position!=42000000{t.Fatalf("unexpected DBus snapshot: %+v",snap)}
 if err:=spotifyMediaCommand(snap,"PlayPause");err!=nil{t.Fatal(err)};if player.calls.Load()!=1{t.Fatal("transport command not delivered")}
 if err:=spotifySetProperty(snap,"Volume",.2);err!=nil{t.Fatal(err)}
 if err:=spotifySetProperty(snap,"Shuffle",true);err!=nil{t.Fatal(err)}
 if err:=spotifySetProperty(snap,"LoopStatus","Track");err!=nil{t.Fatal(err)}
 snap=readSpotifyMedia();if snap.Volume!=.2||!snap.Shuffle||snap.Loop!="Track"{t.Fatalf("controls did not update player: %+v",snap)}
}
