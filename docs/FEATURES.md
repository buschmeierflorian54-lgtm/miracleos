# Desktop
The main desktop environment

Start by `go run -v .` or by downloading an executable in the github actions
## Native terminal and applications

The desktop uses GIO to discover launchers in the standard XDG application directories. The start menu and Settings Apps share a live, searchable catalog. Packages without an application launcher are not shown as desktop apps.

The integrated terminal uses GTK 3 and VTE with a real local shell. Build dependencies now include `libvte-2.91-dev` and `libglib2.0-dev`; the runtime needs `libvte-2.91-0` and `gio`. On Debian: `sudo apt install libvte-2.91-dev libglib2.0-dev libglib2.0-bin`.

Launch it with `miracleos-desktop --terminal`. A Terminal application entry is registered in the user's XDG applications directory. Copy/paste shortcuts are Ctrl+Shift+C and Ctrl+Shift+V.

The Start menu offers poweroff, reboot, and password-only session locking.
Install runtime dependencies with: sudo apt install gtklock python3-pil
runtime/lockScreen.py prepares the current wallpaper and accent for GTKLock; PAM validates the Debian account password.
Taskbar buttons use Sway container IDs to focus, restore, and close one specific window. Context colors apply to individual open Sway windows and are discarded when those windows close.

Standard wallpapers are installed as root-owned files in /usr/share/backgrounds/miracleos (light.png and dark.png). Copy cmd/desktop/images/wallpapers with sudo install -o root -g root -m 0644, and create the directory with mode 0755. Theme switching selects the matching default; custom wallpapers remain optional.
Boot login uses greetd with gtkgreet; the previous tty1 autologin is disabled. Install greetd and gtkgreet, then scripts/install/configureBootLogin.py prepares the greeter and MiracleOS session as administrator.

Recent files and favorites use the shared file drag source. Holding a drag still for one second shows cards for open windows. Files, file pickers and Terminal accept drops through a per-user Unix socket bridge; other applications are focused on card hover so the drag can continue into their native drop target. Cards show application icons and window titles, not live thumbnails.

The status capsule opens Samsung-style quick settings with palette-aware rounded tiles. Hardware discovery and commands run outside GTK's main loop. Audio, Wi-Fi and Bluetooth availability are checked against the actual services and devices; brightness requires a writable backlight interface. Unavailable or unimplemented features remain disabled. Dark mode uses the shared preferences and matching wallpaper. Escape or the close button dismisses the panel.

Run python3 scripts/install/installBarRecovery.py in the Debian session to install miracleos-bar.service. Sway starts the service after importing its display environment. The user service restarts the taskbar after two seconds on unexpected exit, keeps crash logs in the journal, and skips startup when the session socket is gone. For updates, use systemctl --user restart miracleos-bar.service; to stop intentionally, use systemctl --user stop miracleos-bar.service.
