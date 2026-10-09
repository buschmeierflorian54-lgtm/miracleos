package main
import (
 "github.com/dlasky/gotk3-layershell/layershell"
 "github.com/gotk3/gotk3/gdk"
 "github.com/gotk3/gotk3/gtk"
)
func createFreeStart()*gtk.Window {
 win,_:=gtk.WindowNew(gtk.WINDOW_TOPLEVEL);activeMainMenu=win;win.SetTitle("Start — MiracleOS");win.SetDecorated(false);win.SetResizable(false)
 win.SetAppPaintable(true);if screen:=win.GetScreen();screen!=nil{if v,e:=screen.GetRGBAVisual();e==nil{win.SetVisual(v)}}
 layershell.InitForWindow(win);layershell.SetNamespace(win,"miracleos-start");layershell.SetLayer(win,layershell.LAYER_SHELL_LAYER_OVERLAY)
 layershell.SetKeyboardMode(win,layershell.LAYER_SHELL_KEYBOARD_MODE_EXCLUSIVE);layershell.SetAnchor(win,layershell.LAYER_SHELL_EDGE_BOTTOM,true);layershell.SetMargin(win,layershell.LAYER_SHELL_EDGE_BOTTOM,72)
 dismissPopupOutside(win,func(){win.Hide()})
 root,_:=gtk.BoxNew(gtk.ORIENTATION_VERTICAL,14);menuClass(root,"free-start")
 root.SetSizeRequest(990,610)
 header,_:=gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL,14);header.PackStart(menuLabel("Dein Start","free-start-title"),true,true,0)
 search,_:=gtk.EntryNew();search.SetPlaceholderText("Apps durchsuchen …");menuClass(search,"settings-search");search.SetSizeRequest(320,40);header.PackStart(search,false,false,0)
 header.PackEnd(settingsButton("Einstellungen",launchSettings),false,false,0);root.PackStart(header,false,false,0)
 body,_:=gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL,18)
 apps,_:=gtk.BoxNew(gtk.ORIENTATION_VERTICAL,10);apps.SetSizeRequest(250,-1);menuClass(apps,"free-start-apps")
 apps.PackStart(menuLabel("Alle Apps","settings-card-title"),false,false,0);apps.PackStart(installedAppsView(win,search),true,true,0)
 body.PackStart(apps,false,false,0)
 canvas:=newWidgetCanvas(win,"start");canvas.root.SetSizeRequest(650,440)
 scroll,_:=gtk.ScrolledWindowNew(nil,nil);scroll.SetPolicy(gtk.POLICY_AUTOMATIC,gtk.POLICY_AUTOMATIC);scroll.Add(canvas.root)
 body.PackStart(scroll,true,true,0);root.PackStart(body,true,true,0)
 footer,_:=gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL,12)
 footer.PackStart(menuLabel("Rechtsklick: hinzufügen · Am Titel ziehen · Ecke ziehen: Größe","settings-muted"),true,true,0)
 footer.PackEnd(createPowerButtons(),false,false,0);root.PackEnd(footer,false,false,0)
 win.Add(root);win.Connect("key-press-event",func(_ *gtk.Window,e *gdk.Event)bool{if gdk.EventKeyNewFromEvent(e).KeyVal()==gdk.KEY_Escape{win.Hide();return true};return false})
 return win
}