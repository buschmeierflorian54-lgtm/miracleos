/* MiracleOS compositor-owned title decoration. GPL-2.0-or-later (Sway). */
#include <cairo.h>
#include <pango/pangocairo.h>
#include <wlr/interfaces/wlr_buffer.h>
#include <wlr/types/wlr_buffer.h>
static struct border_colors *container_get_current_colors(struct sway_container *con);

#include <sys/stat.h>
#include <strings.h>
#include <stdio.h>
#include "sway/tree/view.h"

// Reload the small, user-owned palette only when the atomically replaced file changes.
static void miracleos_app_colors(struct sway_container *con, struct border_colors *colors) {
 struct rule { char kind[8], identity[256]; unsigned int color, text; };
 static struct rule rules[512]; static int count;
 static struct timespec modified; static ino_t inode;
 char path[4096]; const char *home=getenv("HOME"); if(!home || !con->view)return;
 snprintf(path,sizeof(path),"%s/.config/miracleos/app-frames.tsv",home);
 struct stat st;
 if(stat(path,&st)!=0) {count=0;inode=0;return;}
 if(st.st_ino!=inode || st.st_mtim.tv_sec!=modified.tv_sec || st.st_mtim.tv_nsec!=modified.tv_nsec) {
  count=0;inode=st.st_ino;modified=st.st_mtim;
  FILE *file=fopen(path,"r"); if(file) {
   char line[1024];
   while(count<512 && fgets(line,sizeof(line),file)) {
    struct rule *r=&rules[count];
    if(sscanf(line,"%7[^\t]\t%255[^\t]\t#%6x\t#%6x",r->kind,r->identity,&r->color,&r->text)==4)count++;
   }
   fclose(file);
  }
 }
 const char *app=view_get_app_id(con->view), *class=view_get_class(con->view);
 for(int i=0;i<count;i++) {
  struct rule *r=&rules[i];
  bool match=(!strcmp(r->kind,"app") && app && !strcmp(app,r->identity)) ||
   (!strcmp(r->kind,"class") && class && !strcasecmp(class,r->identity)) ||
   (!strcmp(r->kind,"title") && con->title && !strncmp(con->title,r->identity,strlen(r->identity)));
  if(!match)continue;
  for(int c=0;c<4;c++) {
   float value=c==3?1.0f:((r->color >> (16-8*c))&255)/255.0f;
   colors->background[c]=colors->border[c]=colors->child_border[c]=value;
   colors->text[c]=c==3?1.0f:((r->text >> (16-8*c))&255)/255.0f;
  }
  return;
 }
}
struct miracleos_header_buffer {
	struct wlr_buffer base;
	cairo_surface_t *surface;
};
static void miracleos_header_destroy(struct wlr_buffer *base) {
	struct miracleos_header_buffer *b = wl_container_of(base, b, base);
	cairo_surface_destroy(b->surface); free(b);
}
static bool miracleos_header_access(struct wlr_buffer *base, uint32_t flags,
		void **data, uint32_t *format, size_t *stride) {
	struct miracleos_header_buffer *b = wl_container_of(base, b, base);
	*data = cairo_image_surface_get_data(b->surface);
	*format = DRM_FORMAT_ARGB8888;
	*stride = cairo_image_surface_get_stride(b->surface); return true;
}
static void miracleos_header_end(struct wlr_buffer *base) { }
static const struct wlr_buffer_impl miracleos_header_impl = {
	.destroy = miracleos_header_destroy, .begin_data_ptr_access = miracleos_header_access,
	.end_data_ptr_access = miracleos_header_end,
};
static void miracleos_color(cairo_t *cr, float c[4]) {
	cairo_set_source_rgba(cr, c[0], c[1], c[2], c[3]);
}
static void miracleos_round(cairo_t *cr, double x, double y, double w, double h, double r) {
	cairo_new_sub_path(cr); cairo_arc(cr,x+w-r,y+r,r,-1.5707963,0);
	cairo_arc(cr,x+w-r,y+h-r,r,0,1.5707963);
	cairo_arc(cr,x+r,y+h-r,r,1.5707963,3.1415926);
	cairo_arc(cr,x+r,y+r,r,3.1415926,4.7123889); cairo_close_path(cr);
}
static void miracleos_tab(cairo_t *cr, double x, double width, double height, bool left) {
	const double r=15, join=12;
	cairo_new_path(cr);
	cairo_move_to(cr,x+r,0);
	cairo_line_to(cr,x+width-r,0);
	cairo_curve_to(cr,x+width,0,x+width,r/2,x+width,r);
	if(left) {
		cairo_line_to(cr,x+width,height-join);
		cairo_curve_to(cr,x+width,height,x+width+join/2,height,x+width+join,height);
	} else cairo_line_to(cr,x+width,height);
	cairo_line_to(cr,left ? x : x-join,height);
	if(!left) cairo_curve_to(cr,x-join/2,height,x,height,x,height-join);
	cairo_line_to(cr,x,r);
	cairo_curve_to(cr,x,r/2,x,0,x+r,0);
	cairo_close_path(cr);
}
static void miracleos_text(cairo_t *cr, const char *text, double x, double y,
		int width, bool bold) {
	PangoLayout *layout = pango_cairo_create_layout(cr);
	PangoFontDescription *font = pango_font_description_from_string(bold ? "Sans Semi-Bold 10" : "Sans 9");
	pango_layout_set_font_description(layout,font);
	pango_layout_set_text(layout,text ? text : "",-1);
	pango_layout_set_width(layout,width*PANGO_SCALE);
	pango_layout_set_ellipsize(layout,PANGO_ELLIPSIZE_END);
	cairo_move_to(cr,x,y); pango_cairo_show_layout(cr,layout);
	pango_font_description_free(font); g_object_unref(layout);
}
static void miracleos_icon(cairo_t *cr, int kind, double x, double y) {
	cairo_set_line_width(cr,1.6); cairo_set_line_cap(cr,CAIRO_LINE_CAP_ROUND);
	if(kind==0) { // down chevron and baseline
		cairo_move_to(cr,x+3,y+4); cairo_line_to(cr,x+8,y+9); cairo_line_to(cr,x+13,y+4);
		cairo_move_to(cr,x+3,y+13); cairo_line_to(cr,x+13,y+13);
	} else if(kind==1) {
		cairo_move_to(cr,x+3,y+3); cairo_line_to(cr,x+7,y+3); cairo_move_to(cr,x+3,y+3); cairo_line_to(cr,x+3,y+7);
		cairo_move_to(cr,x+3,y+3); cairo_line_to(cr,x+7,y+7);
		cairo_move_to(cr,x+13,y+13); cairo_line_to(cr,x+9,y+13); cairo_move_to(cr,x+13,y+13); cairo_line_to(cr,x+13,y+9);
		cairo_move_to(cr,x+13,y+13); cairo_line_to(cr,x+9,y+9);
	} else { cairo_move_to(cr,x+4,y+4); cairo_line_to(cr,x+12,y+12); cairo_move_to(cr,x+12,y+4); cairo_line_to(cr,x+4,y+12); }
	cairo_stroke(cr);
}
static void miracleos_render_header(struct sway_container *con, struct border_colors *colors) {
	if(!con->view || con->title_width<150) return;
 struct border_colors app_colors=*colors; miracleos_app_colors(con,&app_colors); colors=&app_colors;
	int width=con->title_width, height=container_titlebar_height();
	if(height<24 || width>16384) return;
	struct miracleos_header_buffer *b=calloc(1,sizeof(*b)); if(!b) return;
	b->surface=cairo_image_surface_create(CAIRO_FORMAT_ARGB32,width,height);
	if(cairo_surface_status(b->surface)!=CAIRO_STATUS_SUCCESS) {cairo_surface_destroy(b->surface);free(b);return;}
	cairo_t *cr=cairo_create(b->surface);
	int right=114, left=width-right-16;
	if(left>300) left=300;
	if(left<100) left=100;
	// Transparent middle with independent rounded tabs, joining the frame at bottom.
	miracleos_color(cr,colors->background);
	miracleos_tab(cr,0,left,height,true); cairo_fill(cr);
	miracleos_tab(cr,width-right,right,height,false); cairo_fill(cr);
	cairo_rectangle(cr,0,height-3,width,3); cairo_fill(cr);
	miracleos_color(cr,colors->text);
	// A restrained application glyph; actual title is never treated as markup.
	miracleos_round(cr,12,(height-20)/2,20,20,5); cairo_set_line_width(cr,1.4);cairo_stroke(cr);
	cairo_move_to(cr,16,(height-20)/2+6);cairo_line_to(cr,28,(height-20)/2+6);cairo_stroke(cr);
	miracleos_text(cr,con->formatted_title,40,(height-17)/2,left-52,true);
	
	int position=width-right;
	for(int i=0;i<3;i++) {
		int size=38;
		if(i==2) {
			miracleos_color(cr,colors->indicator);
			miracleos_round(cr,position+5,5,size-10,height-10,12);cairo_fill(cr);
			cairo_set_source_rgba(cr,1,1,1,1);
		} else miracleos_color(cr,colors->text);
		int icon_x=position+11;
		miracleos_icon(cr,i,icon_x,(height-16)/2);
		
		position+=size;
	}
	cairo_destroy(cr);cairo_surface_flush(b->surface);
	wlr_buffer_init(&b->base,&miracleos_header_impl,width,height);
	if(con->title_bar.miracleos_buffer) wlr_scene_node_destroy(&con->title_bar.miracleos_buffer->node);
	con->title_bar.miracleos_buffer=wlr_scene_buffer_create(con->title_bar.tree,&b->base);
	wlr_buffer_drop(&b->base);
	wlr_scene_node_set_enabled(&con->title_bar.background->node,false);
	wlr_scene_node_set_enabled(&con->title_bar.border->node,false);
	if(con->title_bar.title_text) wlr_scene_node_set_enabled(con->title_bar.title_text->node,false);
	if(con->title_bar.marks_text) wlr_scene_node_set_enabled(con->title_bar.marks_text->node,false);
	for(int i=0;i<3;i++) if(con->title_bar.controls[i]) wlr_scene_node_set_enabled(con->title_bar.controls[i]->node,false);
}

