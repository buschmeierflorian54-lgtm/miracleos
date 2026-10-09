package main
import "testing"
func TestWidgetFitsAvailableSpace(t *testing.T){
 for _,test:=range []struct{x,y,w,h,cw,ch int}{{600,400,260,170,650,440},{0,0,900,800,300,200},{-10,-20,100,100,650,440},{20,30,400,300,100,80}}{
  item:=fitCanvasItem(canvasItem{Kind:"note",X:test.x,Y:test.y,Width:test.w,Height:test.h},test.cw,test.ch)
  if item.X<0||item.Y<0||item.X+item.Width>test.cw||item.Y+item.Height>test.ch{t.Fatalf("outside canvas: %+v",item)}
 }
 item:=fitCanvasItem(canvasItem{Kind:"note",X:20,Y:30,Width:300,Height:180},650,440)
 if item.X!=20||item.Y!=30||item.Width!=300||item.Height!=180{t.Fatal("changed fitting widget")}
}
func TestMusicWidgetMinimumFitsSurface(t *testing.T){
 item:=fitCanvasItem(canvasItem{Kind:"music",X:600,Y:400,Width:260,Height:170},650,440)
 if item.Width!=260||item.Height!=250||item.X+item.Width>650||item.Y+item.Height>440{t.Fatalf("clipped music widget: %+v",item)}
 tiny:=fitCanvasItem(item,200,150);if tiny.Width>200||tiny.Height>150{t.Fatal("exceeded small surface")}
}
func TestAllWidgetMinimumsStayInsideSurface(t *testing.T){
 for _,kind:=range []string{"clock","clock-compact","note","notes","music","calendar","frame","app-grid"}{
  w,h:=widgetMinimum(kind);item:=fitCanvasItem(canvasItem{Kind:kind,Width:50,Height:50,X:620,Y:430},650,440)
  if item.Width<w||item.Height<h||item.X+item.Width>650||item.Y+item.Height>440{t.Fatalf("%s: %+v",kind,item)}
  small:=fitCanvasItem(item,120,90);if small.Width>120||small.Height>90{t.Fatal(kind)}
 }
}
