package main
/*
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>
static void layout_preview(void *context,void *widget) {
 GtkWidget *w=GTK_WIDGET(widget);int width=gtk_widget_get_allocated_width(w),height=gtk_widget_get_allocated_height(w);
 if(width<1||height<1)return;
 cairo_surface_t *surface=cairo_image_surface_create(CAIRO_FORMAT_ARGB32,width,height);
 cairo_t *cr=cairo_create(surface);gtk_widget_draw(w,cr);cairo_destroy(cr);
 GdkPixbuf *pix=gdk_pixbuf_get_from_surface(surface,0,0,width,height);
 if(pix){gtk_drag_set_icon_pixbuf(context,pix,20,16);g_object_unref(pix);}cairo_surface_destroy(surface);
}
static void layout_icon(void *context,void *image) {
 GtkImage *im=GTK_IMAGE(image);GdkPixbuf *pix=NULL;
 if(gtk_image_get_storage_type(im)==GTK_IMAGE_PIXBUF) {
  pix=gtk_image_get_pixbuf(im);if(pix)g_object_ref(pix);
 } else if(gtk_image_get_storage_type(im)==GTK_IMAGE_ICON_NAME) {
  const gchar *name=NULL;GtkIconSize size;gtk_image_get_icon_name(im,&name,&size);
  if(name)pix=gtk_icon_theme_load_icon(gtk_icon_theme_get_default(),name,24,GTK_ICON_LOOKUP_FORCE_SIZE,NULL);
 }
 if(pix){gtk_drag_set_icon_pixbuf(context,pix,12,12);g_object_unref(pix);}
 else gtk_drag_set_icon_name(context,"application-x-executable",12,12);
}
static void layout_finish(void *ctx,int ok,unsigned int time) {gtk_drag_finish(ctx,ok,FALSE,time);}
static void layout_motion(void *event,double *x,double *y){gdk_event_get_root_coords(event,x,y);}
static void layout_set(void *s,const char *text,int n) {gtk_selection_data_set(s,gdk_atom_intern_static_string("application/x-miracleos-layout"),8,(const guchar*)text,n);}
static const guchar *layout_get(void *s,int *n) {*n=gtk_selection_data_get_length(s);return gtk_selection_data_get_data(s);}
*/
import "C"
import (
 "encoding/json"
 "unsafe"
 "github.com/gotk3/gotk3/gdk"
 "github.com/gotk3/gotk3/gtk"
)
type layoutDrag struct {AppID string `json:"app_id,omitempty"`;ItemID string `json:"item_id,omitempty"`}
func layoutDragSource(w gtk.IWidget,payload layoutDrag,preview ...gtk.IWidget) {
 target,_:=gtk.TargetEntryNew("application/x-miracleos-layout",0,0)
 widget:=w.ToWidget();widget.DragSourceSet(gdk.BUTTON1_MASK,[]gtk.TargetEntry{*target},gdk.ACTION_COPY|gdk.ACTION_MOVE)
 widget.Connect("drag-begin",func(_ interface{},ctx *gdk.DragContext){if len(preview)>0 {C.layout_preview(unsafe.Pointer(ctx.GObject),unsafe.Pointer(preview[0].ToWidget().GObject));return};if app,ok:=draggedApp(payload);ok {icon:=installedAppIcon(app);C.layout_icon(unsafe.Pointer(ctx.GObject),unsafe.Pointer(icon.GObject));icon.Destroy()} else {gtk.DragSetIconName(ctx,"applications-other",12,12)}})
 widget.Connect("drag-data-get",func(_ interface{},_ *gdk.DragContext,data *gtk.SelectionData,_ uint,_ uint){
  bytes,_:=json.Marshal(payload);text:=C.CString(string(bytes));defer C.free(unsafe.Pointer(text));C.layout_set(unsafe.Pointer(data),text,C.int(len(bytes)))
 })
}
func layoutDropTarget(w gtk.IWidget,drop func(layoutDrag,int,int) bool) {
 target,_:=gtk.TargetEntryNew("application/x-miracleos-layout",0,0);widget:=w.ToWidget()
 widget.DragDestSet(gtk.DEST_DEFAULT_ALL,[]gtk.TargetEntry{*target},gdk.ACTION_COPY|gdk.ACTION_MOVE)
 widget.Connect("drag-data-received",func(_ interface{},ctx *gdk.DragContext,x,y int,data *gtk.SelectionData,_ uint,stamp uint){
  var n C.int;b:=C.layout_get(unsafe.Pointer(data),&n);var payload layoutDrag
  ok:=n>0 && n<16384 && json.Unmarshal(C.GoBytes(unsafe.Pointer(b),n),&payload)==nil
  if ok {ok=drop(payload,x,y)}
  finishLayoutDrop(ctx,ok,stamp)
 })
}
func draggedApp(p layoutDrag)(installedApp,bool) {
 id:=p.AppID
 if p.ItemID!="" {for _,item:=range loadDesktopLayout().Items {if item.ID==p.ItemID && item.Kind=="app" {id=item.App.ID;break}}}
 for _,app:=range installedApplications(){if app.ID==id{return app,true}}
 return installedApp{},false
}
func finishLayoutDrop(ctx *gdk.DragContext,ok bool,stamp uint){success:=0;if ok{success=1};C.layout_finish(unsafe.Pointer(ctx.GObject),C.int(success),C.uint(stamp))}
func layoutMotion(event *gdk.Event)(float64,float64){var x,y C.double;C.layout_motion(unsafe.Pointer(event.Native()),&x,&y);return float64(x),float64(y)}

func dropOnSurface(surface string,p layoutDrag,x,y int) bool {
 var item canvasItem;found:=false
 if p.ItemID!="" {for _,v:=range loadDesktopLayout().Items {if v.ID==p.ItemID{item=v;found=true;break}}}else if app,ok:=draggedApp(p);ok {item=newCanvasItem("app",surface,x,y);item.App=app;item.Width=100;item.Height=105;found=true}
 if !found{return false};item.Surface=surface;item.X=max(0,x-20);item.Y=max(0,y-16);return placeCanvasItem(item)==nil
}
