/* Shared frame-tab layout for rendering and hit testing. */
#include <json-c/json.h>
#include "sway/ipc-server.h"
struct miracleos_tabs {
 char name[256], ids[256][64], titles[256][256];
 int count, active, first, shown, start, end, cell, nav;
 bool can_new;
};
static bool miracleos_read_tabs(struct sway_container *con, struct miracleos_tabs *s) {
 memset(s,0,sizeof(*s)); const char *runtime=getenv("XDG_RUNTIME_DIR"); if(!runtime)return false;
 char path[4096];snprintf(path,sizeof(path),"%s/miracleos-app-tabs/%zu.json",runtime,con->node.id);
 struct stat st;if(stat(path,&st)||st.st_size>65536)return false;
 json_object *root=json_object_from_file(path),*v=NULL,*list=NULL;if(!root)return false;
 if(json_object_object_get_ex(root,"name",&v))snprintf(s->name,sizeof(s->name),"%s",json_object_get_string(v));
 if(json_object_object_get_ex(root,"can_new",&v))s->can_new=json_object_get_boolean(v);
 const char *selected="";if(json_object_object_get_ex(root,"active",&v))selected=json_object_get_string(v);
 if(json_object_object_get_ex(root,"tabs",&list)&&json_object_is_type(list,json_type_array)) {
  int count=json_object_array_length(list);if(count>256)count=256;
  for(int i=0;i<count;i++) {json_object *item=json_object_array_get_idx(list,i),*id=NULL,*title=NULL;
   if(!json_object_object_get_ex(item,"id",&id)||!json_object_object_get_ex(item,"title",&title))continue;
   const char *key=json_object_get_string(id);if(!key||strlen(key)>63||strspn(key,"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-")!=strlen(key))continue;
   int n=s->count++;snprintf(s->ids[n],64,"%s",key);snprintf(s->titles[n],256,"%s",json_object_get_string(title));if(!strcmp(key,selected))s->active=n;
  }
 }
 json_object_put(root);if(!s->count)return false;
 s->start=0;int room=con->title_width-114-32;
 s->nav=s->count*80>room?20:0;room-=s->nav*2;
 s->shown=room/80;if(s->shown<1)return false;if(s->shown>s->count)s->shown=s->count;
 s->first=(s->active/s->shown)*s->shown;if(s->first+s->shown>s->count)s->first=s->count-s->shown;
 s->cell=room/s->shown;if(s->cell>220)s->cell=220;
 s->end=s->start+s->nav*2+s->shown*s->cell;return true;
}
bool miracleos_tab_click(struct sway_container *con,double x) {
 struct miracleos_tabs s;if(!miracleos_read_tabs(con,&s)||x<s.start||x>=s.end+32)return false;
 char event[128];const char *action="select",*id=NULL;
 if(x>=s.end){if(!s.can_new)return true;action="new";id="";}
 else if(x<s.start+s.nav)id=s.ids[(s.active+s.count-1)%s.count];
 else if(x>=s.end-s.nav)id=s.ids[(s.active+1)%s.count];
 else {int index=(x-s.start-s.nav)/s.cell;if(index>=s.shown)return true;int local=(int)(x-s.start-s.nav)%s.cell;id=s.ids[s.first+index];if(local>=s.cell-22)action="close";}
 // The active tab doubles as the draggable title grip. Leave its press to
 // Sway's normal move operation, without an asynchronous geometry reset.
 if(!strcmp(action,"select")&&!strcmp(id,s.ids[s.active]))return false;
 snprintf(event,sizeof(event),"miracleos-tab:%s:%s",action,id);ipc_event_window(con,event);return true;
}
