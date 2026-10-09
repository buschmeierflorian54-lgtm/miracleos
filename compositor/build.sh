#!/bin/sh
set -eu
here=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
source_dir=${1:?Usage: build.sh /path/to/pristine/sway-1.10.1}
python3 "$here/patch-source.py" "$source_dir"
meson setup "$source_dir/build-miracleos" "$source_dir" --buildtype=release -Dman-pages=disabled
ninja -C "$source_dir/build-miracleos" sway/sway
"$source_dir/build-miracleos/sway/sway" --version
