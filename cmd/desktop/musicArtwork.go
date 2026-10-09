package main
import("fmt";"io";"net/http";"net/url";"os";"time")
func readMusicArtwork(raw string)([]byte,error){
 u,e:=url.Parse(raw);if e!=nil{return nil,e};var reader io.ReadCloser
 switch u.Scheme{
 case "https","http":
  client:=http.Client{Timeout:5*time.Second};response,e:=client.Get(raw);if e!=nil{return nil,e};if response.StatusCode!=200{response.Body.Close();return nil,fmt.Errorf("cover HTTP %d",response.StatusCode)};reader=response.Body
 case "file":
  if u.Host!=""&&u.Host!="localhost"{return nil,fmt.Errorf("non-local artwork")};f,e:=os.Open(u.Path);if e!=nil{return nil,e};reader=f
 default:return nil,fmt.Errorf("unsupported artwork URI")
 }
 defer reader.Close();data,e:=io.ReadAll(io.LimitReader(reader,8*1024*1024+1));if len(data)>8*1024*1024{return nil,fmt.Errorf("artwork too large")};return data,e
}