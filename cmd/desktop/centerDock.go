package main
import (
 "encoding/json"
 "strings"
 "time"
 "github.com/MiracleOS-Team/libxdg-go/desktopFiles"
 "github.com/gotk3/gotk3/gdk"
 "github.com/gotk3/gotk3/glib"
 "github.com/gotk3/gotk3/gtk"
)
var centerDockWidth int
func dockAppMatches(app installedApp,id,title string)bool {
 if app.Name=="Dateien" && strings.HasPrefix(title,"Dateien —"){return true}
 if app.Name=="Einstellungen" && strings.HasPrefix(title,"Einstellungen —"){return true}
 base:=strings.TrimSuffix(app.ID,".desktop")
 return strings.EqualFold(base,id) || app.WMClass!=""&&strings.EqualFold(app.WMClass,id)
}
func createCenterDock(apps []desktopFiles.DesktopFile)*gtk.Box {
 dock,_:=gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL,6);dock.SetHAlign(gtk.ALIGN_CENTER);menuClass(dock,"center-dock")
 start:=createMainMenu(apps);toggle:=settingsButton("",func(){if start.IsVisible(){start.Hide()}else{start.ShowAll()}});menuClass(toggle,"dock-start");toggle.SetTooltipText("Start öffnen")
 quickIconPaths["apps-grid"] = `<g fill="currentColor" stroke="none"><circle cx="5" cy="5" r="2"/><circle cx="12" cy="5" r="2"/><circle cx="19" cy="5" r="2"/><circle cx="5" cy="12" r="2"/><circle cx="12" cy="12" r="2"/><circle cx="19" cy="12" r="2"/><circle cx="5" cy="19" r="2"/><circle cx="12" cy="19" r="2"/><circle cx="19" cy="19" r="2"/></g>`
 if child,err:=toggle.GetChild();err==nil{child.ToWidget().Destroy()}
 toggle.Add(quickIcon("apps-grid",quickTextColor(),22))
 dock.PackStart(toggle,false,false,0)
 area,_:=gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL,6);dock.PackStart(area,false,false,0)
 hint:=menuLabel("＋ Anheften","dock-drop-hint");hint.SetNoShowAll(true);dock.PackEnd(hint,false,false,0)
 page:=0;snapshot:=taskbarSnapshot{};last:="";alive:=true;dragging:=false
 dock.Connect("destroy",func(){alive=false})
 var render func()
 render=func(){
  area.GetChildren().Foreach(func(v interface{}){v.(gtk.IWidget).ToWidget().Destroy()})
  pruneWindowColors(snapshot)
  type entry struct{app installedApp;id int64;title,icon,appID string;minimized bool}
  entries:=[]entry{};used:=map[int64]bool{}
  for _,app:=range loadDesktopLayout().Pins {
   e:=entry{app:app,title:app.Name,icon:app.Icon}
   for _,w:=range snapshot.Windows{if dockAppMatches(app,w.Window.AppID,w.Window.Title){e.id=w.ID;e.title=w.Window.Title;e.icon=w.IconPath;e.appID=w.Window.AppID;used[w.ID]=true;break}}
   if e.id==0{for _,w:=range snapshot.Minimized{if dockAppMatches(app,w.AppID,w.Name){e.id=w.ID;e.title=w.Name;e.icon=snapshot.MinimizedIcons[w.ID];e.minimized=true;used[w.ID]=true;break}}}
   entries=append(entries,e)
  }
  for _,w:=range snapshot.Windows{if !used[w.ID]{entries=append(entries,entry{id:w.ID,title:w.Window.Title,icon:w.IconPath,appID:w.Window.AppID})}}
  for _,w:=range snapshot.Minimized{if !used[w.ID]{entries=append(entries,entry{id:w.ID,title:w.Name,icon:snapshot.MinimizedIcons[w.ID],appID:w.AppID,minimized:true})}}
  pages:=max(1,(len(entries)+4)/5);if page>=pages{page=pages-1}
  if pages>1 {b:=settingsButton("‹",func(){if page>0{page--;render()}});b.SetSensitive(page>0);area.PackStart(b,false,false,0)}
  for _,e:=range entries[page*5:min(len(entries),(page+1)*5)] {
   e:=e;var b *gtk.Button
   if e.id!=0 {b=taskbarWindowButton(e.id,e.appID,e.title,e.icon,e.minimized)}else{b,_=gtk.ButtonNew();menuClass(b,"app");b.Add(installedAppIcon(e.app));b.Connect("clicked",func(){launchInstalledApp(e.app)});attachAppContext(b,start,e.app)}
   if e.app.ID!=""{b.SetTooltipText(e.title+" · angeheftet");layoutDragSource(b,layoutDrag{AppID:e.app.ID});menuClass(b,"dock-pinned")}
   area.PackStart(b,false,false,0)
  }
  if pages>1 {b:=settingsButton("›",func(){if page+1<pages{page++;render()}});b.SetSensitive(page+1<pages);area.PackStart(b,false,false,0)}
  area.ShowAll()
 }
 layoutDropTarget(dock,func(payload layoutDrag,_ int,_ int)bool{app,ok:=draggedApp(payload);if !ok{return false};err:=pinApp(app,true);dragging=false;hint.Hide();glib.IdleAdd(render);return err==nil})
 dock.Connect("drag-motion",func()bool{dragging=true;hint.Show();return false})
 dock.Connect("drag-leave",func(){dragging=false;hint.Hide()})
 target,_:=gtk.TargetEntryNew("application/x-miracleos-layout",0,0);toggle.DragDestSet(0,[]gtk.TargetEntry{*target},gdk.ACTION_COPY|gdk.ACTION_MOVE)
 hovering:=false;toggle.Connect("drag-motion",func()bool{if !hovering{hovering=true;glib.TimeoutAdd(400,func()bool{if hovering&&alive{start.ShowAll()};return false})};return false});toggle.Connect("drag-leave",func(){hovering=false})
 render()
 glib.TimeoutAdd(200,func()bool{if !alive{return false};centerDockWidth=dock.GetAllocatedWidth();return true})
 go func(){for range time.NewTicker(time.Second).C{
  tree,e:=desktopTree();if e!=nil{continue};s,e:=readTaskbarSnapshot(tree,apps);if e!=nil{continue};pins,_:=json.Marshal(loadDesktopLayout().Pins);signature:=windowSignature(tree)+string(pins)
  glib.IdleAdd(func(){if !alive{return};if !dragging&&signature!=last{snapshot=s;last=signature;render()}})
 }}()
 return dock
}
