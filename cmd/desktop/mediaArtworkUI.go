package main
import (
 "github.com/gotk3/gotk3/gdk"
 "github.com/gotk3/gotk3/glib"
 "github.com/gotk3/gotk3/gtk"
)
// The returned updater runs on GTK; downloading never blocks the UI.
func mediaArtworkUpdater(box *gtk.Box,size int)func(string){
 alive:=true;current:="";box.SetSizeRequest(size,size);box.SetHAlign(gtk.ALIGN_CENTER);box.SetVAlign(gtk.ALIGN_CENTER);box.SetHExpand(false);box.SetVExpand(false);box.Connect("destroy",func(){alive=false})
 return func(url string){if url==current{return};current=url
  box.GetChildren().Foreach(func(v interface{}){v.(gtk.IWidget).ToWidget().Destroy()});box.PackStart(quickIcon("music","#ffd9ec",22),true,true,0);box.ShowAll()
  if url==""{return};go func(){data,err:=readMusicArtwork(url);if err!=nil{return};glib.IdleAdd(func(){if !alive||current!=url{return};pix,e:=gdk.PixbufNewFromBytesOnly(data);if e!=nil{return};scaled,e:=pix.ScaleSimple(size,size,gdk.INTERP_BILINEAR);if e!=nil{return};img,_:=gtk.ImageNewFromPixbuf(scaled);box.GetChildren().Foreach(func(v interface{}){v.(gtk.IWidget).ToWidget().Destroy()});box.PackStart(img,true,true,0);box.ShowAll()})}()
 }
}
