#!/usr/bin/env python3
"""Prepare user configuration for the compositor installed by compositor/install.sh.
Does not reload Sway or close applications. Activated on next desktop login.
"""
import pathlib
home=pathlib.Path.home();config=home/'.config/sway/config';backup=config.with_name('config.before-central-frames')
text=config.read_text()
if not backup.exists():backup.write_text(text)
text=text.replace('default_border pixel 2','default_border normal 6').replace('default_floating_border normal 2','default_floating_border normal 6').replace('border none','border normal 6').replace('border normal 1','border normal 6')
theme=home/'.config/miracleos/window-theme.conf';theme.parent.mkdir(parents=True,exist_ok=True)
if not theme.exists():theme.write_text('font pango:DejaVu Sans 12\ntitlebar_padding 12 10\ntitlebar_border_thickness 0\n')
include=f'include {theme}'
if include not in text:text+='\n# Shared MiracleOS window palette\n'+include+'\n'
config.write_text(text)
launcher=home/'.local/bin/miracleos-start-bar'
s=launcher.read_text()
if 'MIRACLEOS_SERVER_DECORATIONS' not in s:s=s.replace('XDG_SESSION_TYPE GTK_USE_PORTAL','XDG_SESSION_TYPE GTK_USE_PORTAL MIRACLEOS_SERVER_DECORATIONS QT_WAYLAND_DISABLE_WINDOWDECORATION GTK_CSD')
launcher.write_text(s)
print('Central window configuration prepared for the next session; backup retained.')
