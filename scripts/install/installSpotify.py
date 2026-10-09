#!/usr/bin/env python3
"""Install Spotify from an already configured, signed APT repository.

Run inside the Debian guest. Does not log in, launch playback, or alter accounts.
"""
from pathlib import Path
import subprocess


def main():
    subprocess.run(['sudo', '-n', 'apt-get', 'install', '-y', 'spotify-client'], check=True)
    entry = Path('/usr/share/spotify/spotify.desktop')
    if not entry.is_file():
        raise SystemExit('Spotify desktop entry missing from the installed package')
    subprocess.run(['sudo', '-n', 'install', '-m', '644', str(entry),
                    '/usr/share/applications/spotify.desktop'], check=True)
    subprocess.run(['dpkg-query', '-W', 'spotify-client'], check=True)
    print('Spotify installed and registered in the application catalog.')


if __name__ == '__main__':
    main()
