package main

// Keep the requested footprint within the usable surface, including edge drops.
func fitCanvasItem(item canvasItem,width,height int) canvasItem {
 if width<1||height<1{return item}
 w,h:=item.Width,item.Height
 mw,mh:=widgetMinimum(item.Kind);w=max(mw,w);h=max(mh,h);if item.Kind=="app"{w=32;h=32}
 w=max(1,min(w,width));h=max(1,min(h,height))
 item.X=max(0,min(item.X,width-w));item.Y=max(0,min(item.Y,height-h))
 item.Width=w;item.Height=h
 return item
}
func(c *widgetCanvas) bounds()(int,int) {
 w,h:=c.root.GetAllocatedWidth(),c.root.GetAllocatedHeight()
 if c.surface=="desktop" {h=max(1,h-80)}
 return w,h
}
func(c *widgetCanvas) fit(item canvasItem)canvasItem {
 w,h:=c.bounds();if w<=1||h<=1{return item};return fitCanvasItem(item,w,h)
}

func widgetMinimum(kind string)(int,int){switch kind{case "music":return 260,250;case "clock":return 160,190;case "clock-compact":return 190,120;case "notes":return 300,250;case "note":return 240,150;case "calendar":return 280,270;case "frame","app-grid":return 240,200};return 100,100}
