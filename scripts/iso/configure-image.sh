#!/bin/sh
set -eu
export DEBIAN_FRONTEND=noninteractive
mkdir -p /etc/apt/keyrings /etc/miracleos /var/lib/miracleos /etc/skel/.config/sway /usr/local/bin
mkdir -p /etc/skel/.config/miracleos /etc/systemd/user
# Spotify is supplied as its official signed Debian package, with no login baked in.
curl -fsSL https://download.spotify.com/debian/pubkey_5384CE82BA52C83A.asc -o /tmp/spotify-key.asc
gpg --dearmor --yes -o /etc/apt/keyrings/spotify.gpg /tmp/spotify-key.asc
echo 'deb [signed-by=/etc/apt/keyrings/spotify.gpg] https://repository.spotify.com stable non-free' > /etc/apt/sources.list.d/spotify.list
apt-get update
apt-get install -y spotify-client
install -m 644 /usr/share/spotify/spotify.desktop /usr/share/applications/spotify.desktop
rm /tmp/spotify-key.asc
ln -sf /opt/miracleos/miracleos-desktop /usr/local/bin/miracleos-desktop
ln -sf /opt/miracleos/miracleos-desktop /usr/local/bin/miracleos-login
cat > /usr/share/applications/miracleos-settings.desktop <<'EOF'
[Desktop Entry]
Type=Application
Name=Einstellungen
Exec=miracleos-desktop --settings
Icon=preferences-system
Terminal=false
Categories=Settings;
EOF
cat > /usr/share/applications/miracleos-files.desktop <<'EOF'
[Desktop Entry]
Type=Application
Name=Dateien
Exec=miracleos-desktop --files %f
Icon=system-file-manager
Terminal=false
Categories=System;FileManager;
MimeType=inode/directory;
StartupWMClass=miracleos-files
EOF
mkdir -p /etc/xdg
cat > /etc/xdg/mimeapps.list <<'EOF'
[Default Applications]
inode/directory=miracleos-files.desktop;
EOF
# Register the existing MiracleOS file picker for sandboxed and GTK applications.
mkdir -p /usr/share/xdg-desktop-portal/portals /usr/share/dbus-1/services /etc/xdg/xdg-desktop-portal
cat > /usr/share/xdg-desktop-portal/portals/miracleos.portal <<'EOF'
[portal]
DBusName=org.freedesktop.impl.portal.desktop.miracleos
Interfaces=org.freedesktop.impl.portal.FileChooser;
UseIn=sway;MiracleOS;
EOF
cat > /usr/share/dbus-1/services/org.freedesktop.impl.portal.desktop.miracleos.service <<'EOF'
[D-BUS Service]
Name=org.freedesktop.impl.portal.desktop.miracleos
Exec=/usr/local/bin/miracleos-desktop --portal
EOF
cat > /etc/xdg/xdg-desktop-portal/sway-portals.conf <<'EOF'
[preferred]
default=wlr;
org.freedesktop.impl.portal.FileChooser=miracleos;
org.freedesktop.impl.portal.Secret=gnome-keyring;
EOF
cat > /usr/local/bin/miracleos-session <<'EOF'
#!/bin/sh
export GTK_USE_PORTAL=1 XDG_CURRENT_DESKTOP=sway XDG_SESSION_TYPE=wayland
export WLR_NO_HARDWARE_CURSORS=1
export MIRACLEOS_SERVER_DECORATIONS=1 QT_WAYLAND_DISABLE_WINDOWDECORATION=1 GTK_CSD=0
export DBUS_SESSION_BUS_ADDRESS="unix:path=$XDG_RUNTIME_DIR/bus"
if systemd-detect-virt --quiet; then export WLR_RENDERER=pixman; fi
exec /usr/local/libexec/miracleos-sway --unsupported-gpu --config "$HOME/.config/sway/config"
EOF
cat > /usr/local/bin/miracleos-greeter <<'EOF'
#!/bin/sh
export WLR_NO_HARDWARE_CURSORS=1
if systemd-detect-virt --quiet; then export WLR_RENDERER=pixman; fi
exec dbus-run-session sway --unsupported-gpu --config /etc/miracleos/greeter-config
EOF
cat > /usr/local/bin/miracleos-start <<'EOF'
#!/bin/sh
dbus-update-activation-environment --systemd WAYLAND_DISPLAY DISPLAY SWAYSOCK XDG_CURRENT_DESKTOP XDG_SESSION_TYPE
systemctl --user import-environment WAYLAND_DISPLAY DISPLAY SWAYSOCK XDG_CURRENT_DESKTOP XDG_SESSION_TYPE DBUS_SESSION_BUS_ADDRESS GTK_USE_PORTAL MIRACLEOS_SERVER_DECORATIONS QT_WAYLAND_DISABLE_WINDOWDECORATION GTK_CSD
systemctl --user start miracleos-bar.service
if [ -f /run/live/medium/live/filesystem.squashfs ]; then
  exec miracleos-desktop --install-system
