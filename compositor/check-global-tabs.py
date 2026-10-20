#!/usr/bin/env python3
"""Exercise real compositor clicks and window grouping in an isolated session."""
import json, os, pathlib, subprocess, tempfile, time

binary = os.environ.get('MIRACLEOS_TEST_SWAY', '/home/builder/global-tabs-sway/build-miracleos/sway/sway')
broker_binary = os.environ.get('MIRACLEOS_TEST_TABS', '/home/builder/miracleos-app-tabs')
with tempfile.TemporaryDirectory(prefix='miracleos-global-tabs-') as temp:
    path = pathlib.Path(temp)
    runtime = path / 'runtime'; runtime.mkdir(mode=0o700)
    apps = path / '.local/share/applications'; apps.mkdir(parents=True)
    (apps/'luna-tab-test.desktop').write_text('[Desktop Entry]\nType=Application\nName=Tab Test\nExec=xterm -class LunaTabTest -title Additional-window\nStartupWMClass=LunaTabTest\n')
    config = path/'config'
    config.write_text('default_border normal 6\ndefault_floating_border normal 6\nfont pango:Sans 11\ntitlebar_padding 12 10\nfor_window [all] floating enable, resize set 900 500\noutput * resolution 1280x800\n')
    env = dict(os.environ)
    for key in ('WAYLAND_DISPLAY','DISPLAY','SWAYSOCK'): env.pop(key,None)
    socket = str(path/'ipc.sock')
    env.update(HOME=temp,XDG_RUNTIME_DIR=str(runtime),SWAYSOCK=socket,WLR_BACKENDS='headless',WLR_HEADLESS_OUTPUTS='1',WLR_RENDERER='pixman',MIRACLEOS_GLOBAL_TABS='1')
    log = open(path/'sway.log','w')
    process = subprocess.Popen([binary,'--unsupported-gpu','-c',str(config)],env=env,stdout=log,stderr=log)
    broker = None
    def ipc(*args): return json.loads(subprocess.check_output(['swaymsg','-s',socket,'-r',*args],text=True))
    def windows(n): return ([n] if n.get('pid') or n.get('window') else [])+[w for k in ('nodes','floating_nodes') for c in n.get(k,[]) for w in windows(c)]
    def current(): return [w for w in windows(ipc('-t','get_tree')) if w.get('window_properties',{}).get('class')=='LunaTabTest']
    def wait(predicate,description):
        deadline=time.monotonic()+15
        while time.monotonic()<deadline:
            if predicate(): return
            if process.poll() is not None: raise AssertionError((path/'sway.log').read_text())
            time.sleep(.15)
        raise AssertionError(description+'\n'+str(current())+'\n'+(path/'broker.log').read_text()+'\n'+(path/'sway.log').read_text()[-3000:])
    def click(w,x):
        r=w['rect']; d=w['deco_rect']
        ipc(f'seat seat0 cursor set {r["x"]+int(x)} {d["y"]+max(8,d["height"]//2)}')
        ipc('seat seat0 cursor press button1');ipc('seat seat0 cursor release button1')
    try:
        wait(lambda:pathlib.Path(socket).exists(),'compositor startup')
        blog=open(path/'broker.log','w')
        ipc('exec env > '+str(path/'session-env'))
        wait(lambda:(path/'session-env').exists(),'session environment')
        for line in (path/'session-env').read_text().splitlines():
            key,_,value=line.partition('=')
            if key in ('DISPLAY','WAYLAND_DISPLAY'): env[key]=value
        broker=subprocess.Popen([broker_binary],env=env,stdout=blog,stderr=blog)
        ipc('exec xterm -class LunaTabTest -title First-window')
        wait(lambda:len(current())==1,'first window')
        first=current()[0]['id']
        state=lambda:runtime/'miracleos-app-tabs'/f'{first}.json'
        wait(lambda:state().exists(),'first tab snapshot')
        assert json.loads(state().read_text())['can_new']
        assert json.loads(state().read_text())['name']=='Tab Test'
        # IPC cursor warps do not dispatch pointer_motion; drag is checked
        # separately with the test VM's real virtual mouse input.
        w=current()[0];ipc('exec grim /home/builder/global-tabs-before.png');time.sleep(1);click(w,236)
        wait(lambda:len(current())==2,'plus opens a second real window')
        wait(lambda:sum(w.get('visible',False) for w in current())==1,'only active tab visible')
        wait(lambda:len(json.loads(state().read_text())['tabs'])==2,'two tab labels')
        visible=next(w for w in current() if w.get('visible'))
        # Previous tab arrow, then close the active tab through its per-tab cross.
        click(visible,70)
        wait(lambda:any(w['id']==first and w.get('visible') for w in current()),'tab switch')
        w=next(w for w in current() if w.get('visible'))
        s=json.loads(state().read_text());index=next(i for i,t in enumerate(s['tabs']) if t['id']==s['active'])
        cell=min(220,(w['rect']['width']-114-32)//len(s['tabs']))
        click(w,index*cell+cell-12)
        wait(lambda:len(current())==1 and current()[0].get('visible'),'close active tab restores remaining window')
        out=pathlib.Path('/home/builder/global-tabs-smoke.png')
        ipc('exec grim '+str(out));time.sleep(.5)
        assert broker.poll() is None
        print('PASS: native frame + button, real second window, one visible tab, switch, close and restore')
    finally:
        if broker: broker.terminate();broker.wait(timeout=5)
        process.terminate();process.wait(timeout=10);log.close()
