#!/usr/bin/env python3
"""Apply MiracleOS central window controls to pristine Sway 1.10.1."""
from pathlib import Path
import sys

root = Path(sys.argv[1])
def edit(name, old, new):
    p = root / name
    text = p.read_text()
    if old not in text:
        raise SystemExit(f'Unsupported source at {name}: {old[:60]!r}')
    position = text.rfind(old) if name == 'sway/input/seatop_default.c' and old.startswith('\t// Handle beginning floating move') else text.find(old)
    p.write_text(text[:position] + new + text[position + len(old):])

edit('meson.build', "add_project_arguments('-DSWAY_VERSION=@0@'.format(version), language: 'c')", "version = '\"@0@-miracleos.2\"'.format(meson.project_version())\nadd_project_arguments('-DSWAY_VERSION=@0@'.format(version), language: 'c')")
edit('include/sway/tree/container.h', 'struct sway_text_node *marks_text;', '''struct sway_text_node *marks_text;
		struct sway_text_node *controls[3];''')
edit('include/sway/tree/container.h', 'struct sway_text_node *controls[3];', 'struct sway_text_node *controls[3];\n\t\tstruct wlr_scene_buffer *miracleos_buffer;')
(root / 'sway/miracleos-header.h').write_text(Path(__file__).with_name('miracleos-header.h').read_text())
edit('sway/tree/container.c', '#include "stringop.h"', '#include "stringop.h"\n#include "../miracleos-header.h"')
edit('sway/tree/container.c', 'sway_text_node_set_background(con->title_bar.marks_text, colors->background);\n\t}\n}', 'sway_text_node_set_background(con->title_bar.marks_text, colors->background);\n\t}\n\tmiracleos_render_header(con, colors);\n}')
edit('include/sway/tree/container.h', 'bool is_sticky;', '''bool is_sticky;
	struct wlr_scene_buffer *miracleos_frame;
	int miracleos_frame_width, miracleos_frame_height, miracleos_frame_top, miracleos_frame_border;
	float miracleos_frame_color[4];
	bool miracleos_maximized;
	double miracleos_restore_x, miracleos_restore_y;
	double miracleos_restore_width, miracleos_restore_height;''')
edit('include/sway/tree/container.h', 'void container_arrange_title_bar(struct sway_container *con);', 'void container_arrange_title_bar(struct sway_container *con);\nvoid miracleos_render_frame(struct sway_container *con, int width, int height, int top, int border);')
edit('sway/desktop/transaction.c', 'border_left, border_top);\n\t} else {', 'border_left, border_top);\n\t\tmiracleos_render_frame(con, width, height, border_top, border_width);\n\t} else {')
edit('sway/tree/container.c', '\tif (con->title_bar.marks_text) {\n\t\tsway_text_node_set_color', '''	for (int i = 0; i < 3; i++) {
		if (con->title_bar.controls[i]) {
			sway_text_node_set_color(con->title_bar.controls[i], colors->text);
			sway_text_node_set_background(con->title_bar.controls[i], colors->background);
		}
	}
	if (con->title_bar.marks_text) {
		sway_text_node_set_color''')
edit('sway/tree/container.c', '\tint height = container_titlebar_height();', '''	int height = container_titlebar_height();
	int controls_width = con->view ? 114 : 0;''')
edit('sway/tree/container.c', 'width - h_padding - config->titlebar_h_padding);\n\t\talloc_width = MAX(alloc_width, 0);\n\n\t\tsway_text_node_set_max_width(node, alloc_width);\n\t\twlr_scene_node_set_position(node->node,\n\t\t\th_padding, (height - node->height) >> 1);\n\n\t\tpixman_region32_union_rect(&text_area, &text_area,\n\t\t\tnode->node->x, node->node->y, alloc_width, node->height);\n\t}\n\n\t// silence', '''width - h_padding - config->titlebar_h_padding - controls_width);
		alloc_width = MAX(alloc_width, 0);

		sway_text_node_set_max_width(node, alloc_width);
		wlr_scene_node_set_position(node->node,
			h_padding, (height - node->height) >> 1);

		pixman_region32_union_rect(&text_area, &text_area,
			node->node->x, node->node->y, alloc_width, node->height);
	}
	for (int i = 0; i < 3; i++) {
		struct sway_text_node *control = con->title_bar.controls[i];
		if (!control) continue;
		wlr_scene_node_set_enabled(control->node, width >= 150);
		wlr_scene_node_set_position(control->node,
			width - (3 - i) * 38 + (38 - control->width) / 2,
			(height - control->height) / 2);
	}

	// silence''')
