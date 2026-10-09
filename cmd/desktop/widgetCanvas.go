package main

import (
 "encoding/json"
 "fmt"

 "github.com/gotk3/gotk3/gdk"
 "github.com/gotk3/gotk3/glib"
 "github.com/gotk3/gotk3/gtk"
)
type widgetCanvas struct {root *gtk.EventBox;fixed *gtk.Fixed;owner *gtk.Window;surface string;last string;alive,busy bool}
func newWidgetCanvas(owner *gtk.Window,surface string)*widgetCanvas {
 c:=&widgetCanvas{owner:owner,surface:surface,alive:true}
 c.root,_=gtk.EventBoxNew();c.root.SetVisibleWindow(false);c.root.SetHExpand(true);c.root.SetVExpand(true)
 c.fixed,_=gtk.FixedNew();c.fixed.SetHExpand(true);c.fixed.SetVExpand(true);c.root.Add(c.fixed)
 c.root.Connect("destroy",func(){c.alive=false})
 c.root.AddEvents(int(gdk.BUTTON_PRESS_MASK))
 c.root.Connect("button-press-event",func(_ interface{},event *gdk.Event)bool{
  e:=gdk.EventButtonNewFromEvent(event);if e.Button()!=3{return false}
  x,y:=int(e.X()),int(e.Y());actions:=[]contextAction{}
  for _,kind:=range []string{"frame","app-grid","clock","clock-compact","calendar","notes","note","music"}{kind:=kind;actions=append(actions,contextAction{"Widget hinzufügen: "+widgetTitle(kind),func(){_ = placeCanvasItem(c.fit(newCanvasItem(kind,surface,x,y)));c.refresh()},false})}
  actions=append(actions,contextAction{"",nil,false},contextAction{"Einstellungen",launchSettings,false})
  glassContextMenu(owner,event,actions);return true
 })
 layoutDropTarget(c.root,func(p layoutDrag,x,y int)bool{
  var item canvasItem;found:=false
  if p.ItemID!="" {for _,v:=range loadDesktopLayout().Items {if v.ID==p.ItemID {item=v;found=true;break}}} else if app,ok:=draggedApp(p);ok {item=newCanvasItem("app",surface,x,y);item.App=app;item.Width=100;item.Height=105;found=true}
  if !found{return false};item.Surface=surface;item.X=max(0,x-20);item.Y=max(0,y-16)
  e:=placeCanvasItem(c.fit(item));glib.IdleAdd(func(){if c.alive{c.refresh()}});return e==nil
 })
 c.refresh();glib.TimeoutAdd(650,func()bool{if !c.alive{return false};if !c.busy{c.refresh()};return true})
 return c
}
func widgetTitle(kind string)string {switch kind{case "app-grid":return "App-Raster";case "frame":return "Leerer Rahmen";case "clock-compact":return "Uhr · kompakt";case "notes":return "Notizen";case "clock":return "Uhr";case "calendar":return "Kalender";case "note":return "Schnellnotiz";case "music":return "Musik";case "app":return "App"};return "Widget"}
func(c *widgetCanvas) refresh(){
 p:=loadDesktopLayout();items:=[]canvasItem{};for _,v:=range p.Items{if v.Surface==c.surface{items=append(items,c.fit(v))}}
 fingerprint:=append([]canvasItem(nil),items...);for i:=range fingerprint{fingerprint[i].Text=""};b,_:=json.Marshal(fingerprint);if string(b)==c.last{return};c.last=string(b)
 c.fixed.GetChildren().Foreach(func(v interface{}){v.(gtk.IWidget).ToWidget().Destroy()})
 for _,item:=range items{c.add(item)};c.fixed.ShowAll()
}
func(c *widgetCanvas)add(item canvasItem){
 if item.Kind=="app" {
  button,_:=gtk.ButtonNew();menuClass(button,"canvas-app-icon");button.SetSizeRequest(32,32)
  button.SetTooltipText(item.App.Name);button.Add(installedAppIcon(item.App))
  button.Connect("clicked",func(){launchInstalledApp(item.App)})
  layoutDragSource(button,layoutDrag{ItemID:item.ID})
  button.Connect("drag-begin",func(){c.busy=true});button.Connect("drag-end",func(){c.busy=false})
  button.AddEvents(int(gdk.BUTTON_PRESS_MASK));button.Connect("button-press-event",func(_ interface{},event *gdk.Event)bool{
   if gdk.EventButtonNewFromEvent(event).Button()!=3{return false}
   actions:=append(appContextActions(item.App),contextAction{"Vom Bildschirm entfernen",func(){_ = removeCanvasItem(item.ID);c.refresh()},false})
   glassContextMenu(c.owner,event,actions);return true
  })
  c.fixed.Put(button,item.X,item.Y);return
 }
 outer,_:=gtk.EventBoxNew();outer.SetVisibleWindow(false)
 shell,_:=gtk.BoxNew(gtk.ORIENTATION_VERTICAL,5)
 if item.Kind=="app" {menuClass(shell,"canvas-app")}
 viewport,_:=gtk.BoxNew(gtk.ORIENTATION_VERTICAL,5);menuClass(viewport,"canvas-widget");viewport.SetSizeRequest(item.Width,item.Height)
 contentScroll,_:=gtk.ScrolledWindowNew(nil,nil);contentScroll.SetPolicy(gtk.POLICY_AUTOMATIC,gtk.POLICY_AUTOMATIC);shell.SetHExpand(true);shell.SetVExpand(true);shell.SetHAlign(gtk.ALIGN_FILL);shell.SetVAlign(gtk.ALIGN_FILL);
 header,_:=gtk.EventBoxNew();header.SetVisibleWindow(false)
 label:=widgetTitle(item.Kind);if item.Kind=="app"{label=item.App.Name}
 heading:=menuLabel(label,"canvas-widget-title");heading.SetXAlign(0);header.Add(heading)
 viewport.PackStart(header,false,false,0)
 if item.Kind=="clock"||item.Kind=="clock-compact"{
  // Clocks draw to their allocation and do not need a scrolling viewport.
  // Packing directly keeps their center tied to the card after moves/resizes.
  contentScroll.Destroy();viewport.PackStart(shell,true,true,0)
 }else{contentScroll.Add(shell);viewport.PackStart(contentScroll,true,true,0)}
 layoutDragSource(header,layoutDrag{ItemID:item.ID},viewport)
 header.Connect("drag-begin",func(){c.busy=true;viewport.SetOpacity(0)});header.Connect("drag-end",func(){viewport.SetOpacity(1);c.busy=false})
 manageWidget:=func(_ interface{},event *gdk.Event)bool{
  if gdk.EventButtonNewFromEvent(event).Button()!=3{return false}
  removeLabel:="Vom Bildschirm entfernen";if item.Kind=="frame"||item.Kind=="app-grid"{removeLabel="Rahmen auflösen · Inhalt behalten"};actions:=[]contextAction{{removeLabel,func(){_ = removeCanvasItem(item.ID);c.refresh()},false}}
  other:="desktop";name:="Auf den Desktop verschieben";if c.surface=="desktop"{other="start";name="Ins Startmenü verschieben"}
  actions=append(actions,contextAction{name,func(){item.Surface=other;item.X=12;item.Y=12;_ = placeCanvasItem(item);c.refresh()},false})
  if item.Kind=="app"{actions=append(appContextActions(item.App),actions...)}
  if item.Kind=="app-grid"{actions=append(gridSortActions(item.ID),actions...)}
  glassContextMenu(c.owner,event,actions);return true
 }
 alive:=true;shell.Connect("destroy",func(){alive=false})
 switch item.Kind {
 case "app-grid":
  addAppGrid(shell,c.owner,item)
 case "frame":
  menuClass(shell,"canvas-frame")
  inner:=newWidgetCanvas(c.owner,"frame:"+item.ID)
  inner.root.SetSizeRequest(120,80)
  shell.PackStart(inner.root,true,true,0)
 case "app":
  button,_:=gtk.ButtonNew();button.Add(installedAppIcon(item.App));button.Connect("clicked",func(){launchInstalledApp(item.App)})
  layoutDragSource(button,layoutDrag{ItemID:item.ID});attachAppContext(button,c.owner,item.App)
  shell.PackStart(button,true,true,0)
 case "clock","clock-compact":
  if item.Kind=="clock"{menuClass(viewport,"widget-clock-large")};addDemoClock(shell,item.Kind=="clock-compact",viewport)
 case "notes":
  menuClass(viewport,"widget-notes");addDemoNotes(shell,item)
 case "calendar":
  calendar,_:=gtk.CalendarNew();shell.PackStart(calendar,true,true,0)
 case "note":
  menuClass(viewport,"widget-quick-note")
  entry,_:=gtk.TextViewNew();entry.SetWrapMode(gtk.WRAP_WORD_CHAR);buffer,_:=entry.GetBuffer();buffer.SetText(item.Text)
  scroll,_:=gtk.ScrolledWindowNew(nil,nil);scroll.SetPolicy(gtk.POLICY_NEVER,gtk.POLICY_AUTOMATIC);scroll.Add(entry);shell.PackStart(scroll,true,true,0)
  generation:=0;buffer.Connect("changed",func(){generation++;token:=generation;glib.TimeoutAdd(400,func()bool{
   if !alive||token!=generation{return false};start,end:=buffer.GetBounds();text,_:=buffer.GetText(start,end,true)
   _=changeDesktopLayout(func(p *desktopLayout){for i:=range p.Items{if p.Items[i].ID==item.ID{p.Items[i].Text=text}}});return false
  })})
 case "music":
  addCanvasMusic(shell);style,_:=shell.GetStyleContext();style.RemoveClass("canvas-music");menuClass(viewport,"canvas-music")
 }
 grip,_:=gtk.EventBoxNew();grip.SetVisibleWindow(false);grip.SetHAlign(gtk.ALIGN_END);grip.SetSizeRequest(22,18);grip.Add(menuLabel("◢","canvas-resize"));grip.SetTooltipText("Ziehen zum Vergrößern oder Verkleinern")
 viewport.PackEnd(grip,false,false,0)
 resizing:=false;var sx,sy float64;var width,height int
 grip.AddEvents(int(gdk.BUTTON_PRESS_MASK|gdk.BUTTON_RELEASE_MASK|gdk.POINTER_MOTION_MASK))
 grip.Connect("button-press-event",func(_ interface{},event *gdk.Event)bool{e:=gdk.EventButtonNewFromEvent(event);if e.Button()!=1{return false};resizing=true;c.busy=true;sx=e.XRoot();sy=e.YRoot();width=viewport.GetAllocatedWidth();height=viewport.GetAllocatedHeight();return true})
 grip.Connect("motion-notify-event",func(_ interface{},event *gdk.Event)bool{if !resizing{return false};mx,my:=layoutMotion(event);cw,ch:=c.bounds();item.Width=max(1,min(max(100,width+int(mx-sx)),cw-item.X));item.Height=max(1,min(max(100,height+int(my-sy)),ch-item.Y));item=c.fit(item);viewport.SetSizeRequest(item.Width,item.Height);c.fixed.Move(outer,item.X,item.Y);return true})
 grip.Connect("button-release-event",func()bool{if !resizing{return false};resizing=false;c.busy=false;_ = changeDesktopLayout(func(p *desktopLayout){for i:=range p.Items{if p.Items[i].ID==item.ID{p.Items[i].Width=item.Width;p.Items[i].Height=item.Height;p.Items[i].X=item.X;p.Items[i].Y=item.Y}}});return true})
 outer.Add(viewport)
 bindWidgetManagement(outer,manageWidget)
 c.fixed.Put(outer,item.X,item.Y)
 _=fmt.Sprintf
}
