package apptabs

import (
 "bufio"
 "context"
 "encoding/json"
 "fmt"
 "log"
 "net"
 "os"
 "os/exec"
 "path/filepath"
 "sort"
 "strconv"
 "strings"
 "time"
)
type rect struct{ X,Y,Width,Height int }
type node struct {
	Deco rect `json:"deco_rect"`; Fullscreen int `json:"fullscreen_mode"`
 ID int64 `json:"id"`; PID int `json:"pid"`; Name string `json:"name"`; Type string `json:"type"`; AppID string `json:"app_id"`; Focused bool `json:"focused"`; Visible bool `json:"visible"`; Rect rect `json:"rect"`
 WindowProperties struct{ Class string `json:"class"`; TransientFor interface{} `json:"transient_for"` } `json:"window_properties"`
 Nodes []node `json:"nodes"`; Floating []node `json:"floating_nodes"`
 workspace string
	workarea rect
}
type group struct { members []int64; active int64; workspace string; geometry rect }
type broker struct { windows map[int64]node; groups map[string]*group; states map[int64]State; pending map[string]time.Time; native map[int64]int }
func collect(n node,ws string,out map[int64]node) {
 collectArea(n,ws,rect{},out)
}
func collectArea(n node,ws string,area rect,out map[int64]node) {
 if n.Type=="workspace"{ws=n.Name;area=n.Rect};n.workspace=ws;n.workarea=area
 if (n.AppID!=""||n.WindowProperties.Class!="")&&n.WindowProperties.TransientFor==nil {out[n.ID]=n}
 for _,c:=range n.Nodes {collectArea(c,ws,area,out)};for _,c:=range n.Floating {collectArea(c,ws,area,out)}
}
func identity(n node) string { if n.AppID!=""{return strings.ToLower(n.AppID)};return strings.ToLower(n.WindowProperties.Class) }
func sway(args ...string)([]byte,error){ctx,cancel:=context.WithTimeout(context.Background(),3*time.Second);defer cancel();return exec.CommandContext(ctx,"swaymsg",append([]string{"-r"},args...)...).Output()}
func command(id int64,cmd string) error { b,e:=sway(fmt.Sprintf("[con_id=%d] %s",id,cmd));if e!=nil{return e};var replies []struct{Success bool};if json.Unmarshal(b,&replies)!=nil||len(replies)==0||!replies[0].Success{return fmt.Errorf("Sway rejected %s",cmd)};return nil }
func nativeState(pid int)(State,bool){var s State;b,e:=os.ReadFile(filepath.Join(directory(),fmt.Sprintf("native-%d.json",pid)));if e!=nil||len(b)>65536||json.Unmarshal(b,&s)!=nil||len(s.Tabs)==0||len(s.Tabs)>256{return s,false};return s,true}
func (b *broker) refresh() {
 data,e:=sway("-t","get_tree");if e!=nil{return};var tree node;if json.Unmarshal(data,&tree)!=nil{return}
 windows:=map[int64]node{};collect(tree,"",windows)
 for id,n:=range windows {if _,known:=b.windows[id];!known&&n.workspace!="__i3_scratch"&&n.Fullscreen==0&&n.workarea.Width>400&&n.workarea.Height>500 {
  if n.Deco.Y<n.workarea.Y||n.Rect.Width>n.workarea.Width||n.Rect.Y+n.Rect.Height>n.workarea.Y+n.workarea.Height {
   width:=n.Rect.Width;if width>n.workarea.Width-80{width=n.workarea.Width-80};height:=n.Rect.Height;if height>n.workarea.Height-120{height=n.workarea.Height-120}
   command(id,fmt.Sprintf("resize set width %d px height %d px, move position center",width,height))
  }
 }}
 b.windows=windows;b.native=map[int64]int{}
 for id:=range b.states {if _,ok:=windows[id];!ok{os.Remove(filepath.Join(directory(),fmt.Sprintf("%d.json",id)));delete(b.states,id)}}
 // Native tab providers never join the generic window groups.
 generic:=map[string][]int64{}
 for id,n:=range windows {if s,ok:=nativeState(n.PID);ok{b.native[id]=n.PID;b.publish(id,s)}else{generic[identity(n)]=append(generic[identity(n)],id)}}
 for key,g:=range b.groups {if len(generic[key])==0{delete(b.groups,key);continue};kept:=g.members[:0];for _,id:=range g.members{if _,ok:=windows[id];ok{kept=append(kept,id)}};g.members=kept}
 for key,ids:=range generic {
  sort.Slice(ids,func(i,j int)bool{return ids[i]<ids[j]});g:=b.groups[key];if g==nil{g=&group{};b.groups[key]=g}
  fresh:=[]int64{};for _,id:=range ids {found:=false;for _,old:=range g.members{if old==id{found=true}};if !found{g.members=append(g.members,id);fresh=append(fresh,id)}}
  previousActive:=g.active
  if _,ok:=windows[g.active];!ok{g.active=0}
  for _,id:=range ids {n:=windows[id];if n.Focused {g.active=id}}
  if g.active==0 {g.active=ids[0];for _,id:=range ids{if windows[id].workspace!="__i3_scratch"{g.active=id;break}}}
  if previousActive!=0 {if _,alive:=windows[previousActive];!alive&&g.workspace!="" {b.selectWindow(g,g.active)}}
  current:=windows[g.active];if current.workspace!="__i3_scratch" {g.workspace=current.workspace;g.geometry=current.Rect}
  if deadline,ok:=b.pending[key];ok {if len(fresh)>0 {delete(b.pending,key);b.selectWindow(g,fresh[len(fresh)-1])} else if time.Now().After(deadline){delete(b.pending,key);log.Printf("%s opened no additional window",key)}}
  // Never pull a minimized group back into view during reconciliation.
  if windows[g.active].workspace!="__i3_scratch" {for _,id:=range g.members {if id!=g.active&&windows[id].workspace!="__i3_scratch"{command(id,"mark --add __miracleos_tab_hidden, move scratchpad")}}}
  launcher:=b.launcher(current)
  s:=State{Name:applicationName(current,launcher),Active:strconv.FormatInt(g.active,10),CanNew:launcher!=""}
  for _,id:=range g.members{s.Tabs=append(s.Tabs,Tab{ID:strconv.FormatInt(id,10),Title:windows[id].Name})}
  for _,id:=range g.members{b.publish(id,s)}
 }
}
func(b *broker) publish(id int64,s State) {
 old,_:=json.Marshal(b.states[id]);now,_:=json.Marshal(s);if string(old)==string(now){return}
 if atomicJSON(filepath.Join(directory(),fmt.Sprintf("%d.json",id)),s)!=nil{return};b.states[id]=s
 // Re-render the compositor header without changing the application's title.
 command(id,"title_format %title")
}
func(b *broker) selectWindow(g *group,id int64) {
 n,ok:=b.windows[id];if !ok||g.workspace==""{return}
 cmd:="move container to workspace "+strconv.Quote(g.workspace)+", floating enable"
 if g.geometry.Width>0 {cmd+=fmt.Sprintf(", resize set width %d px height %d px, move position %d px %d px",g.geometry.Width,g.geometry.Height,g.geometry.X,g.geometry.Y)}
 cmd+=", unmark __miracleos_tab_hidden, focus"
 if command(n.ID,cmd)!=nil{return};old:=g.active;g.active=id
 if old!=id{command(old,"mark --add __miracleos_tab_hidden, move scratchpad")}
}
func applicationName(n node, launcher string) string {
 data,_:=os.ReadFile(launcher)
 section:=false
 for _,line:=range strings.Split(string(data),"\n") {
  line=strings.TrimSpace(line)
  if strings.HasPrefix(line,"["){section=line=="[Desktop Entry]"}
  if section&&strings.HasPrefix(line,"Name=") {
   if name:=strings.TrimSpace(strings.TrimPrefix(line,"Name="));name!=""{return name}
  }
 }
 if n.WindowProperties.Class!=""{return n.WindowProperties.Class}
 return n.AppID
}
func(b *broker) launcher(n node) string {
 dirs:=[]string{filepath.Join(os.Getenv("HOME"),".local/share/applications"),"/usr/local/share/applications","/usr/share/applications","/var/lib/flatpak/exports/share/applications"}
 key:=identity(n)
 for _,dir:=range dirs {entries,_:=os.ReadDir(dir);for _,ent:=range entries {if !strings.HasSuffix(ent.Name(),".desktop"){continue};path:=filepath.Join(dir,ent.Name());data,e:=os.ReadFile(path);if e!=nil{continue};section:=false;wm:="";hidden:=false
  for _,line:=range strings.Split(string(data),"\n"){line=strings.TrimSpace(line);if strings.HasPrefix(line,"["){section=line=="[Desktop Entry]"};if section&&strings.HasPrefix(line,"StartupWMClass="){wm=strings.ToLower(strings.TrimPrefix(line,"StartupWMClass="))};if section&&line=="Hidden=true"{hidden=true}}
  if !hidden&&(strings.ToLower(strings.TrimSuffix(ent.Name(),".desktop"))==key||wm==key){return path}
 }};return ""
}
func(b *broker) action(id int64,a Action){
 s,ok:=b.states[id];if !ok{return}
 if a.Action!="new" {found:=false;for _,t:=range s.Tabs{if t.ID==a.ID{found=true}};if !found{return}}
 if pid:=b.native[id];pid>0 {conn,e:=net.DialUnix("unixgram",nil,&net.UnixAddr{Name:filepath.Join(directory(),fmt.Sprintf("native-%d.sock",pid)),Net:"unixgram"});if e==nil {data,_:=json.Marshal(a);conn.Write(data);conn.Close()};return}
 n,ok:=b.windows[id];if !ok{return};g:=b.groups[identity(n)];if g==nil{return}
 switch a.Action {
 case "select":target,e:=strconv.ParseInt(a.ID,10,64);if e==nil{b.selectWindow(g,target)}
 case "close":target,e:=strconv.ParseInt(a.ID,10,64);if e==nil{command(target,"kill")}
 case "new":key:=identity(n);if _,busy:=b.pending[key];busy{return};path:=b.launcher(n);if path==""{return};cmd:=exec.Command("gio","launch",path);if cmd.Start()==nil {b.pending[key]=time.Now().Add(10*time.Second);go cmd.Wait()}
 }
}
// Run is a single-owner event loop; no GTK and no shell command construction.
func Run(ctx context.Context) error {
 if !Enabled(){return fmt.Errorf("global tabs are disabled")};if e:=os.MkdirAll(directory(),0700);e!=nil{return e}
 events:=make(chan struct{ID int64; Action Action},64)
 cmd:=exec.CommandContext(ctx,"swaymsg","-m","-r","-t","subscribe","[\"window\"]");out,e:=cmd.StdoutPipe();if e!=nil{return e};if e=cmd.Start();e!=nil{return e};defer cmd.Process.Kill()
 go func(){defer close(events);scanner:=bufio.NewScanner(out);scanner.Buffer(make([]byte,4096),1048576);for scanner.Scan(){var ev struct{Change string `json:"change"`;Container node `json:"container"`};if json.Unmarshal(scanner.Bytes(),&ev)!=nil{continue};if !strings.HasPrefix(ev.Change,"miracleos-tab:"){continue};parts:=strings.SplitN(strings.TrimPrefix(ev.Change,"miracleos-tab:"),":",2);a:=Action{Action:parts[0]};if len(parts)>1{a.ID=parts[1]};select{case events<-struct{ID int64;Action Action}{ev.Container.ID,a}:case <-ctx.Done():return}}}()
 b:=&broker{windows:map[int64]node{},groups:map[string]*group{},states:map[int64]State{},pending:map[string]time.Time{},native:map[int64]int{}};b.refresh();tick:=time.NewTicker(500*time.Millisecond);defer tick.Stop()
 for{select{case <-ctx.Done():return nil;case ev,ok:=<-events:if !ok{return cmd.Wait()};b.action(ev.ID,ev.Action);b.refresh();case <-tick.C:b.refresh()}}
}
