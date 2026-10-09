package main

import "testing"

func TestSetupWiFiSkipsEthernetOnly(t *testing.T) {
 if got:=setupWiFiDevice("enp0s3:ethernet\nlo:loopback\n");got!=""{t.Fatal(got)}
 if got:=setupWiFiDevice("enp0s3:ethernet\nwlan0:wifi\n");got!="wlan0"{t.Fatal(got)}
}

func TestSetupWiFiEscapedSSIDAndStrongestAP(t *testing.T) {
 raw:="AA\\:BB\\:CC\\:DD\\:EE\\:01:Home\\: Network:WPA2:30\nAA\\:BB\\:CC\\:DD\\:EE\\:02:Home\\: Network:WPA2:85\nAA\\:BB\\:CC\\:DD\\:EE\\:03::WPA2:90\ninvalid:Other:WPA2:99"
 points:=setupAccessPoints(raw)
 if len(points)!=1||points[0].ssid!="Home: Network"||points[0].bssid!="AA:BB:CC:DD:EE:02"{t.Fatalf("Unexpected networks: %+v",points)}
}
