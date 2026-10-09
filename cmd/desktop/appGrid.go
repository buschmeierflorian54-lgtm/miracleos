package main

import (
 "encoding/json"
 "sort"
 "strings"
 "time"
 "github.com/gotk3/gotk3/gdk"
 "github.com/gotk3/gotk3/glib"
 "github.com/gotk3/gotk3/gtk"
)

func gridItems(p desktopLayout,id string) []canvasItem {
 items:=[]canvasItem{};mode:="manual"
 for _,v:=range p.Items {if v.ID==id {mode=v.Sort};if v.Surface=="frame:"+id&&v.Kind=="app" {items=append(items,v)}}
 sort.SliceStable(items,func(i,j int)bool{switch mode {case "alphabet":return strings.ToLower(items[i].App.Name)<strings.ToLower(items[j].App.Name);case "added":return items[i].AddedAt>items[j].AddedAt};return items[i].Order<items[j].Order})
 return items
}
func gridSortActions(id string) []contextAction {
 actions:=[]contextAction{}
 for _,choice:=range []struct{key,label string}{{"manual","Anordnung: selbst verschieben"},{"alphabet","Anordnung: alphabetisch"},{"added","Anordnung: zuletzt hinzugefügt zuerst"}} {choice:=choice;actions=append(actions,contextAction{choice.label,func(){_ = changeDesktopLayout(func(p *desktopLayout){for i:=range p.Items {if p.Items[i].ID==id {p.Items[i].Sort=choice.key}}})},false})}
 return actions
}
func insertGridApp(p *desktopLayout,id string,item canvasItem,before string) bool {
 found:=false;for _,v:=range p.Items {if v.ID==id&&v.Kind=="app-grid"{found=true}};if !found||item.Kind!="app" {return false}
 if item.Surface!="frame:"+id||item.AddedAt==0 {item.AddedAt=time.Now().UnixNano()}
 item.Surface="frame:"+id
 items:=gridItems(*p,id);out:=[]canvasItem{};inserted:=false
 for _,v:=range items {if v.ID==item.ID{continue};if v.ID==before {out=append(out,item);inserted=true};out=append(out,v)}
 if !inserted {out=append(out,item)}
 for order,v:=range out {v.Order=order;exists:=false;for i:=range p.Items {if p.Items[i].ID==v.ID {p.Items[i]=v;exists=true;break}};if !exists {p.Items=append(p.Items,v)}}
 if before!="" {for i:=range p.Items {if p.Items[i].ID==id {p.Items[i].Sort="manual"}}}
 return true
}
func addAppGrid(shell *gtk.Box,owner *gtk.Window,item canvasItem) {
 edit,_:=gtk.ButtonNewWithLabel("✎");edit.SetHAlign(gtk.ALIGN_END);edit.SetTooltipText("Anordnung ändern");edit.Connect("clicked",func(){glassContextMenu(owner,nil,gridSortActions(item.ID))});shell.PackStart(edit,false,false,0)
 scroll,_:=gtk.ScrolledWindowNew(nil,nil);scroll.SetPolicy(gtk.POLICY_NEVER,gtk.POLICY_AUTOMATIC);scroll.SetMinContentHeight(80)
 flow,_:=gtk.FlowBoxNew();flow.SetSelectionMode(gtk.SELECTION_NONE);flow.SetMinChildrenPerLine(1);flow.SetMaxChildrenPerLine(20);flow.SetRowSpacing(8);flow.SetColumnSpacing(8);flow.SetVAlign(gtk.ALIGN_START);scroll.Add(flow);shell.PackStart(scroll,true,true,0)
 alive,busy:=true,false;last:="";shell.Connect("destroy",func(){alive=false})
 drop:=func(payload layoutDrag,before string)bool {
  var app canvasItem
  if payload.ItemID!="" {for _,v:=range loadDesktopLayout().Items {if v.ID==payload.ItemID {app=v;break}}} else if a,ok:=draggedApp(payload);ok {app=newCanvasItem("app","",0,0);app.App=a}
  ok:=false;err:=changeDesktopLayout(func(p *desktopLayout){ok=insertGridApp(p,item.ID,app,before)});return err==nil&&ok
 }
 layoutDropTarget(flow,func(p layoutDrag,x,y int)bool{return drop(p,"")})
 refresh:=func()bool {
  if !alive{return false};if busy{return true};items:=gridItems(loadDesktopLayout(),item.ID);data,_:=json.Marshal(items);if string(data)==last{return true};last=string(data)
  flow.GetChildren().Foreach(func(v interface{}){v.(gtk.IWidget).ToWidget().Destroy()})
  for _,app:=range items {app:=app;button,_:=gtk.ButtonNew();menuClass(button,"canvas-app-icon");button.SetSizeRequest(48,48);button.SetTooltipText(app.App.Name);button.Add(installedAppIcon(app.App));button.Connect("clicked",func(){launchInstalledApp(app.App)})
   layoutDragSource(button,layoutDrag{ItemID:app.ID});button.Connect("drag-begin",func(){busy=true});button.Connect("drag-end",func(){busy=false});layoutDropTarget(button,func(p layoutDrag,x,y int)bool{return drop(p,app.ID)})
   button.Connect("button-press-event",func(_ interface{},event *gdk.Event)bool {if gdk.EventButtonNewFromEvent(event).Button()!=3{return false};actions:=append(appContextActions(app.App),contextAction{"Aus dem Raster entfernen",func(){_ = removeCanvasItem(app.ID)},false});actions=append(actions,gridSortActions(item.ID)...);actions=append(actions,contextAction{"Raster auflösen · Apps behalten",func(){_ = removeCanvasItem(item.ID)},false});glassContextMenu(owner,event,actions);return true})
   flow.Add(button)
  };flow.ShowAll();return true
 };refresh();glib.TimeoutAdd(350,refresh)
}

// Attach before default handlers so text views and controls also offer widget management.
func bindWidgetManagement(w gtk.IWidget,handler func(interface{},*gdk.Event)bool) {
 widget:=w.ToWidget();widget.AddEvents(int(gdk.BUTTON_PRESS_MASK));widget.Connect("button-press-event",handler)
 if container,ok:=w.(interface{GetChildren() *glib.List});ok {container.GetChildren().Foreach(func(v interface{}){if child,ok:=v.(gtk.IWidget);ok {bindWidgetManagement(child,handler)}})}
}
