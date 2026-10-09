package main
import (
 "fmt"
 "math"
 "time"
 "github.com/gotk3/gotk3/cairo"
 "github.com/gotk3/gotk3/glib"
 "github.com/gotk3/gotk3/gtk"
)
type widgetNote struct {ID string `json:"id"`;Text string `json:"text"`;Done bool `json:"done"`}
func addDemoClock(shell *gtk.Box,compact bool,card *gtk.Box){
 shell.SetHExpand(true);shell.SetVExpand(true)
 alive:=true;shell.Connect("destroy",func(){alive=false})
 row,_:=gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL,14);row.SetHExpand(true);row.SetHAlign(gtk.ALIGN_FILL);row.SetVAlign(gtk.ALIGN_FILL);row.SetVExpand(true);shell.PackStart(row,true,true,0)
 face,_:=gtk.DrawingAreaNew();face.SetSizeRequest(100,100);face.SetHExpand(true);face.SetVExpand(true);face.SetHAlign(gtk.ALIGN_FILL);face.SetVAlign(gtk.ALIGN_FILL)
 face.Connect("draw",func(_ *gtk.DrawingArea,cr *cairo.Context)bool{
  now:=time.Now();cr.Save();scale:=math.Min(float64(face.GetAllocatedWidth()),float64(face.GetAllocatedHeight()))/130;cr.Translate(float64(face.GetAllocatedWidth())/2,float64(face.GetAllocatedHeight())/2);cr.Scale(scale,scale);cr.SetSourceRGBA(1,.85,.93,.12);cr.Arc(0,0,63,0,2*math.Pi);cr.Fill()
  for i:=0;i<12;i++{a:=float64(i)*math.Pi/6;cr.SetSourceRGBA(1,.85,.93,.7);cr.SetLineWidth(2);cr.MoveTo(math.Sin(a)*53,-math.Cos(a)*53);cr.LineTo(math.Sin(a)*59,-math.Cos(a)*59);cr.Stroke()}
  hand:=func(angle,length,width float64,r,g,b float64){cr.SetSourceRGB(r,g,b);cr.SetLineWidth(width);cr.MoveTo(0,0);cr.LineTo(math.Sin(angle)*length,-math.Cos(angle)*length);cr.Stroke()}
  hand((float64(now.Hour()%12)+float64(now.Minute())/60)*math.Pi/6,32,5,1,1,1);hand((float64(now.Minute())+float64(now.Second())/60)*math.Pi/30,48,3,1,.85,.93);hand(float64(now.Second())*math.Pi/30,54,1.5,.95,.38,.61)
  cr.SetSourceRGB(.95,.38,.61);cr.Arc(0,0,4,0,2*math.Pi);cr.Fill();cr.Restore();return false
 })
 if !compact{row.PackStart(face,true,true,0)}
 text,_:=gtk.BoxNew(gtk.ORIENTATION_VERTICAL,5);clock:=menuLabel("","demo-clock-time");day:=menuLabel("","demo-clock-day");date:=menuLabel("","demo-clock-date");clock.SetXAlign(0);day.SetXAlign(0);date.SetXAlign(0)
 text.SetVAlign(gtk.ALIGN_CENTER);text.PackStart(clock,false,false,0);text.PackStart(day,false,false,0);text.PackStart(date,false,false,0);row.PackStart(text,compact,compact,0)
 // Hidden children do not consume box space. Keep the text parented so that
 // resizing never removes and re-adds widgets during GTK allocation.
 if !compact{text.SetNoShowAll(true);text.Hide()}
 lastSmall:=false;first:=true
 card.Connect("size-allocate",func(){
  small:=card.GetAllocatedWidth()<300||card.GetAllocatedHeight()<210
  if !first&&small==lastSmall{return}
  first=false;lastSmall=small
  day.SetNoShowAll(small);date.SetNoShowAll(small)
  if small {day.Hide();date.Hide()}else{day.Show();date.Show()}
  if compact{return}
  if small{
   text.Hide()
   row.SetSpacing(0);face.SetSizeRequest(100,100)
  }else{
   text.SetNoShowAll(false);text.ShowAll();text.SetNoShowAll(true);row.SetSpacing(14);face.SetSizeRequest(100,100)
  }
 })
 update:=func()bool{if !alive{return false};now:=time.Now();format:="15:04";if compact{format="15:04:05"};clock.SetText(now.Format(format));day.SetText([]string{"Sonntag","Montag","Dienstag","Mittwoch","Donnerstag","Freitag","Samstag"}[now.Weekday()]);date.SetText(now.Format("02.01.2006"));if !compact{face.QueueDraw()};return true};update();glib.TimeoutAdd(1000,update)
}
func addDemoNotes(shell *gtk.Box,item canvasItem){
 notes:=append([]widgetNote(nil),item.Notes...)
 count:=menuLabel("","canvas-subtitle");count.SetXAlign(1);shell.PackStart(count,false,false,0)
 list,_:=gtk.BoxNew(gtk.ORIENTATION_VERTICAL,6);scroll,_:=gtk.ScrolledWindowNew(nil,nil);scroll.SetPolicy(gtk.POLICY_NEVER,gtk.POLICY_AUTOMATIC);scroll.SetMinContentHeight(50);scroll.Add(list);shell.PackStart(scroll,true,true,0)
 save:=func(){copyNotes:=append([]widgetNote(nil),notes...);_ = changeDesktopLayout(func(p *desktopLayout){for i:=range p.Items{if p.Items[i].ID==item.ID{p.Items[i].Notes=copyNotes}}})}
 var render func();render=func(){list.GetChildren().Foreach(func(v interface{}){v.(gtk.IWidget).ToWidget().Destroy()});count.SetText(fmt.Sprintf("%d Notizen",len(notes)))
  if len(notes)==0{list.PackStart(menuLabel("Noch keine Notizen.","canvas-subtitle"),true,true,0)}
  for _,n:=range notes{n:=n;row,_:=gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL,6);menuClass(row,"demo-note-row");check,_:=gtk.CheckButtonNewWithLabel(n.Text);check.SetActive(n.Done);check.Connect("toggled",func(){for i:=range notes{if notes[i].ID==n.ID{notes[i].Done=check.GetActive()}};save()});row.PackStart(check,true,true,0);remove,_:=gtk.ButtonNewWithLabel("×");remove.SetTooltipText("Notiz löschen");remove.Connect("clicked",func(){for i:=range notes{if notes[i].ID==n.ID{notes=append(notes[:i],notes[i+1:]...);break}};save();render()});row.PackEnd(remove,false,false,0);list.PackStart(row,false,false,0)};list.ShowAll()
 }
 input,_:=gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL,8);entry,_:=gtk.EntryNew();entry.SetPlaceholderText("Neue Notiz …");entry.SetWidthChars(12);add,_:=gtk.ButtonNewWithLabel("+");menuClass(add,"demo-note-add");commit:=func(){text,_:=entry.GetText();if text==""{return};notes=append(notes,widgetNote{ID:fmt.Sprint(time.Now().UnixNano()),Text:text});save();entry.SetText("");render()};entry.Connect("activate",commit);add.Connect("clicked",commit);input.PackStart(entry,true,true,0);input.PackEnd(add,false,false,0);shell.PackEnd(input,false,false,0);render()
}
