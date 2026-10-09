package main
import("testing";"sync")
func TestLayoutPinsAndCrossSurfaceMove(t *testing.T){
 t.Setenv("HOME",t.TempDir())
 a:=installedApp{ID:"one.desktop",Name:"One"};if e:=pinApp(a,true);e!=nil{t.Fatal(e)};_ = pinApp(a,true)
 if len(loadDesktopLayout().Pins)!=1{t.Fatal("duplicate pin")}
 item:=newCanvasItem("note","start",4,8);item.Text="Keep my text";_ = placeCanvasItem(item);item.Surface="desktop";item.Width=410;_ = placeCanvasItem(item)
 count:=0;for _,v:=range loadDesktopLayout().Items{if v.ID==item.ID{count++;if v.Surface!="desktop"||v.Text!=item.Text||v.Width!=410{t.Fatal(v)}}};if count!=1{t.Fatal("move duplicated item")}
 _ = pinApp(a,false);if len(loadDesktopLayout().Pins)!=0{t.Fatal("unpin")}
}
func TestLayoutConcurrentWriters(t *testing.T){
 t.Setenv("HOME",t.TempDir());var wg sync.WaitGroup
 for i:=0;i<12;i++{wg.Add(1);go func(){defer wg.Done();if e:=placeCanvasItem(newCanvasItem("clock","desktop",0,0));e!=nil{t.Error(e)}}()};wg.Wait()
 if len(loadDesktopLayout().Items)!=14{t.Fatal("lost concurrent layout update")}
}
func TestFrameMoveAndDissolvePreservesChildren(t *testing.T){
 t.Setenv("HOME",t.TempDir())
 frame:=newCanvasItem("frame","start",20,30);if e:=placeCanvasItem(frame);e!=nil{t.Fatal(e)}
 child:=newCanvasItem("note","frame:"+frame.ID,5,8);child.Text="Keep this";_ = placeCanvasItem(child)
 nested:=newCanvasItem("frame","frame:"+frame.ID,0,0);_ = placeCanvasItem(nested)
 invalid:=frame;invalid.Surface="frame:"+nested.ID;if placeCanvasItem(invalid)==nil{t.Fatal("cycle accepted")}
 frame.Surface="desktop";_ = placeCanvasItem(frame);_ = removeCanvasItem(frame.ID)
 for _,v:=range loadDesktopLayout().Items{if v.ID==child.ID{if v.Surface!="desktop"||v.Text!="Keep this"||v.X!=25||v.Y!=38{t.Fatal(v)};return}}
 t.Fatal("child lost when dissolving frame")
}

func TestAppGridOrderingAndDissolve(t *testing.T) {
 t.Setenv("HOME",t.TempDir())
 grid:=newCanvasItem("app-grid","start",10,20);_ = placeCanvasItem(grid)
 a:=newCanvasItem("app","start",0,0);a.App=installedApp{ID:"z.desktop",Name:"Zulu"}
 b:=newCanvasItem("app","start",0,0);b.App=installedApp{ID:"a.desktop",Name:"Alpha"}
 _ = changeDesktopLayout(func(p *desktopLayout){if !insertGridApp(p,grid.ID,a,"")||!insertGridApp(p,grid.ID,b,""){t.Fatal("drop failed")}})
 items:=gridItems(loadDesktopLayout(),grid.ID);if len(items)!=2||items[0].ID!=a.ID{t.Fatal(items)}
 added:=items[1].AddedAt
 _ = changeDesktopLayout(func(p *desktopLayout){for i:=range p.Items{if p.Items[i].ID==grid.ID{p.Items[i].Sort="alphabet"}}})
 items=gridItems(loadDesktopLayout(),grid.ID);if items[0].ID!=b.ID{t.Fatal("alphabet")}
 _ = changeDesktopLayout(func(p *desktopLayout){insertGridApp(p,grid.ID,items[0],a.ID)})
 items=gridItems(loadDesktopLayout(),grid.ID);if items[0].ID!=b.ID||items[0].AddedAt!=added{t.Fatal("reorder lost timestamp")}
 _ = changeDesktopLayout(func(p *desktopLayout){if insertGridApp(p,grid.ID,newCanvasItem("note","start",0,0),""){t.Fatal("accepted widget")}})
 if e:=placeCanvasItem(items[0]);e!=nil{t.Fatal(e)}
 _ = removeCanvasItem(grid.ID)
 for _,v:=range loadDesktopLayout().Items{if v.ID==a.ID||v.ID==b.ID{if v.Surface!="start"{t.Fatal("app not preserved")}}}
}