fi
exit 0
EOF
chmod 755 /usr/local/bin/miracleos-session /usr/local/bin/miracleos-greeter /usr/local/bin/miracleos-start
cat > /etc/miracleos/sway-config <<'EOF'
set $mod Mod4
font pango:DejaVu Sans 10
output * bg /usr/share/backgrounds/miracleos/light.png fill
input type:keyboard xkb_layout de
default_border normal 6
default_floating_border normal 6
gaps inner 8
bindsym $mod+Return exec miracleos-desktop --terminal
bindsym $mod+e exec miracleos-desktop --files
bindsym $mod+b exec firefox-esr
bindsym $mod+Shift+q kill
bindsym $mod+1 workspace number 1
bindsym $mod+2 workspace number 2
bindsym $mod+3 workspace number 3
bindsym Mod1+f exec miracleos-desktop --spotify-overlay
floating_modifier $mod normal
for_window [app_id=".*"] floating enable
for_window [class=".*"] floating enable
for_window [title="^Ersteinrichtung"] fullscreen enable
exec /usr/bin/lxpolkit
exec /usr/local/bin/miracleos-start
include ~/.config/miracleos/window-theme.conf
EOF
cp /etc/miracleos/sway-config /etc/skel/.config/sway/config
cat > /etc/skel/.config/miracleos/window-theme.conf <<'EOF'
font pango:DejaVu Sans 12
titlebar_padding 12 10
titlebar_border_thickness 0
client.focused #edd5e6 #edd5e6 #352737 #bf69a4 #edd5e6
client.unfocused #edd5e6 #edd5e6 #352737 #bf69a4 #edd5e6
EOF
cat > /etc/skel/.config/miracleos/desktop-layout.json <<'EOF'
{"version":1,"pins":[{"ID":"miracleos-files.desktop","Name":"Dateien","Icon":"system-file-manager","File":"/usr/share/applications/miracleos-files.desktop","WMClass":"miracleos-files"},{"ID":"miracleos-settings.desktop","Name":"Einstellungen","Icon":"preferences-system","File":"/usr/share/applications/miracleos-settings.desktop","WMClass":"miracleos-settings"},{"ID":"spotify.desktop","Name":"Spotify","Icon":"spotify-client","File":"/usr/share/applications/spotify.desktop","WMClass":"Spotify"}],"items":[{"id":"welcome-clock","kind":"clock","surface":"start","x":12,"y":12,"width":280,"height":210},{"id":"welcome-note","kind":"note","surface":"start","x":310,"y":12,"width":270,"height":200,"text":"Dein Platz.\nRechtsklick zum Hinzufügen.\nAlt + F: Spotify-Steuerung."}]}
EOF
cat > /etc/systemd/user/miracleos-bar.service <<'EOF'
[Unit]
Description=MiracleOS desktop and taskbar
StartLimitIntervalSec=0
[Service]
Type=simple
ExecStart=/opt/miracleos/miracleos-desktop
Restart=always
RestartSec=2
Environment=GDK_BACKEND=wayland
KillMode=process
EOF
cat > /etc/miracleos/greeter-config <<'EOF'
output * bg /usr/share/backgrounds/miracleos/light.png fill
input type:keyboard xkb_layout de
exec "/usr/local/bin/miracleos-login --login; swaymsg exit"
EOF
cat > /etc/greetd/config.toml <<'EOF'
[terminal]
vt = 1
[default_session]
command = "/usr/local/bin/miracleos-greeter"
user = "_greetd"
[initial_session]
command = "/usr/local/bin/miracleos-session"
user = "miracle"
EOF
if ! id miracle >/dev/null 2>&1; then useradd -m -s /bin/bash -G audio,video,input miracle; fi
passwd -l miracle
# Only onboarding commands can run without authentication. The helper also
# checks the root-owned pending marker or verifies it is running from the ISO.
cat > /etc/sudoers.d/miracleos-onboarding <<'EOF'
miracle ALL=(root) NOPASSWD: /usr/local/lib/miracleos/system-helper.py setup, /usr/local/lib/miracleos/system-helper.py disks, /usr/local/lib/miracleos/system-helper.py install
EOF
chmod 440 /etc/sudoers.d/miracleos-onboarding
touch /var/lib/miracleos/setup-pending
cat > /etc/systemd/system/miracleos-reset-live.service <<'EOF'
[Unit]
Description=Reinstall MiracleOS after authenticated reset from inserted ISO
After=live-config.service
Before=greetd.service
ConditionKernelCommandLine=miracleos.reset
[Service]
Type=oneshot
ExecStart=/usr/local/lib/miracleos/system-helper.py auto-reset
TimeoutStartSec=30min
[Install]
WantedBy=multi-user.target
EOF
systemctl enable greetd NetworkManager miracleos-reset-live.service
systemctl mask getty@tty1.service
systemctl disable ssh.service 2>/dev/null || true
sed -i 's/^# *de_DE.UTF-8 UTF-8/de_DE.UTF-8 UTF-8/' /etc/locale.gen
locale-gen
echo 'LANG=de_DE.UTF-8' > /etc/default/locale
ln -sf /usr/share/zoneinfo/Europe/Berlin /etc/localtime
echo 'Europe/Berlin' > /etc/timezone
echo 'NAME="MiracleOS"' > /etc/os-release
echo 'ID=miracleos' >> /etc/os-release
echo 'ID_LIKE=debian' >> /etc/os-release
echo 'PRETTY_NAME="MiracleOS 0.1 (Debian 13)"' >> /etc/os-release
echo 'VERSION_ID="0.1"' >> /etc/os-release
echo 'GRUB_DISTRIBUTOR=MiracleOS' >> /etc/default/grub
echo 'GRUB_CMDLINE_LINUX_DEFAULT="quiet"' >> /etc/default/grub
# Setup and ISO images contain neither machine identity nor build credentials.
rm -f /etc/ssh/ssh_host_* /var/lib/dbus/machine-id
truncate -s 0 /etc/machine-id
