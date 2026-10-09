#!/usr/bin/env python3
"""Prepare the MiracleOS password-only GTKLock theme, then securely lock Sway."""
import json, os, pathlib, subprocess, pwd, xml.etree.ElementTree as ET
from PIL import Image, ImageFilter, ImageOps

home = pathlib.Path.home()
cache = home / '.cache/miracleos-lock'
cache.mkdir(parents=True, exist_ok=True)
try:
    prefs = json.loads((home / '.config/miracleos/settings.json').read_text())
except (OSError, ValueError):
    prefs = {}
wallpaper = pathlib.Path(prefs.get('wallpaper') or ('/usr/share/backgrounds/miracleos/dark.png' if prefs.get('dark') else '/usr/share/backgrounds/miracleos/light.png'))
background = cache / 'background.png'
try:
    with Image.open(wallpaper) as image:
        image = ImageOps.fit(image.convert('RGB'), (1920, 1080)).filter(ImageFilter.GaussianBlur(28))
        image.save(background)
except OSError:
    Image.new('RGB', (1920, 1080), '#bb83bd').save(background)
accent = prefs.get('accent', '#b8306a')
if len(accent) != 7 or accent[0] != '#' or any(c not in '0123456789abcdefABCDEF' for c in accent[1:]):
    accent = '#b8306a'
ui = ET.fromstring(subprocess.check_output(['gresource', 'extract', '/usr/bin/gtklock', '/gtklock/gtklock.ui']))
def node(id):
    return next(n for n in ui.iter('object') if n.get('id') == id)
def prop(n, name, value):
    p = next((p for p in n.findall('property') if p.get('name') == name), None)
    if p is None: p = ET.SubElement(n, 'property', name=name)
    p.text = value
prop(node('input-field'), 'placeholder-text', 'Bitte Passwort eingeben')
prop(node('input-field'), 'width-request', '300')
prop(node('input-field'), 'secondary-icon-name', '')
for signal in list(node('input-field').findall('signal')):
    if signal.get('name') == 'icon-release': node('input-field').remove(signal)
prop(node('input-label'), 'visible', 'false')
prop(node('input-label'), 'no-show-all', 'true')
prop(node('unlock-button'), 'label', 'Entsperren →')
prop(node('time-box'), 'margin-bottom', '80')
info = node('info-box')
child = ET.SubElement(info, 'child'); avatar = ET.SubElement(child, 'object', {'class':'GtkImage'})
prop(avatar, 'icon-name', 'avatar-default-symbolic'); prop(avatar, 'pixel-size', '72')
child = ET.SubElement(info, 'child'); welcome = ET.SubElement(child, 'object', {'class':'GtkLabel'})
prop(welcome, 'name', 'welcome'); prop(welcome, 'label', 'Willkommen, ' + pwd.getpwuid(os.getuid()).pw_name + '!')
prop(welcome, 'margin-top', '14'); prop(welcome, 'margin-bottom', '24')
layout = cache / 'lock.ui'; ET.ElementTree(ui).write(layout, encoding='utf-8', xml_declaration=True)
style = cache / 'lock.css'
style.write_text('''
window { color: white; background-color: #bd80ae; }
#window-box { margin: 24px; }
#clock-label { font-size: 36px; font-weight: bold; color: white; }
#date-label { font-size: 14px; color: rgba(255,255,255,.8); }
#welcome { font-size: 23px; font-weight: bold; color: white; }
#input-field { border-radius: 24px; padding: 10px 18px; background: rgba(255,240,250,.25); border: 1px solid rgba(255,255,255,.45); color: white; min-height: 26px; }
#unlock-button { border-radius: 16px; padding: 8px 18px; color: white; border: none; background: ACCENT; }
#error-label, #warning-label { color: #ffe4f0; }
'''.replace('ACCENT', accent))
os.execvp('gtklock', ['gtklock', '-s', str(style), '-x', str(layout), '-b', str(background), '-t', '%H:%M', '-D', '%A, %d. %B'])
