#!/usr/bin/env python3
"""Install session-scoped automatic recovery for the MiracleOS taskbar."""
import pathlib,subprocess,os,glob,signal
home=pathlib.Path.home();repo=pathlib.Path(__file__).resolve().parents[2]
unitdir=home/'.config/systemd/user';unitdir.mkdir(parents=True,exist_ok=True)
launcher=home/'.local/bin/miracleos-start-bar';launcher.parent.mkdir(parents=True,exist_ok=True)
launcher.write_text('''#!/bin/sh
systemctl --user import-environment WAYLAND_DISPLAY DISPLAY SWAYSOCK XDG_CURRENT_DESKTOP XDG_SESSION_TYPE GTK_USE_PORTAL MIRACLEOS_SERVER_DECORATIONS QT_WAYLAND_DISABLE_WINDOWDECORATION GTK_CSD
systemctl --user reset-failed miracleos-bar.service
exec systemctl --user restart miracleos-bar.service
''');launcher.chmod(0o755)
(unitdir/'miracleos-bar.service').write_text(f'''[Unit]
Description=MiracleOS taskbar (automatic recovery)
StartLimitIntervalSec=0

[Service]
Type=simple
WorkingDirectory={repo}
ExecCondition=/bin/sh -c 'test -S "$SWAYSOCK"'
ExecStart={repo}/miracleos-desktop
Restart=always
RestartSec=2
TimeoutStopSec=5
KillMode=process
Environment=GDK_BACKEND=wayland
StandardOutput=journal
StandardError=journal
''')
config=home/'.config/sway/config';text=config.read_text();lines=text.splitlines()
lines=[line for line in lines if not (line.startswith('exec ') and ('./miracleos-desktop' in line or 'miracleos-start-bar' in line))]
lines.append(f'exec {launcher}')
config.write_text('\n'.join(lines)+'\n')
subprocess.run(['systemctl','--user','daemon-reload'],check=True)
# Adopt the environment of the actual logged-in Sway session when installing over SSH.
sock=max(glob.glob('/run/user/%d/sway-ipc.*.sock'%os.getuid()),key=os.path.getmtime)
pid=int(pathlib.Path(sock).name.split('.')[2]);env=dict(x.split('=',1) for x in pathlib.Path(f'/proc/{pid}/environ').read_bytes().decode().split('\0') if '=' in x)
os.environ.update(env);os.environ['SWAYSOCK']=sock
wayland=max((p for p in pathlib.Path('/run/user/%d'%os.getuid()).glob('wayland-*') if p.is_socket()),key=lambda p:p.stat().st_mtime)
os.environ['WAYLAND_DISPLAY']=wayland.name
os.environ['XDG_SESSION_TYPE']='wayland'
os.environ['XDG_CURRENT_DESKTOP']='sway'
for f in glob.glob('/proc/[0-9]*/cmdline'):
 try:
  args=pathlib.Path(f).read_bytes().rstrip(b'\0').split(b'\0')
  if args in ([b'./miracleos-desktop'],[b'./miracleos-desktop',b'--welcome']):os.kill(int(f.split('/')[2]),signal.SIGTERM)
 except (OSError,ValueError):pass
subprocess.run([str(launcher)],check=True)
