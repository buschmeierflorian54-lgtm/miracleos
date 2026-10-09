import importlib.util
import json
import pathlib
import unittest
import tempfile
import socket
from types import SimpleNamespace
from unittest.mock import patch

spec=importlib.util.spec_from_file_location('helper',pathlib.Path(__file__).with_name('system-helper.py'))
helper=importlib.util.module_from_spec(spec);spec.loader.exec_module(helper)

class InstallerSafetyTests(unittest.TestCase):
    def test_profile_copy_skips_runtime_links_and_sockets(self):
        with tempfile.TemporaryDirectory() as directory:
            source=pathlib.Path(directory)/'source';source.mkdir()
            target=pathlib.Path(directory)/'target'
            (source/'prefs').write_text('preserve preferences')
            (source/'SingletonLock').symlink_to('/missing-process-lock')
            (source/'outside').symlink_to('/etc/passwd')
            with socket.socket(socket.AF_UNIX) as runtime:
                runtime.bind(str(source/'SingletonSocket'))
                helper.copy_setup_profile(source,target)
            self.assertEqual((target/'prefs').read_text(),'preserve preferences')
            self.assertEqual({p.name for p in target.iterdir()},{'prefs'})

    def test_setup_retry_reuses_its_partially_created_account(self):
        with tempfile.TemporaryDirectory() as directory:
            state=pathlib.Path(directory);(state/'setup-pending').touch()
            accounts={}
            def lookup(name):
                if name not in accounts:raise KeyError(name)
                return accounts[name]
            def command(*args,**kwargs):
                if args[0]=='useradd':
                    accounts[args[-1]]=SimpleNamespace(pw_uid=1001,pw_dir='/home/'+args[-1])
                return ''
            with patch.object(helper,'STATE',state),patch.object(helper,'live',return_value=False),patch.dict(helper.os.environ,{'SUDO_USER':'miracle'}),patch.object(helper.pwd,'getpwnam',side_effect=lookup),patch.object(helper,'run',side_effect=command) as commands,patch.object(helper,'finish_setup_account',side_effect=[RuntimeError('copy interrupted'),'done']):
                data={'name':'Test','password':'testing123'}
                with self.assertRaises(RuntimeError):helper.setup(data)
                self.assertEqual(helper.setup(data),'done')
                self.assertEqual(sum(c.args[0]=='useradd' for c in commands.call_args_list),1)

    def test_filters_live_mounted_removable_and_small_disks(self):
        def disk(name,size=32*1024**3,mount=None,rm=False):
            return {'name':name,'size':size,'type':'disk','model':'Test','rm':rm,
                    'children':[{'name':name+'1','type':'part','mountpoints':[mount]}]}
        data={'blockdevices':[disk('/dev/sda'),disk('/dev/sdb',mount='/run/live/medium'),
                             disk('/dev/sdc',mount='/'),disk('/dev/sdd',rm=True),
                             disk('/dev/sde',size=1024),disk('/dev/sdf',mount='/home')]}
        with patch.object(helper,'run',return_value=json.dumps(data)):
            self.assertEqual([d['path'] for d in helper.disk_list()],['/dev/sda'])

    def test_install_refuses_running_system_before_any_command(self):
        with patch.object(helper,'live',return_value=False),patch.object(helper,'run') as command:
            with self.assertRaises(RuntimeError):helper.install('/dev/sda')
            command.assert_not_called()

    def test_install_refuses_unlisted_disk_before_wipe(self):
        with patch.object(helper,'live',return_value=True),patch.object(helper,'disk_list',return_value=[]),patch.object(helper,'run') as command:
            with self.assertRaises(RuntimeError):helper.install('/dev/sda')
            command.assert_not_called()

    def test_setup_cannot_run_in_live_session(self):
        with patch.object(helper,'live',return_value=True),patch.object(helper,'run') as command:
            with self.assertRaises(RuntimeError):helper.setup({'name':'Someone','password':'secret123'})
            command.assert_not_called()

    def test_reset_check_never_loads_kernel_or_reboots(self):
        def command(*args,**kwargs):
            if args[0]=='findmnt':return '/dev/sda3'
            if args[0]=='blkid':return 'abcd-1234'
            return ''
        with patch.object(helper,'run',side_effect=command) as commands,patch.object(helper,'reset_medium',return_value=(pathlib.Path('/tmp/iso'),'/dev/sr0')),patch.object(helper.shutil,'which',return_value='/usr/sbin/kexec'):
            helper.reset(True)
            self.assertFalse(any(c.args[0] in ('kexec','systemctl','wipefs') for c in commands.call_args_list))

if __name__=='__main__':unittest.main()
