// Package apptabs connects application-owned tabs to the MiracleOS frame.
package apptabs

import (
 "encoding/json"
 "fmt"
 "net"
 "os"
 "path/filepath"
 "sync"
)

type Tab struct { ID string `json:"id"`; Title string `json:"title"` }
type State struct { Name string `json:"name"`; Active string `json:"active"`; Tabs []Tab `json:"tabs"`; CanNew bool `json:"can_new"` }
type Action struct { Action string `json:"action"`; ID string `json:"id"` }
type Client struct { socket *net.UnixConn; path string; mu sync.Mutex }
func directory() string { return filepath.Join(os.Getenv("XDG_RUNTIME_DIR"),"miracleos-app-tabs") }
func Enabled() bool { return os.Getenv("MIRACLEOS_GLOBAL_TABS")=="1" && filepath.IsAbs(os.Getenv("XDG_RUNTIME_DIR")) }
func Connect(handle func(Action)) (*Client,error) {
 if !Enabled() { return nil,nil }
 dir:=directory(); if err:=os.MkdirAll(dir,0700);err!=nil { return nil,err }
 path:=filepath.Join(dir,fmt.Sprintf("native-%d",os.Getpid()))
 conn,err:=net.ListenUnixgram("unixgram",&net.UnixAddr{Name:path+".sock",Net:"unixgram"});if err!=nil{return nil,err}
 os.Chmod(path+".sock",0600)
 c:=&Client{socket:conn,path:path}
 go func(){ b:=make([]byte,4096);for { n,_,e:=conn.ReadFromUnix(b);if e!=nil{return};var a Action;if json.Unmarshal(b[:n],&a)==nil&&(a.Action=="select"||a.Action=="close"||a.Action=="new"){handle(a)} } }()
 return c,nil
}
func(c *Client) Publish(s State) error { if c==nil{return nil};c.mu.Lock();defer c.mu.Unlock();return atomicJSON(c.path+".json",s) }
func(c *Client) Close() { if c==nil{return};c.mu.Lock();defer c.mu.Unlock();c.socket.Close();os.Remove(c.path+".sock");os.Remove(c.path+".json") }
func atomicJSON(path string,value interface{}) error { b,e:=json.Marshal(value);if e!=nil{return e};if e=os.WriteFile(path+".tmp",b,0600);e!=nil{return e};return os.Rename(path+".tmp",path) }
