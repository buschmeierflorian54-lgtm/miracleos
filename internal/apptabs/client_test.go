package apptabs

import (
 "encoding/json"
 "net"
 "os"
 "path/filepath"
 "testing"
 "time"
)
func TestNativeTransport(t *testing.T){
 t.Setenv("MIRACLEOS_GLOBAL_TABS","1");t.Setenv("XDG_RUNTIME_DIR",t.TempDir())
 received:=make(chan Action,1);c,e:=Connect(func(a Action){received<-a});if e!=nil{t.Fatal(e)};defer c.Close()
 s:=State{Name:"Dateien",Active:"1",CanNew:true,Tabs:[]Tab{{"1","Dokumente"},{"2","Downloads"}}}
 if e=c.Publish(s);e!=nil{t.Fatal(e)};raw,e:=os.ReadFile(c.path+".json");if e!=nil{t.Fatal(e)};var got State;if json.Unmarshal(raw,&got)!=nil||len(got.Tabs)!=2||got.Active!="1"{t.Fatal(string(raw))}
 info,_:=os.Stat(c.path+".json");if info.Mode().Perm()!=0600{t.Fatal(info.Mode())}
 conn,e:=net.DialUnix("unixgram",nil,&net.UnixAddr{Name:c.path+".sock",Net:"unixgram"});if e!=nil{t.Fatal(e)};defer conn.Close()
 conn.Write([]byte(`{"action":"select","id":"2"}`));select{case a:=<-received:if a.Action!="select"||a.ID!="2"{t.Fatal(a)};case <-time.After(time.Second):t.Fatal("missing native action")}
 c.Close();if _,e=os.Stat(c.path+".json");!os.IsNotExist(e){t.Fatal("snapshot not removed")}
}
func TestCollectKeepsWorkspaceAndExcludesDialogs(t *testing.T){
 var tree node
 json.Unmarshal([]byte(`{"type":"workspace","name":"2","nodes":[{"id":1,"pid":5,"app_id":"editor"},{"id":2,"pid":5,"window_properties":{"class":"editor","transient_for":42}}],"floating_nodes":[{"id":3,"pid":6,"app_id":"editor"}]}`),&tree)
 out:=map[int64]node{};collect(tree,"",out)
 if len(out)!=2||out[1].workspace!="2"||out[3].workspace!="2"{t.Fatal(out)}
}
func TestNativeStateRejectsUnboundedOrEmptyData(t *testing.T){
 t.Setenv("XDG_RUNTIME_DIR",t.TempDir());os.MkdirAll(directory(),0700)
 os.WriteFile(filepath.Join(directory(),"native-42.json"),[]byte(`{"name":"Empty","tabs":[]}`),0600)
 if _,ok:=nativeState(42);ok{t.Fatal("empty provider accepted")}
}
