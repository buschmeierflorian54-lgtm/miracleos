#!/usr/bin/env python3
"""Drive Debian passwd over a private PTY; secrets arrive only over stdin."""
import json, os, pty, select, signal, sys, time
data = json.load(sys.stdin)
old, new = data['current'], data['new']
if not old or not new or any(c in old + new for c in '\r\n\x00'):
    sys.exit(2)
pid, fd = pty.fork()
if pid == 0:
    os.environ['LC_ALL'] = 'C'
    os.execv('/usr/bin/passwd', ['passwd'])
buffer = b''
deadline = time.monotonic() + 35
responses = 0
try:
    while time.monotonic() < deadline:
        if select.select([fd], [], [], .2)[0]:
            try: chunk = os.read(fd, 4096)
            except OSError: break
            if not chunk: break
            buffer = (buffer + chunk)[-8192:]
            lower = buffer.lower()
            if lower.rstrip().endswith(b'password:'):
                value = old if b'current' in lower or b'old password' in lower else new
                responses += 1
                if responses > 5: break
                os.write(fd, (value + '\n').encode())
                buffer = b''
        done, status = os.waitpid(pid, os.WNOHANG)
        if done:
            sys.exit(os.waitstatus_to_exitcode(status))
    for _ in range(20):
        done, status = os.waitpid(pid, os.WNOHANG)
        if done: sys.exit(os.waitstatus_to_exitcode(status))
        time.sleep(.05)
    os.kill(pid, signal.SIGTERM)
    os.waitpid(pid, 0)
    sys.exit(1)
finally:
    os.close(fd)
