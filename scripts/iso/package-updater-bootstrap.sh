#!/bin/sh
# One-time installation for systems made from images without the update service.
set -eu
out=${1:?Usage: package-updater-bootstrap.sh /absolute/output-directory}
here=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
stage=$(mktemp -d)
trap 'rm -rf "$stage"' EXIT
mkdir -p "$out" "$stage/DEBIAN" "$stage/usr/local/lib/miracleos" "$stage/etc/miracleos"
cat > "$stage/DEBIAN/control" <<'EOF'
Package: miracleos-updater
Version: 0.1.20261020.2
Architecture: amd64
Maintainer: MiracleOS Team
Depends: python3 (>= 3.11), openssl, systemd, ca-certificates
Description: Signed MiracleOS updates from the official GitHub repository
EOF
for file in update-manager.py system-helper.py configure-updates.sh; do
 install -m755 "$here/$file" "$stage/usr/local/lib/miracleos/$file"
done
install -m644 "$here/update-public.pem" "$stage/etc/miracleos/update-public.pem"
cat > "$stage/DEBIAN/postinst" <<'EOF'
#!/bin/sh
set -e
if [ "$1" = configure ]; then /usr/local/lib/miracleos/configure-updates.sh; fi
EOF
chmod 755 "$stage/DEBIAN/postinst"
dpkg-deb --root-owner-group --build "$stage" "$out/miracleos-updater_0.1.20261020.2_amd64.deb"
