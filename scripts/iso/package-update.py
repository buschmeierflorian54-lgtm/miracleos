#!/usr/bin/python3
"""Build and sign update assets. The private key must stay outside the repo."""
import argparse
import hashlib
import json
import pathlib
import subprocess
import tarfile

parser = argparse.ArgumentParser(description=__doc__)
for flag in ('desktop', 'luna', 'sway', 'broker', 'key', 'output', 'tag', 'version'):
    parser.add_argument('--'+flag, required=True)
parser.add_argument('--sequence', required=True, type=int)
args = parser.parse_args()
out = pathlib.Path(args.output)
out.mkdir(parents=True, exist_ok=True)
source = pathlib.Path(__file__).resolve().parents[2]
bundle = out / 'miracleos-update-amd64.tar.gz'
with tarfile.open(bundle, 'w:gz') as archive:
    for name, path in [('miracleos-desktop', args.desktop), ('luna-browser', args.luna), ('miracleos-sway', args.sway), ('miracleos-app-tabs', args.broker)]:
        archive.add(path, arcname=name, recursive=False)
    for path in sorted((source / 'runtime').glob('*.py')):
        archive.add(path, arcname='runtime/'+path.name, recursive=False)
with bundle.open('rb') as stream:
    digest = hashlib.file_digest(stream, 'sha256').hexdigest()
manifest = out / 'miracleos-update.json'
manifest.write_text(json.dumps(dict(format=1, repository='buschmeierflorian54-lgtm/miracleos', architecture='amd64', tag=args.tag, sequence=args.sequence, version=args.version, asset=bundle.name, size=bundle.stat().st_size, sha256=digest), sort_keys=True)+'\n')
subprocess.run(['openssl','pkeyutl','-sign','-inkey',args.key,'-rawin','-in',str(manifest),'-out',str(out/'miracleos-update.json.sig')], check=True)
print('Signed update assets:', out)
