package main

import (
 "encoding/json"
 "fmt"
 "os"
 "path/filepath"
 "syscall"
 "strings"
 "time"
)

type canvasItem struct {
 ID string `json:"id"`
 Kind string `json:"kind"`
 Surface string `json:"surface"`
 X int `json:"x"`; Y int `json:"y"`; Width int `json:"width"`; Height int `json:"height"`
 App installedApp `json:"app,omitempty"`
 Notes []widgetNote `json:"notes,omitempty"`
 Text string `json:"text,omitempty"`
 Sort string `json:"sort,omitempty"`
 AddedAt int64 `json:"added_at,omitempty"`
 Order int `json:"order,omitempty"`
}
type desktopLayout struct { Version int `json:"version"`; Pins []installedApp `json:"pins"`; Items []canvasItem `json:"items"` }
func layoutPath() string {return filepath.Join(filepath.Dir(preferencesPath()),"desktop-layout.json")}
func loadDesktopLayout() desktopLayout {
 p:=desktopLayout{Version:1}
 data,e:=os.ReadFile(layoutPath());if e==nil {if json.Unmarshal(data,&p)==nil{return p}}
 p.Items=[]canvasItem{{ID:"welcome-clock",Kind:"clock",Surface:"start",X:12,Y:12,Width:280,Height:130},{ID:"welcome-note",Kind:"note",Surface:"start",X:310,Y:12,Width:270,Height:200,Text:"Dein Platz.\nRechtsklick zum Hinzufügen.\nAm Kopf ziehen, unten rechts vergrößern."}}
 return p
}
// All processes edit the same layout under a lock; settings drags cannot overwrite
// desktop changes. A failed write leaves the previous layout intact.
func changeDesktopLayout(change func(*desktopLayout)) error {
 path:=layoutPath();if e:=os.MkdirAll(filepath.Dir(path),0700);e!=nil{return e}
 lock,e:=os.OpenFile(path+".lock",os.O_CREATE|os.O_RDWR,0600);if e!=nil{return e};defer lock.Close()
 if e=syscall.Flock(int(lock.Fd()),syscall.LOCK_EX);e!=nil{return e};defer syscall.Flock(int(lock.Fd()),syscall.LOCK_UN)
 p:=loadDesktopLayout();change(&p);data,e:=json.MarshalIndent(p,"","  ");if e!=nil{return e}
 f,e:=os.CreateTemp(filepath.Dir(path),".layout-*");if e!=nil{return e};tmp:=f.Name();defer os.Remove(tmp)
 if _,e=f.Write(data);e!=nil{f.Close();return e};if e=f.Sync();e!=nil{f.Close();return e};if e=f.Close();e!=nil{return e}
 return os.Rename(tmp,path)
}
func pinnedApp(id string) bool {for _,a:=range loadDesktopLayout().Pins {if a.ID==id{return true}};return false}
func pinApp(app installedApp, pin bool) error {return changeDesktopLayout(func(p *desktopLayout){
 out:=p.Pins[:0];for _,a:=range p.Pins {if a.ID!=app.ID {out=append(out,a)}};p.Pins=out
 if pin {p.Pins=append(p.Pins,app)}
})}
func newCanvasItem(kind,surface string,x,y int) canvasItem {
 item:=canvasItem{ID:fmt.Sprintf("item-%d",time.Now().UnixNano()),Kind:kind,Surface:surface,X:max(0,x),Y:max(0,y),Width:260,Height:170}
 if kind=="clock"{item.Width=330;item.Height=250}
 if kind=="frame"||kind=="app-grid" {item.Width=430;item.Height=320}
 return item
}
func validateCanvasParent(p *desktopLayout,item canvasItem) error {
 surface:=item.Surface;visited:=map[string]bool{item.ID:true}
 for strings.HasPrefix(surface,"frame:") {
  id:=strings.TrimPrefix(surface,"frame:");if visited[id]{return fmt.Errorf("Ein Rahmen kann nicht in sich selbst verschoben werden")};visited[id]=true
  found:=false;for _,parent:=range p.Items {if parent.ID==id&&(parent.Kind=="frame"||parent.Kind=="app-grid") {surface=parent.Surface;found=true;break}}
  if !found{return fmt.Errorf("Der Zielrahmen existiert nicht mehr")}
 }
 if surface!="start"&&surface!="desktop" {return fmt.Errorf("Ungültige Ablagefläche")};return nil
}
func placeCanvasItem(item canvasItem) error {
 var validation error
 err:=changeDesktopLayout(func(p *desktopLayout){
  validation=validateCanvasParent(p,item);if validation!=nil{return}
  for i,v:=range p.Items{if v.ID==item.ID{p.Items[i]=item;return}};p.Items=append(p.Items,item)
 });if validation!=nil{return validation};return err
}
// Removing a frame dissolves the group; apps, widgets and notes are preserved.
func removeCanvasItem(id string) error {return changeDesktopLayout(func(p *desktopLayout){
 var removed canvasItem;for _,v:=range p.Items{if v.ID==id{removed=v;break}}
 out:=p.Items[:0];for _,v:=range p.Items{if v.ID==id{continue};if v.Surface=="frame:"+id{v.Surface=removed.Surface;v.X+=removed.X;v.Y+=removed.Y};out=append(out,v)};p.Items=out
})}
