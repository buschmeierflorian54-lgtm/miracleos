import pathlib,subprocess
root=pathlib.Path('/etc/greetd')
(root/'config.toml').write_text('''[terminal]
vt = 1
[default_session]
command = "env WLR_RENDERER=pixman WLR_NO_HARDWARE_CURSORS=1 sway --unsupported-gpu --config /etc/greetd/sway-config"
user = "_greetd"
''')
(root/'sway-config').write_text('''output * mode 1280x800
output * bg /usr/share/backgrounds/miracleos/dark.png fill
input type:keyboard xkb_layout de
exec "gtkgreet -l -c /usr/local/bin/miracleos-session -b /usr/share/backgrounds/miracleos/dark.png -s /etc/greetd/miracleos.css; swaymsg exit"
''')
(root/'miracleos.css').write_text('''
window { color: #fff0fa; background: transparent; font-family: sans-serif; }
box { border-radius: 22px; }
entry { min-width: 280px; padding: 12px 18px; border-radius: 24px; background: rgba(255,235,247,.25); border: 1px solid rgba(255,255,255,.5); color: white; }
button { border-radius: 16px; padding: 10px 20px; background: #b8306a; color: white; border: none; }
label { color: white; font-size: 16px; }
''')
session=pathlib.Path('/usr/local/bin/miracleos-session')
session.write_text('''#!/bin/sh
export GTK_USE_PORTAL=1 XDG_CURRENT_DESKTOP=sway XDG_SESSION_TYPE=wayland
export WLR_RENDERER=pixman WLR_NO_HARDWARE_CURSORS=1
export DBUS_SESSION_BUS_ADDRESS="unix:path=$XDG_RUNTIME_DIR/bus"
exec sway --unsupported-gpu > "$HOME/sway.log" 2>&1
''')
session.chmod(0o755)
auto=pathlib.Path('/etc/systemd/system/getty@tty1.service.d/autologin.conf')
if auto.exists():auto.rename(auto.with_suffix('.conf.disabled'))
subprocess.run(['systemctl','mask','getty@tty1.service'],check=True)
subprocess.run(['systemctl','enable','greetd'],check=True)
subprocess.run(['systemctl','daemon-reload'],check=True)
print('Password login configured for next boot; existing desktop preserved.')
