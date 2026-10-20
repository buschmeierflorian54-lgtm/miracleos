#!/bin/sh
# Installs a reviewed build; does not restart or change the running compositor.
set -eu
binary=${1:?Usage: sudo install.sh /path/to/patched/sway}
"$binary" --version
install -D -m 0755 "$binary" /usr/local/libexec/miracleos-sway
if [ ! -e /usr/local/bin/miracleos-session.before-central-frames ]; then
    cp -p /usr/local/bin/miracleos-session /usr/local/bin/miracleos-session.before-central-frames
fi
cat > /usr/local/bin/miracleos-session <<'SESSION'
#!/bin/sh
export GTK_USE_PORTAL=1 XDG_CURRENT_DESKTOP=sway XDG_SESSION_TYPE=wayland
export WLR_RENDERER=pixman WLR_NO_HARDWARE_CURSORS=1
export MIRACLEOS_SERVER_DECORATIONS=1
export MIRACLEOS_GLOBAL_TABS=1
export QT_WAYLAND_DISABLE_WINDOWDECORATION=1 GTK_CSD=0
export DBUS_SESSION_BUS_ADDRESS="unix:path=$XDG_RUNTIME_DIR/bus"
exec /usr/local/libexec/miracleos-sway --unsupported-gpu > "$HOME/sway.log" 2>&1
SESSION
chmod 0755 /usr/local/bin/miracleos-session
printf '%s\n' 'Installed. The next desktop login starts MiracleOS Sway. The running session and stock greeter are unchanged.'
