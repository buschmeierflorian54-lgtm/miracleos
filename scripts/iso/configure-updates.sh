#!/bin/sh
# Run inside the image, or explicitly on an installed test system as root.
set -eu
mkdir -p /var/lib/miracleos/updates /etc/miracleos /opt/miracleos/updates/releases
test -s /etc/miracleos/update-public.pem
test -x /usr/local/lib/miracleos/update-manager.py
cat > /etc/systemd/system/miracleos-update.service <<'EOF'
[Unit]
Description=Signed MiracleOS updates from GitHub
Wants=network-online.target
After=network-online.target
ConditionPathExists=!/run/live/medium/live/filesystem.squashfs
[Service]
Type=oneshot
ExecStart=/usr/local/lib/miracleos/update-manager.py auto
TimeoutStartSec=30min
UMask=0022
NoNewPrivileges=true
ProtectHome=true
ProtectSystem=strict
ReadWritePaths=/opt/miracleos /usr/local/bin /usr/local/libexec /var/lib/miracleos/updates
PrivateTmp=true
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6
EOF
cat > /etc/systemd/system/miracleos-update.timer <<'EOF'
[Unit]
Description=Check MiracleOS updates daily
[Timer]
OnBootSec=10min
OnUnitActiveSec=24h
RandomizedDelaySec=10min
Persistent=true
[Install]
WantedBy=timers.target
EOF
if [ ! -f /var/lib/miracleos/updates/config.json ]; then
 printf '{"enabled":true}\n' > /var/lib/miracleos/updates/config.json
fi
chmod 644 /var/lib/miracleos/updates/config.json
if [ ! -f /var/lib/miracleos/updates/status.json ]; then
 printf '{"state":"ready","installed":"Bisherige Installation","automatic":true,"message":"Automatische Update-Prüfung ist eingerichtet."}\n' > /var/lib/miracleos/updates/status.json
 chmod 644 /var/lib/miracleos/updates/status.json
fi
# Keep the previous fixed-path runtime fallback available for existing launchers.
systemctl enable miracleos-update.timer
if [ -d /run/systemd/system ]; then
 systemctl daemon-reload
 systemctl start miracleos-update.timer
fi