void miracleos_render_frame(struct sway_container *con, int width, int height, int top, int border) {
	if(!con->view || con->current.border!=B_NORMAL || border<6 || width<150 || height<top+32) {
		if(con->miracleos_frame) {
			wlr_scene_node_destroy(&con->miracleos_frame->node);
			con->miracleos_frame=NULL;
		}
		wlr_scene_node_set_enabled(&con->border.bottom->node,true);
		wlr_scene_node_set_enabled(&con->border.left->node,true);
		wlr_scene_node_set_enabled(&con->border.right->node,true);
		return;
	}
	struct border_colors app_colors=*container_get_current_colors(con);
 miracleos_app_colors(con,&app_colors); struct border_colors *colors=&app_colors;
	if(con->miracleos_frame && con->miracleos_frame_width==width && con->miracleos_frame_height==height && con->miracleos_frame_top==top && con->miracleos_frame_border==border &&
			memcmp(con->miracleos_frame_color,colors->child_border,sizeof(con->miracleos_frame_color))==0) goto hide_stock;
	struct miracleos_header_buffer *b=calloc(1,sizeof(*b)); if(!b)return;
	b->surface=cairo_image_surface_create(CAIRO_FORMAT_ARGB32,width,height);
	if(cairo_surface_status(b->surface)!=CAIRO_STATUS_SUCCESS) {cairo_surface_destroy(b->surface);free(b);return;}
	cairo_t *cr=cairo_create(b->surface);
	miracleos_color(cr,colors->child_border);
	cairo_move_to(cr,0,top);cairo_line_to(cr,width,top);cairo_line_to(cr,width,height-16);
	cairo_curve_to(cr,width,height-4,width-4,height,width-16,height);
	cairo_line_to(cr,16,height);cairo_curve_to(cr,4,height,0,height-4,0,height-16);
	cairo_close_path(cr);cairo_fill(cr);
	// Clear the exact client rectangle. Its square corners remain untouched and
	// fit inside a 16px outer corner with at least 6px frame padding.
	cairo_set_operator(cr,CAIRO_OPERATOR_CLEAR);
	cairo_rectangle(cr,border,top,width-border*2,height-top-border);cairo_fill(cr);
	cairo_destroy(cr);cairo_surface_flush(b->surface);
	wlr_buffer_init(&b->base,&miracleos_header_impl,width,height);
	if(con->miracleos_frame)wlr_scene_node_destroy(&con->miracleos_frame->node);
	con->miracleos_frame=wlr_scene_buffer_create(con->border.tree,&b->base);
	wlr_scene_node_lower_to_bottom(&con->miracleos_frame->node);
	wlr_buffer_drop(&b->base);
	con->miracleos_frame_width=width;con->miracleos_frame_height=height;
	con->miracleos_frame_top=top;con->miracleos_frame_border=border;
	memcpy(con->miracleos_frame_color,colors->child_border,sizeof(con->miracleos_frame_color));
hide_stock:
	wlr_scene_node_set_enabled(&con->border.bottom->node,false);
	wlr_scene_node_set_enabled(&con->border.left->node,false);
	wlr_scene_node_set_enabled(&con->border.right->node,false);
}
