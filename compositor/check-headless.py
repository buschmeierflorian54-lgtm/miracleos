#!/usr/bin/env python3
"""Isolated compositor smoke test; never targets the active desktop socket."""
import json, os, pathlib, subprocess, tempfile, time

binary = '/home/test/sway-source/sway-1.10.1/build/sway/sway'
with tempfile.TemporaryDirectory(prefix='miracleos-frames-') as temp:
    path = pathlib.Path(temp)
    config = path / 'config'
    config.write_text('default_border normal 6\ndefault_floating_border normal 6\nfont pango:Sans 11\ntitlebar_padding 12 10\nclient.focused #fbedf3 #fbedf3 #3a2033 #b92b68 #fbedf3\nclient.unfocused #fbedf3 #fbedf3 #3a2033 #b92b68 #fbedf3\noutput * bg #ddb1cc solid_color\nfor_window [all] floating enable\n')
    env = dict(os.environ)
    for key in ('WAYLAND_DISPLAY', 'DISPLAY', 'SWAYSOCK'):
        env.pop(key, None)
    socket = str(path / 'ipc.sock')
    env.update(HOME=temp, WLR_BACKENDS='headless', WLR_HEADLESS_OUTPUTS='1',
               WLR_RENDERER='pixman', SWAYSOCK=socket)
    with open(path / 'log', 'w') as log:
        process = subprocess.Popen([binary, '--unsupported-gpu', '-c', str(config)], env=env, stdout=log, stderr=log)
        try:
            def ipc(*args):
                return subprocess.check_output(['swaymsg', '-s', socket, '-r', *args], text=True)
            for _ in range(200):
                if pathlib.Path(socket).exists(): break
                if process.poll() is not None: raise RuntimeError((path / 'log').read_text())
                time.sleep(.1)
            ipc('exec xterm -fa Monospace -fs 11 -title MiracleOS-frame-smoke')
            def windows(node):
                return ([node] if node.get('pid') or node.get('window') else []) + [v for k in ('nodes', 'floating_nodes') for n in node.get(k, []) for v in windows(n)]
            for _ in range(200):
                matches = [w for w in windows(json.loads(ipc('-t', 'get_tree'))) if w.get('window_properties', {}).get('class') == 'XTerm']
                if matches: break
                time.sleep(.1)
            assert matches, ipc('-t', 'get_tree') + (path / 'log').read_text()
            time.sleep(.6)
            w = [v for v in windows(json.loads(ipc('-t', 'get_tree'))) if v['id'] == matches[0]['id']][0]
            assert w['border'] == 'normal', w
            assert w['deco_rect']['height'] >= 30, w
            def window():
                return [v for v in windows(json.loads(ipc('-t', 'get_tree'))) if v['id'] == w['id']][0]
            def click_control(index):
                current = window()
                r = current['rect']
                wide = r['width'] >= 480
                x = r['x'] + r['width'] - 114 + index * 38 + 19
                y = current['deco_rect']['y'] + current['deco_rect']['height'] // 2
                ipc(f'seat seat0 cursor set {x} {y}')
                ipc('seat seat0 cursor press button1')
                ipc('seat seat0 cursor release button1')
                time.sleep(.3)
            original = dict(w['rect'])
            click_control(1)
            assert window()['rect']['width'] > original['width'], window()
            click_control(1)
            assert window()['rect'] == original, window()
            click_control(0)
            assert window()['visible'] is False, window()
            ipc(f'[con_id={w["id"]}] scratchpad show')
            time.sleep(.3)
            assert window()['visible'] is True
            ipc(f'[con_id={w["id"]}] fullscreen enable')
            time.sleep(.2)
            f = [v for v in windows(json.loads(ipc('-t', 'get_tree'))) if v['id'] == w['id']][0]
            assert f['fullscreen_mode'] != 0
            ipc(f'[con_id={w["id"]}] fullscreen disable')
            click_control(2)
            assert not [v for v in windows(json.loads(ipc('-t', 'get_tree'))) if v['id'] == w['id']]
            gtk_client = path / 'gtk-csd.py'
            gtk_client.write_text('''import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk
w = Gtk.Window(title="MiracleOS GTK CSD smoke")
w.set_default_size(560,300)
h = Gtk.HeaderBar(title="Client-owned CSD")
h.set_show_close_button(True)
w.set_titlebar(h)
w.add(Gtk.Label(label="Client content remains intact"))
w.connect("destroy", Gtk.main_quit)
w.show_all()
Gtk.main()
''')
            ipc(f'exec env GDK_BACKEND=wayland python3 {gtk_client}')
            for _ in range(100):
                matches = [v for v in windows(json.loads(ipc('-t', 'get_tree'))) if v.get('shell') == 'xdg_shell']
                if matches: break
                time.sleep(.1)
            assert matches, ipc('-t', 'get_tree') + (path / 'log').read_text()
            time.sleep(.6)
            w = matches[0]
            w = window()
            assert w['shell'] == 'xdg_shell' and w['border'] == 'normal' and w['deco_rect']['height'] >= 30, w
            ipc('exec grim /home/test/sway-source/header-v2-preview.png')
            time.sleep(.4)
            original = dict(w['rect'])
            click_control(1)
            assert window()['rect']['width'] > original['width']
            click_control(1)
            assert window()['rect'] == original
            click_control(0)
            assert window()['visible'] is False
            ipc(f'[con_id={w["id"]}] scratchpad show')
            time.sleep(.3)
            ipc(f'[con_id={w["id"]}] resize set 380 300')
            time.sleep(.4)
            original = dict(window()['rect'])
            click_control(1)
            assert window()['rect']['width'] > original['width']
            click_control(1)
            assert window()['rect'] == original
            click_control(2)
            gtk_client.write_text('''import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk
w=Gtk.Window(title="Hoobastank – The Reason")
w.set_default_size(560,400)
css=Gtk.CssProvider()
css.load_from_data(b"window { background: #000; }")
w.get_style_context().add_provider(css,800)
w.connect("destroy",Gtk.main_quit)
w.show_all()
Gtk.main()
''')
            ipc(f'exec env GDK_BACKEND=wayland GTK_CSD=0 python3 {gtk_client}')
            for _ in range(100):
                previews=[v for v in windows(json.loads(ipc('-t','get_tree'))) if v.get('name')=='Hoobastank – The Reason']
                if previews: break
                time.sleep(.1)
            assert previews
            time.sleep(.6)
            ipc('exec grim /home/test/sway-source/header-v2-reference-preview.png')
            time.sleep(.5)

            from PIL import Image
            rules = path / '.config/miracleos/app-frames.tsv'
            rules.parent.mkdir(parents=True, exist_ok=True)
            def frame_pixel():
                shot = path / 'frame.png'
                if shot.exists(): shot.unlink()
                ipc(f'exec grim {shot}')
                for _ in range(100):
                    if shot.exists() and shot.stat().st_size: break
                    time.sleep(.05)
                time.sleep(.2)
                rect = previews[0]['rect']
                return Image.open(shot).convert('RGB').getpixel((rect['x']+2,rect['y']+80))
            baseline = frame_pixel()
            rules.write_text('title\tOther app\t#000000\t#ffffff\n')
            ipc('reload'); time.sleep(.3)
            assert frame_pixel() == baseline, 'Unrelated app changed'
            rules.write_text('title\tHoobastank\t#000000\t#ffffff\n')
            ipc('reload'); time.sleep(.3)
            assert frame_pixel() == (0,0,0), 'App frame did not become black'
            rules.write_text('')
            ipc('reload'); time.sleep(.3)
            assert frame_pixel() == baseline, 'Global frame was not restored'
            ipc(f'[con_id={previews[0]["id"]}] kill')
            print('PASS: isolated compositor; Xwayland and Wayland GTK CSD headers; real button maximize/restore, minimize/show, close; fullscreen round trip')
        finally:
            process.terminate()
            try: process.wait(3)
            except subprocess.TimeoutExpired: process.kill(); process.wait()
