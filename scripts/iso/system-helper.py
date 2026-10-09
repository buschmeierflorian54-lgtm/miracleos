#!/usr/bin/python3
"""Privileged ISO install/setup/reset entry point. JSON arrives through stdin."""
import fcntl, json, os, pathlib, pwd, re, shutil, subprocess, sys, time

STATE = pathlib.Path('/var/lib/miracleos')
IMAGE = pathlib.Path('/etc/miracleos-image.json')
LOG = pathlib.Path('/var/log/miracleos-install.log')

def run(*args, **kwargs):
    result = subprocess.run(args, text=True, stdout=subprocess.PIPE,
                            stderr=subprocess.STDOUT, **kwargs)
    if result.returncode:
        raise RuntimeError(f'{args[0]} failed: {result.stdout[-1800:]}')
    return result.stdout.strip()

def live():
    return pathlib.Path('/run/live/medium/live/filesystem.squashfs').is_file()

def flatten(nodes):
    for node in nodes:
        yield node
        yield from flatten(node.get('children', []))

def disk_list():
    tree = json.loads(run('lsblk','--json','--bytes','--paths','-o','NAME,TYPE,SIZE,MODEL,RM,MOUNTPOINTS'))
    choices = []
    for d in tree['blockdevices']:
        if d['type'] != 'disk' or d['size'] < 16*1024**3 or d.get('rm'):
            continue
        if any(any(m
                       for m in n.get('mountpoints', [])) for n in flatten([d])):
            continue
        choices.append({'path': d['name'], 'label': f"{d['name']} · {(d.get('model') or 'Disk').strip()} · {d['size']/1024**3:.1f} GiB"})
    return choices

