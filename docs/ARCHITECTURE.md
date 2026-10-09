# Architektur

Der Desktop ist ein Go-Programm mit GTK3. Mehrere Startargumente aktivieren Teilanwendungen: `--files`, `--settings`, `--portal`, `--login`, `--spotify-overlay` und `--install-system`. Sway verwaltet Fenster, Layer-Shell-Flächen und Tastenkürzel; die Anpassungen unter compositor liefern gemeinsame Fensterränder.

## Dateien und Standardprogramme

Die systemweiten Desktop-Einträge und `/etc/xdg/mimeapps.list` registrieren Dateien als Handler für `inode/directory`. `miracleos.portal` plus D-Bus-Aktivierungsdatei registrieren den vorhandenen FileChooser; `sway-portals.conf` wählt ihn aus. Der Explorer und der Speichern-Dialog teilen die GTK-Implementierung.

## Einstellungen und Anmeldung

Private Einstellungen liegen in `~/.config/miracleos/settings.json`. Nur Hell/Dunkel und Akzent werden in `/var/lib/miracleos/appearance/<username>/settings.json` veröffentlicht. Der Greeter liest diese Datei, nicht private Home-Dateien. Die Kontoverzeichnisse werden bei Einrichtung angelegt; Benutzer können nur ihre eigene Design-Datei schreiben.

## Spotify

Die offizielle App bleibt der Player. MPRIS auf dem Sitzungs-D-Bus liefert Metadaten und Wiedergabesteuerung. Das Overlay rendert diese Daten lokal. Einrichtung startet Spotify im eigenen geführten Layout, erkennt Playing und blendet das Fenster zugunsten des Widgets aus. Das Dock startet/fokussiert die echte App; Alt+F startet das Overlay. Es gibt keine Spotify-Web-API-Integration.

## Installation

`build.sh` erzeugt mit Debian live-build eine saubere Live-ISO. `configure-image.sh` registriert Dienste, Programme, Portal, Standardkonfiguration und Kontovorlagen. `system-helper.py` führt privilegierte Installations-/Reset-Aktionen mit Validierung aus. Image-Manifeste verbinden installierten Stand und Reset-Medium.