edit('sway/tree/container.c', '\tstruct border_colors *colors = container_get_current_colors(con);\n\n\tif (con->title_bar.title_text)', '''	struct border_colors *colors = container_get_current_colors(con);
	char *symbols[] = {"−", "□", "×"};
	for (int i = 0; i < 3; i++) {
		if (con->title_bar.controls[i]) wlr_scene_node_destroy(con->title_bar.controls[i]->node);
		con->title_bar.controls[i] = con->view ? sway_text_node_create(
			con->title_bar.tree, symbols[i], colors->text, false) : NULL;
	}

	if (con->title_bar.title_text)''')
# A compositor-owned outer title bar also covers clients that draw their own
# decoration. Client content is retained; we never crop arbitrary app pixels.
edit('sway/tree/view.c', '\t\tif (container_is_floating(con)) {\n\t\t\tcon->pending.border = B_CSD;\n\t\t}', '\t\tcon->pending.border = B_NORMAL;')
edit('sway/tree/container.c', 'container->pending.border = B_CSD;', 'container->pending.border = B_NORMAL;')
edit('sway/tree/container.c', 'WLR_XDG_TOPLEVEL_DECORATION_V1_MODE_CLIENT_SIDE);', 'WLR_XDG_TOPLEVEL_DECORATION_V1_MODE_SERVER_SIDE);')
edit('sway/xdg_decoration.c', '\tif (floating && client_mode) {\n\t\tmode = client_mode;\n\t}', '\t(void)floating; // MiracleOS consistently negotiates server-side decorations.')
edit('sway/xdg_decoration.c', 'view_update_csd_from_client(view, csd);', 'view_update_csd_from_client(view, false);\n\t\t(void)csd;')
edit('sway/input/seatop_default.c', '#include "sway/tree/view.h"', '#include "sway/tree/view.h"\n#include "sway/tree/root.h"\n#include "sway/tree/arrange.h"')
edit('sway/input/seatop_default.c', '\t// Handle beginning floating move', '''	// MiracleOS controls belong to each compositor container, never an app class.
	if (cont && cont->view && on_titlebar && !is_fullscreen_or_child &&
			button == BTN_LEFT && state == WL_POINTER_BUTTON_STATE_PRESSED) {
		double control_x = cursor->cursor->x - cont->current.x;
		int width = cont->title_width;
		int controls_width = 114;
		if (width >= 150 && control_x >= width - controls_width && control_x < width) {
			double local_x = control_x - (width - controls_width);
			int control = (int)(local_x / 38);
			if (control == 2) {
				view_close(cont->view);
			} else if (control == 0) {
				if (cont->scratchpad) root_scratchpad_hide(cont);
				else root_scratchpad_add_container(cont, NULL);
			} else if (container_is_floating(cont) && cont->pending.workspace) {
				struct sway_workspace *ws = cont->pending.workspace;
				if (!cont->miracleos_maximized) {
					cont->miracleos_restore_x = cont->pending.x;
					cont->miracleos_restore_y = cont->pending.y;
					cont->miracleos_restore_width = cont->pending.width;
					cont->miracleos_restore_height = cont->pending.height;
					cont->pending.x = ws->x; cont->pending.y = ws->y;
					cont->pending.width = ws->width; cont->pending.height = ws->height;
				} else {
					cont->pending.x = cont->miracleos_restore_x;
					cont->pending.y = cont->miracleos_restore_y;
					cont->pending.width = cont->miracleos_restore_width;
					cont->pending.height = cont->miracleos_restore_height;
				}
				cont->miracleos_maximized = !cont->miracleos_maximized;
				arrange_container(cont);
			}
			transaction_commit_dirty();
			return;
		}
	}

	// Handle beginning floating move''')

# Global frame tabs: rendering and click handling share the same geometry.
(root / 'sway/miracleos-tabs.h').write_text(Path(__file__).with_name('miracleos-tabs.h').read_text())
edit('include/sway/tree/container.h', 'void container_arrange_title_bar(struct sway_container *con);', 'bool miracleos_tab_click(struct sway_container *con, double x);\nvoid container_arrange_title_bar(struct sway_container *con);')
edit('sway/input/seatop_default.c', '\t\tint width = cont->title_width;', '\t\tif (!mod_pressed && miracleos_tab_click(cont, control_x)) return;\n\t\tint width = cont->title_width;')