package main
import (
 "fmt"
 "os/exec"
 "regexp"
 "strings"
 "github.com/gotk3/gotk3/gdk"
 "github.com/gotk3/gotk3/glib"
 "github.com/gotk3/gotk3/gtk"
)
func appContextActions(app installedApp)[]contextAction {
 pinLabel:="An die Leiste anheften";pin:=!pinnedApp(app.ID);if !pin{pinLabel="Von der Leiste lösen"}
 return []contextAction{{"Öffnen",func(){launchInstalledApp(app)},false},{pinLabel,func(){_ = pinApp(app,pin)},false},{"Deinstallieren …",func(){uninstallApp(app)},false}}
}
func attachAppContext(w gtk.IWidget,owner *gtk.Window,app installedApp){
 widget:=w.ToWidget();widget.AddEvents(int(gdk.BUTTON_PRESS_MASK));widget.Connect("button-press-event",func(_ interface{},event *gdk.Event)bool{if gdk.EventButtonNewFromEvent(event).Button()!=3{return false};glassContextMenu(owner,event,appContextActions(app));return true})
}
func uninstallApp(app installedApp){
 // Resolve the owning package, then show APT's complete removal plan before asking
 // for authentication. No shell interpolation or automatic dependency removal.
 go func(){
  out,err:=exec.Command("dpkg-query","-S",app.File).Output()
  pkg:="";if err==nil {pkg=strings.SplitN(string(out),": ",2)[0]}
  if !regexp.MustCompile(`^[a-z0-9][a-z0-9+.-]*(?::[a-z0-9]+)?$`).MatchString(pkg){glib.IdleAdd(func(){appActionMessage("Diese App gehört zu keinem erkannten Debian-Paket. Bitte verwende ihren ursprünglichen Paketmanager.")});return}
  plan,e:=exec.Command("apt-get","--simulate","remove","--",pkg).CombinedOutput()
  glib.IdleAdd(func(){
   if e!=nil{appActionMessage(string(plan));return}
   dialog:=gtk.MessageDialogNew(nil,gtk.DIALOG_MODAL,gtk.MESSAGE_QUESTION,gtk.BUTTONS_OK_CANCEL,"%s deinstallieren?\n\n%s",app.Name,string(plan))
   dialog.SetTitle("App deinstallieren — MiracleOS")
   dialog.Connect("response",func(_ *gtk.Dialog,response gtk.ResponseType){dialog.Destroy();if response!=gtk.RESPONSE_OK{return};go func(){result,e:=exec.Command("pkexec","/usr/bin/apt-get","-y","remove","--",pkg).CombinedOutput();glib.IdleAdd(func(){if e!=nil{appActionMessage("Deinstallation nicht abgeschlossen:\n"+string(result))}else{_ = pinApp(app,false);appActionMessage(app.Name+" wurde deinstalliert.")}})}()})
   dialog.ShowAll()
  })
 }()
}
func appActionMessage(text string){d:=gtk.MessageDialogNew(nil,0,gtk.MESSAGE_INFO,gtk.BUTTONS_CLOSE,"%s",text);d.SetTitle("Apps — MiracleOS");d.Connect("response",func(){d.Destroy()});d.ShowAll()}
var _=fmt.Sprintf