def install(disk, automatic=False):
    if not live():
        raise RuntimeError('Die Installation ist nur von der MiracleOS-ISO aus möglich.')
    allowed = {d['path'] for d in disk_list()}
    disk = os.path.realpath(disk)
    if disk not in allowed:
        raise RuntimeError('Das gewählte Laufwerk ist nicht verfügbar oder enthält das Live-System.')
    root = pathlib.Path('/mnt/miracleos-target')
    root.mkdir(exist_ok=True)
    mounted = []
    LOG.write_text(f'Installing MiracleOS to {disk}\n')
    def execute(*args, **kwargs):
        result = run(*args, **kwargs)
        with LOG.open('a') as log: log.write(' '.join(args) + '\n' + result + '\n')
        return result
    # No shell expressions and no user-defined commands enter privileged execution.
    execute('wipefs','--all',disk)
    execute('parted','--script',disk,'mklabel','gpt','mkpart','BIOS','1MiB','3MiB',
            'set','1','bios_grub','on','mkpart','EFI','fat32','3MiB','515MiB',
            'set','2','esp','on','mkpart','MiracleOS','ext4','515MiB','100%')
    execute('partprobe',disk);execute('udevadm','settle')
    suffix = 'p' if disk[-1].isdigit() else ''
    efi, system = disk+suffix+'2', disk+suffix+'3'
    for _ in range(20):
        if pathlib.Path(system).exists():break
        time.sleep(.25)
    execute('mkfs.vfat','-F','32','-n','MIRACLE_EFI',efi)
    execute('mkfs.ext4','-F','-L','MiracleOS',system)
    try:
        execute('mount',system,str(root));mounted.append(root)
        source = pathlib.Path('/run/live/rootfs/filesystem.squashfs')
        if not source.is_dir():raise RuntimeError('Live-Dateisystem nicht gefunden.')
        execute('rsync','-aHAX','--numeric-ids','--exclude=/proc/*','--exclude=/sys/*',
                '--exclude=/dev/*','--exclude=/run/*','--exclude=/tmp/*',
                str(source)+'/',str(root)+'/')
        (root/'boot/efi').mkdir(parents=True,exist_ok=True)
        execute('mount',efi,str(root/'boot/efi'));mounted.append(root/'boot/efi')
        for name in ['dev','proc','sys','run']:
            (root/name).mkdir(exist_ok=True)
            execute('mount','--rbind','/'+name,str(root/name))
            execute('mount','--make-rslave',str(root/name));mounted.append(root/name)
        root_uuid=execute('blkid','-s','UUID','-o','value',system)
        efi_uuid=execute('blkid','-s','UUID','-o','value',efi)
        (root/'etc/fstab').write_text(f'UUID={root_uuid} / ext4 defaults 0 1\nUUID={efi_uuid} /boot/efi vfat umask=0077 0 2\n')
        (root/'etc/hostname').write_text('miracleos\n')
        (root/'etc/hosts').write_text('127.0.0.1 localhost\n127.0.1.1 miracleos\n::1 localhost ip6-localhost ip6-loopback\n')
        (root/'etc/machine-id').write_text('')
        (root/'var/lib/miracleos').mkdir(parents=True,exist_ok=True)
        (root/'var/lib/miracleos/setup-pending').touch()
        # Live-only services and automatic root privileges never survive installation.
        execute('chroot',str(root),'systemctl','disable','miracleos-reset-live.service')
        execute('chroot',str(root),'update-initramfs','-u','-k','all')
        execute('chroot',str(root),'grub-install','--target=i386-pc','--recheck',disk)
        execute('chroot',str(root),'grub-install','--target=x86_64-efi','--efi-directory=/boot/efi',
                '--bootloader-id=MiracleOS','--removable','--no-nvram')
        execute('chroot',str(root),'update-grub')
        execute('sync')
    finally:
        for mount in reversed(mounted):
            subprocess.run(['umount','-R',str(mount)],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
    if automatic:run('systemctl','reboot')
    return 'Installation abgeschlossen.'

def copy_setup_profile(source, target):
    # Chromium's Singleton* entries and sockets are runtime state, not profile
    # data. Never dereference links while copying a user's files as root.
    def ignore(directory, names):
        return [name for name in names
                if (pathlib.Path(directory)/name).is_symlink()
                or not ((pathlib.Path(directory)/name).is_dir()
                        or (pathlib.Path(directory)/name).is_file())]
    shutil.copytree(source, target, symlinks=True, ignore=ignore, dirs_exist_ok=True)

def setup(data):
    if live() or not (STATE/'setup-pending').exists():
        raise RuntimeError('Die Ersteinrichtung ist bereits abgeschlossen oder das System ist nicht installiert.')
    if os.environ.get('SUDO_USER') not in ('miracle',None):
        raise RuntimeError('Ersteinrichtung nur für das Einrichtungskonto.')
    name=data.get('name','').strip();password=data.get('password','')
    if not name or len(name)>80 or any(c in name for c in ':\n\r\x00') or len(password)<8 or any(c in password for c in '\n\r\x00'):
        raise RuntimeError('Ungültiger Name oder Passwort.')
    import unicodedata
    username=re.sub('[^a-z0-9]','',unicodedata.normalize('NFKD',name).encode('ascii','ignore').decode().lower())[:24] or 'benutzer'
    pending_account=STATE/'setup-account.json'
    if pending_account.exists():
        previous=json.loads(pending_account.read_text())
        if previous['name'] != name:
            raise RuntimeError('Bitte die begonnene Einrichtung mit demselben Namen fortsetzen.')
        username=previous['username']
        account=pwd.getpwnam(username)
        if account.pw_uid < 1000 or account.pw_dir != '/home/'+username:
            raise RuntimeError('Ungültiges Einrichtungskonto.')
    else:
        base=username;number=1
        while True:
            try:pwd.getpwnam(username)
            except KeyError:break
            number+=1;username=base+str(number)
        run('useradd','--create-home','--shell','/bin/bash','--comment',name,'--groups','sudo,audio,video,input',username)
        pending_account.write_text(json.dumps({'username':username,'name':name}))
        pending_account.chmod(0o644) # Name/account only; contains no password.
    run('chpasswd',input=username+':'+password+'\n')
    return finish_setup_account(username, name)

def finish_setup_account(username, name):
    pending_account=STATE/'setup-account.json'
    if not (STATE/'setup-pending').exists() or json.loads(pending_account.read_text()) != {'username':username,'name':name}:
        raise RuntimeError('Keine passende begonnene Einrichtung vorhanden.')
    subprocess.run(['pkill','-u','miracle','-x','spotify'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
    deadline=time.monotonic()+8
    while any(len(parts)==2 and parts[1]=='spotify' and not parts[0].startswith('Z')
              for parts in (line.split() for line in run('ps','-u','miracle','-o','stat=,comm=').splitlines())):
        if time.monotonic()>deadline:
            raise RuntimeError('Spotify wird noch beendet. Bitte erneut versuchen.')
        time.sleep(.1)
    account=pwd.getpwnam(username)
    for relative in ['.config/spotify','.cache/spotify','.local/share/keyrings']:
        source=pathlib.Path('/home/miracle')/relative
        target=pathlib.Path(account.pw_dir)/relative
        if source.is_dir():
            target.parent.mkdir(parents=True,exist_ok=True)
            copy_setup_profile(source,target)
    run('chown','-hR',f'{account.pw_uid}:{account.pw_gid}',account.pw_dir)
    appearance=STATE/'appearance'/username
    appearance.mkdir(parents=True, exist_ok=True)
    appearance.chmod(0o755)
    os.chown(appearance,account.pw_uid,account.pw_gid)
    pathlib.Path('/etc/miracleos-account.json').write_text(json.dumps({'username':username,'name':name}))
    config=pathlib.Path('/etc/greetd/config.toml')
    config.write_text('[terminal]\nvt = 1\n[default_session]\ncommand = "/usr/local/bin/miracleos-greeter"\nuser = "_greetd"\n')
    (STATE/'setup-pending').unlink()
    pending_account.unlink()
    run('systemd-run','--on-active=4','--unit=miracleos-finish-setup','systemctl','restart','greetd')
    return 'Konto eingerichtet.'

def reset_medium():
    if live():raise RuntimeError('Zurücksetzen ist nur im installierten System möglich.')
    directory=pathlib.Path('/run/miracleos-reset-media');directory.mkdir(exist_ok=True)
    expected=json.loads(IMAGE.read_text())['image_id']
    candidates=sorted(pathlib.Path('/dev').glob('sr*'))
    labelled=pathlib.Path('/dev/disk/by-label/MIRACLEOS')
    if labelled.exists():candidates.append(labelled)
    for device in candidates:
        if subprocess.run(['mount','-o','ro',str(device),str(directory)],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL).returncode:continue
        try:
            if json.loads((directory/'miracleos-image.json').read_text())['image_id']==expected:
                if all((directory/p).is_file() for p in ['live/vmlinuz','live/initrd.img','live/filesystem.squashfs']):
                    return directory,os.path.realpath(device)
        except (OSError,ValueError,KeyError):pass
        finally:
            # Successful callers use the mount until kexec has loaded the kernel.
            pass
        run('umount',str(directory))
    raise RuntimeError('Bitte die passende MiracleOS-ISO einlegen. Es wurde nichts gelöscht.')

def reset(check=False):
    source=run('findmnt','-n','-o','SOURCE','/')
    uuid=run('blkid','-s','UUID','-o','value',source)
    if not re.fullmatch('[a-fA-F0-9-]+',uuid):raise RuntimeError('Systemlaufwerk nicht eindeutig ermittelt.')
    directory,device=reset_medium()
    try:
        if not shutil.which('kexec'):raise RuntimeError('Der ISO-Neustart ist auf diesem System nicht verfügbar.')
        if check:return 'Passwort und ISO geprüft.'
        cmdline=f'boot=live components username=miracle live-media={device} miracleos.reset={uuid} noeject quiet'
        run('kexec','--load',str(directory/'live/vmlinuz'),'--initrd='+str(directory/'live/initrd.img'),'--command-line='+cmdline)
    finally:
        run('umount',str(directory))
    run('systemctl','kexec')
    return 'Neuinstallation wird gestartet.'

def auto_reset():
    match=re.search(r'(?:^|\s)miracleos.reset=([a-fA-F0-9-]+)(?:\s|$)',pathlib.Path('/proc/cmdline').read_text())
    if not match:return 'Kein Reset angefordert.'
    root=pathlib.Path('/dev/disk/by-uuid')/match[1]
    if not root.exists():raise RuntimeError('Reset-Ziellaufwerk nicht gefunden; nichts gelöscht.')
    parent=run('lsblk','-no','PKNAME',os.path.realpath(root)).strip()
    if not re.fullmatch('[a-zA-Z0-9_-]+',parent):raise RuntimeError('Reset-Ziel nicht eindeutig.')
    return install('/dev/'+parent,True)

def main():
    if os.geteuid()!=0:raise RuntimeError('Administratorrechte erforderlich.')
    action=sys.argv[1] if len(sys.argv)>1 else ''
    if action=='auto-reset':return auto_reset()
    data=json.load(sys.stdin)
    if action=='disks':
        if not live():raise RuntimeError('Bitte von der MiracleOS-ISO starten.')
        return json.dumps(disk_list())
    if action=='install':
        if not isinstance(data,dict) or data.get('confirmation')!='ERASE':raise RuntimeError('Löschbestätigung fehlt.')
        return install(data.get('disk',''))
    if action=='setup':return setup(data)
    if action in ('check-reset','reset'):
        user=os.environ.get('SUDO_USER','')
        if not user or user=='miracle' or user=='root':raise RuntimeError('Bitte als eingerichteter Benutzer anmelden.')
        return reset(action=='check-reset')
    raise RuntimeError('Unbekannte Aktion.')

if __name__=='__main__':
    try:
        with open('/run/miracleos-operation.lock','w') as lock:
            fcntl.flock(lock,fcntl.LOCK_EX|fcntl.LOCK_NB)
            print(main())
    except Exception as error:
        print(str(error),file=sys.stderr);sys.exit(1)
