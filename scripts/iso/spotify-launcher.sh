#!/bin/sh
# Use native Wayland in MiracleOS; preserve URI arguments and X11 fallback.
if [ -n "${WAYLAND_DISPLAY:-}" ]; then
  exec /usr/bin/spotify --ozone-platform=wayland "$@"
fi
exec /usr/bin/spotify "$@"
