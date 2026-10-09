# Spotify on MiracleOS

Spotify is an optional third-party Debian package, separate from the MiracleOS update bundle.
The test VM uses the signed official Spotify APT repository (`repository.spotify.com`, stable/non-free), with its key stored at `/etc/apt/keyrings/spotify.gpg`.

After configuring that repository, run inside the guest:

```sh
python3 scripts/install/installSpotify.py
```

The helper installs `spotify-client` and registers the package's desktop entry in `/usr/share/applications`. The tested package supplied its entry under `/usr/share/spotify/spotify.desktop` without registering it in the application directory. The existing MiracleOS GIO application catalog discovers the registered entry; `spotify-client` icons come from the package.

The quick-settings media card reads the player's MPRIS interface over the current user's session bus. Playback controls follow the player's capability flags. Spotify can report `mpris:length` as an unsigned 64-bit value and `Position` as a signed 64-bit value. Login and account management remain in Spotify; MiracleOS does not store Spotify credentials.

Verified in the Debian test VM: package `1:1.2.95.453.g0eeebbed`, successful window launch, desktop entry registration, and MPRIS metadata/capability availability. No account login or playback commands were performed by the installation check.
