package main
import (
 "github.com/gotk3/gotk3/glib"
 "github.com/gotk3/gotk3/gdk"
 "github.com/gotk3/gotk3/gtk"
 "github.com/gotk3/gotk3/pango"
)
func addCanvasMusic(shell *gtk.Box){
 menuClass(shell,"canvas-music")
 alive:=true;shell.Connect("destroy",func(){alive=false})
 top,_:=gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL,10)
 art,_:=gtk.BoxNew(gtk.ORIENTATION_VERTICAL,0);menuClass(art,"canvas-music-art");art.SetSizeRequest(64,64);art.SetHAlign(gtk.ALIGN_CENTER);art.SetVAlign(gtk.ALIGN_CENTER);art.SetHExpand(false);art.SetVExpand(false);art.PackStart(quickIcon("music","#ffd9ec",26),true,true,0)
 top.PackStart(art,false,false,0)
 text,_:=gtk.BoxNew(gtk.ORIENTATION_VERTICAL,3);text.SetHExpand(true)
 title:=menuLabel("Keine Wiedergabe","canvas-track");title.SetXAlign(0);title.SetEllipsize(pango.ELLIPSIZE_END);title.SetMaxWidthChars(22)
 artist:=menuLabel("Spotify öffnen","canvas-artist");artist.SetXAlign(0);artist.SetEllipsize(pango.ELLIPSIZE_END);artist.SetMaxWidthChars(24)
 text.PackStart(title,true,true,0);text.PackStart(artist,true,true,0);top.PackStart(text,true,true,0);shell.PackStart(top,true,true,0)
 progress,_:=gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL,0,100,1);progress.SetDrawValue(false);menuClass(progress,"canvas-music-progress");shell.PackStart(progress,false,false,0)
 controls,_:=gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL,18);controls.SetHAlign(gtk.ALIGN_CENTER)
 var media mediaSnapshot;updating:=false;busy:=false
 makeButton:=func(icon,method,hint string)*gtk.Button{
  b,_:=gtk.ButtonNew();menuClass(b,"canvas-player-button");b.Add(quickIcon(icon,"#ffd9ec",19));b.SetTooltipText(hint);b.SetSensitive(false)
  b.Connect("clicked",func(){if busy||!media.Control{return};busy=true;snapshot:=media;go func(){err:=spotifyMediaCommand(snapshot,method);glib.IdleAdd(func(){busy=false;if alive&&err!=nil{artist.SetText("Wiedergabe nicht erreichbar")}})}()});controls.PackStart(b,false,false,0);return b
 }
 previous:=makeButton("previous","Previous","Vorheriger Titel")
 play:=makeButton("play","PlayPause","Wiedergabe / Pause");menuClass(play,"canvas-player-main")
 next:=makeButton("next","Next","Nächster Titel")
 shell.PackStart(controls,false,false,0)
 progress.Connect("change-value",func()bool{if !updating&&media.Seek&&media.Length>0{snapshot:=media;position:=int64(progress.GetValue()*float64(media.Length)/100);go spotifyMediaCommand(snapshot,"SetPosition",snapshot.Track,position)};return false})
 fetching:=false;lastPlaying:=false;hasIcon:=false;artURL:=""
 refresh:=func()bool{
  if !alive{return false};if fetching{return true};fetching=true
  go func(){snap:=readSpotifyMedia();glib.IdleAdd(func(){fetching=false;if !alive{return};media=snap
   if snap.ArtURL!=artURL {
    artURL=snap.ArtURL;requested:=artURL
    art.GetChildren().Foreach(func(v interface{}){v.(gtk.IWidget).ToWidget().Destroy()})
    art.PackStart(quickIcon("music","#ffd9ec",26),true,true,0);art.ShowAll()
    if requested!=""{go func(){bytes,e:=readMusicArtwork(requested);if e!=nil{return};glib.IdleAdd(func(){
     if !alive||artURL!=requested{return};pixbuf,e:=gdk.PixbufNewFromBytesOnly(bytes);if e!=nil{return}
     scaled,e:=pixbuf.ScaleSimple(64,64,gdk.INTERP_BILINEAR);if e!=nil{return};image,_:=gtk.ImageNewFromPixbuf(scaled)
     art.GetChildren().Foreach(func(v interface{}){v.(gtk.IWidget).ToWidget().Destroy()});art.PackStart(image,true,true,0);art.ShowAll()
    })}()}
   }
   if snap.Bus!=""{title.SetText(snap.Title);artist.SetText(snap.Artist)}else{title.SetText("Keine Wiedergabe");artist.SetText("Spotify öffnen")}
   previous.SetSensitive(snap.Control&&snap.Previous);play.SetSensitive(snap.Control);next.SetSensitive(snap.Control&&snap.Next);progress.SetSensitive(snap.Seek&&snap.Length>0)
   if !hasIcon||lastPlaying!=snap.Playing{if child,e:=play.GetChild();e==nil{child.ToWidget().Destroy()};icon:="play";if snap.Playing{icon="pause"};play.Add(quickIcon(icon,"#3a2233",21));play.ShowAll();lastPlaying=snap.Playing;hasIcon=true}
   updating=true;value:=0.0;if snap.Length>0{value=float64(snap.Position)*100/float64(snap.Length)};progress.SetValue(value);updating=false
  })}();return true
 };refresh();glib.TimeoutAdd(1000,refresh)
}