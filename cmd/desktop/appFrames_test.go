package main

import (
 "os"
 "path/filepath"
 "strings"
 "testing"
)
func TestAppFramesPersistIndependentRules(t *testing.T) {
 t.Setenv("HOME",t.TempDir())
 p:=systemPreferences{Accent:"#3d9894",FontSize:12,AppFrames:map[string]appFramePreference{
 "spotify.desktop":appFrameRule(installedApp{ID:"spotify.desktop",Name:"Spotify",WMClass:"Spotify"},"#000000"),
 "files.desktop":appFrameRule(installedApp{ID:"files.desktop",Name:"Dateien"},"#c8e2e1"),
 }}
 if e:=writeWindowTheme(p);e!=nil {t.Fatal(e)}
 path:=filepath.Join(filepath.Dir(preferencesPath()),"app-frames.tsv")
 b,e:=os.ReadFile(path);if e!=nil {t.Fatal(e)}
 for _,want:=range []string{"class\tSpotify\t#000000\t#fff0fb","title\tDateien — MiracleOS\t#c8e2e1\t#352737"} {if !strings.Contains(string(b),want){t.Fatalf("missing %q in %s",want,b)}}
 delete(p.AppFrames,"spotify.desktop")
 if e:=writeWindowTheme(p);e!=nil {t.Fatal(e)}
 b,_=os.ReadFile(path)
 if strings.Contains(strings.ToLower(string(b)),"spotify") || !strings.Contains(string(b),"Dateien") {t.Fatalf("reset affected wrong app: %s",b)}
}