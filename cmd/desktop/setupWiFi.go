package main

import (
 "context"
 "net"
 "os/exec"
 "strconv"
 "strings"
 "time"
 "github.com/gotk3/gotk3/glib"
 "github.com/gotk3/gotk3/gtk"
)

func splitNMFields(line string) []string {
 fields:=[]string{};var value strings.Builder;escaped:=false
 for _,r:=range line {if escaped {value.WriteRune(r);escaped=false} else if r=='\\' {escaped=true} else if r==':' {fields=append(fields,value.String());value.Reset()} else {value.WriteRune(r)}}
 if escaped {value.WriteRune('\\')};return append(fields,value.String())
}

func setupWiFiDevice(output string) string {
 for _,line:=range strings.Split(output,"\n") {fields:=splitNMFields(line);if len(fields)==2&&fields[1]=="wifi" {return fields[0]}}
 return ""
}

type setupAccessPoint struct { bssid,ssid,security string; signal int }

func setupAccessPoints(output string) []setupAccessPoint {
 points:=[]setupAccessPoint{};seen:=map[string]int{}
 for _,line:=range strings.Split(output,"\n") {
  f:=splitNMFields(line);if len(f)!=4||f[1]==""{continue};if _,err:=net.ParseMAC(f[0]);err!=nil{continue}
  strength,_:=strconv.Atoi(f[3]);point:=setupAccessPoint{f[0],f[1],f[2],strength}
  if index,ok:=seen[f[1]];ok {if strength>points[index].signal{points[index]=point};continue}
  seen[f[1]]=len(points);points=append(points,point)
 };return points
}

func newSetupWiFi() (*gtk.Box,bool) {
 ctx,cancel:=context.WithTimeout(context.Background(),3*time.Second)
 output,_:=exec.CommandContext(ctx,"nmcli","--colors","no","-t","-f","DEVICE,TYPE","device","status").Output();cancel()
 device:=setupWiFiDevice(string(output))
 box,_:=gtk.BoxNew(gtk.ORIENTATION_VERTICAL,12)
 if device=="" {return box,false}
 networks,_:=gtk.ComboBoxTextNew();box.PackStart(networks,false,false,0)
 password,_:=gtk.EntryNew();password.SetVisibility(false);password.SetInputPurpose(gtk.INPUT_PURPOSE_PASSWORD);password.SetPlaceholderText("WLAN-Passwort");menuClass(password,"setup-entry");box.PackStart(password,false,false,0)
 note:=settingsText("WLAN-Netze werden gesucht …");box.PackStart(note,false,false,0)
 controls,_:=gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL,12);box.PackStart(controls,false,false,0)
 var choices []setupAccessPoint;alive,busy:=true,false
 box.Connect("destroy",func(){alive=false})
 var scan,connect *gtk.Button
 setBusy:=func(value bool){busy=value;scan.SetSensitive(!value);connect.SetSensitive(!value);networks.SetSensitive(!value);password.SetSensitive(!value)}
 refresh:=func(enable bool){
  if busy{return};setBusy(true);note.SetText("WLAN-Netze werden gesucht …")
  go func(){
   ctx,cancel:=context.WithTimeout(context.Background(),20*time.Second);defer cancel()
   if enable {_=exec.CommandContext(ctx,"nmcli","radio","wifi","on").Run()}
   raw,err:=exec.CommandContext(ctx,"nmcli","--colors","no","-t","-f","BSSID,SSID,SECURITY,SIGNAL","device","wifi","list","ifname",device,"--rescan","auto").Output()
   points:=setupAccessPoints(string(raw))
   glib.IdleAdd(func(){if !alive{return};setBusy(false);choices=points;networks.RemoveAll();for _,p:=range choices {security:=p.security;if security==""||security=="--"{security="Offen"};networks.AppendText(p.ssid+" · "+security+" · "+strconv.Itoa(p.signal)+" %")};if len(choices)>0{networks.SetActive(0);note.SetText("Wähle dein WLAN und verbinde dich.")}else if err!=nil{note.SetText("WLAN-Suche fehlgeschlagen. Prüfe, ob WLAN eingeschaltet ist.")}else{note.SetText("Kein WLAN gefunden. Du kannst erneut suchen oder später verbinden.")}})
  }()
 }
 scan=settingsButton("WLAN einschalten und suchen",func(){refresh(true)});controls.PackStart(scan,false,false,0)
 connect=settingsButton("Verbinden",func(){
  index:=networks.GetActive();if busy||index<0||index>=len(choices){return};point:=choices[index]
  if strings.Contains(point.security,"802.1X"){note.SetText("Dieses Unternehmensnetz benötigt zusätzliche Anmeldedaten. Richte es nach der Ersteinrichtung in den Netzwerkeinstellungen ein.");return}
  secret,_:=password.GetText();if strings.ContainsAny(secret,"\r\n\x00"){note.SetText("Ungültiges WLAN-Passwort.");return}
  setBusy(true);note.SetText("Verbindung wird hergestellt …")
  go func(){
   ctx,cancel:=context.WithTimeout(context.Background(),40*time.Second);defer cancel()
   // Passwords travel through stdin, never in process arguments or logs.
   cmd:=exec.CommandContext(ctx,"nmcli","--colors","no","--wait","35","--ask","device","wifi","connect",point.bssid,"ifname",device)
   cmd.Stdin=strings.NewReader(secret+"\n");err:=cmd.Run()
   glib.IdleAdd(func(){if !alive{return};setBusy(false);if err!=nil{note.SetText("Verbindung fehlgeschlagen. Prüfe dein WLAN-Passwort und versuche es erneut.")}else{password.SetText("");note.SetText("Verbunden mit "+point.ssid+". Du kannst weitergehen.")}})
  }()
 });controls.PackStart(connect,false,false,0)
 refresh(false)
 return box,true
}
