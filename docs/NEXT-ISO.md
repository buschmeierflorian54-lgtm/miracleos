# Merkliste für die nächste MiracleOS-ISO

Stand: 9. Oktober 2026. Nutzer hat den ISO-Build inzwischen ausdrücklich gestartet. Frischer Build aus aktuellem Quellstand auf Builder: /home/builder/desktop-release/scripts/iso/build.sh, Ziel /home/builder/miracleos-release, Log /home/builder/miracleos-release-build.log. Vorhandene veröffentlichte ISO-Dateien enthalten noch nicht alle folgenden Änderungen.

## Für den nächsten Build übernehmen

- Nach Start des aktuellen Builds zusätzlich geändert: Spotify-Overlay-Slider mit 4px abgerundeter Schiene, rosa Verlauf, rundem 12px-Regler und Hover-Ring. desktop.css nachführen; der bereits gestartete Quellsnapshot enthält diese Änderung noch nicht.

- Unser Explorer ist Standard-Dateimanager: systemweite /etc/xdg/mimeapps.list ordnet inode/directory dem miracleos-files.desktop zu; Exec übergibt den Ordner mit %f. Desktop-Datei trägt MimeType und StartupWMClass. In der Test-VM ist diese Zuordnung auch für florian gesetzt und per xdg-mime query bestätigt.
- Laufende Explorer- und Dateiauswahl-Fenster verwenden im Dock unser Dateien-Icon statt des generischen Symbols; Zuordnung in taskbarActions.go anhand App-ID und Fenstertitel. Einstellungen erhalten entsprechend ihr eigenes Icon. Diese letzte Symboländerung muss vor der nächsten VM-Aktualisierung noch gebaut/geprüft werden.
- Alle dauerhaften Systemzuordnungen müssen aus configure-image.sh/build.sh entstehen, nicht nur aus manuellen VM-Anpassungen. Insbesondere frisches Benutzerkonto testen.
  Standard-Dock-Pins in /etc/skel enthalten jetzt Dateien, Einstellungen und Spotify; damit gelten sie auch bei einer frischen Installation.

- Eigenen Explorer als Öffnen-/Speichern-Dateidialog registrieren: miracleos.portal, D-Bus-Aktivierungsdatei und sway-portals.conf in configure-image.sh. Implementierung filePortal.go/filePicker.go existiert bereits; Installation hatte die Registrierungen ausgelassen. Firefox nutzt widget.use-xdg-desktop-portal.file-picker=1, GTK_USE_PORTAL=1 ist in der Sitzung gesetzt. End-to-End-Speichern und Abbrechen prüfen.
  Registrierung in Test-VM installiert, direkter D-Bus-SaveFile-Aufruf zeigte eigenen Explorer mit Dateiname/Speichern/Abbrechen. Firefox-Profil liegt unter ~/.config/mozilla/firefox; user.js enthält Portal-Präferenz. Bereits laufender Firefox muss die Präferenz ggf. erst nach Neustart übernehmen. Tatsächliches Speichern aus Firefox noch prüfen.

- Spotify-Icon in Dock/Taskleiste/Startmenü öffnet bzw. fokussiert die echte Spotify-App. Keine Umleitung zum Overlay. Das Overlay wird ausschließlich per Alt + F aufgerufen. Sonderbehandlung in centerDock.go und taskbarActions.go entfernt.

- Uhr-Widget: Zifferblatt auch bei unterschiedlichen Positionen und Größen zentrieren. Widget-Zusammenführung wurde ausdrücklich verworfen.
- Ersteinrichtung: WLAN-Seite nur bei tatsächlich erkanntem WLAN-Adapter; Netz auswählen und verbinden. Ethernet-only-VM überspringt diesen Schritt. Nur das aktive Eingabefeld erhält RGB-Glow.
- Spotify vorinstallieren. Erst Funktionen erklären, dann Einrichtung anbieten. Echtes Spotify rechts, Einrichtungsbeschreibung links, farbig weicher Hintergrund; keine Desktop-/Einstellungsleisten. Fensterrand mit abgesenkter transparenter Mitte wie normale MiracleOS-Apps. Minimieren deaktiviert. Schließen fragt nach Jetzt/Später einrichten. Hinweis unter dem Fenster: „Bitte den Spotify-QR-Code scannen.“
- Bei erkannter Wiedergabe Spotify ausblenden, Prozess/Musik weiterlaufen lassen und das eigene Musik-Widget zeigen. Vollbild während der geführten Ansicht zurücknehmen. Positionskorrektur und Abschluss auf dem GTK-Thread serialisieren, damit ein verspäteter Poll das Fenster nicht wieder hervorholt. Beim Abschluss auch Vollbild entfernen und Spotify in den Scratchpad verschieben.
- Setup-Abschluss: keine erneute Passworteingabe bei Profilkopierfehler; Wiederholung mit erhaltenen Eingaben. Teilweise angelegtes Konto wiederverwenden. Spotify vor Profilübernahme beenden und abwarten. Sockets/Symlinks nicht als normale Profildateien kopieren. Schlüsselbund-Pakete mitliefern.
- Anmeldebildschirm übernimmt Hell-/Dunkelmodus und Akzentfarbe des eingerichteten Kontos. Setup-Design unverändert lassen. Hintergrund light.png/dark.png verwenden; vorheriger Verweis auf nicht vorhandenes login.png verursachte Weiß-auf-Weiß. Lesbarer Fallback auch ohne Hintergrunddatei.
- Nur öffentliche Designeinstellungen in /var/lib/miracleos/appearance/<username>/settings.json veröffentlichen (dark, accent). Elternverzeichnis root-kontrolliert, Kontoverzeichnis 0755 und dem Konto zugeordnet; private Home-Konfiguration nicht für den Greeter öffnen. Anlage in system-helper.py, Veröffentlichung bei Desktopstart/Änderung in systemTheme.go, Lesen in loginScreen.go.
- Spotify-Overlay: einheitliche symbolische Icons statt Emoji/Textzeichen; runde, kompakte Buttons, erkennbare Play/Pause-Symbole, Hover-/Deaktiviert-Zustände, kontrastierender Hauptknopf.
- Einstellungen und Dateien: systemweite Desktop-Dateien in configure-image.sh ergänzt. Im frischen Konto fehlten zuvor die Startmenü-Einträge. Test-VM-Pins ergänzt; Standard-Pins für neue Konten noch ergänzen/prüfen.
- libglib2.0-bin explizit in Image-Paketliste aufgenommen: App-Start verwendet gio, das in der bisherigen Installation fehlte. desktop-file-utils ggf. ergänzen, falls update-desktop-database im Installationsablauf verwendet wird; dieses Kommando war ebenfalls nicht installiert.

## Teststand und offene Prüfungen

- Go-Tests und sieben Python-Tests bestanden vor den letzten Launcher-/Paketlistenänderungen; zuletzt gebautes UI enthält Theme-Übernahme und Overlay-Icons. Erneut passend testen, bevor eine ISO veröffentlicht wird.
- Test-VM ist MiracleOS-ISO-Test (620fc277-2ae8-41ae-8129-ccf195594d9b). Original MiracleOS-Test unangetastet lassen. Neues Konto florian, UID 1001. Keine Einrichtung zurücksetzen, ohne den inzwischen erreichten Stand zu berücksichtigen.
- Sichtprüfung der Spotify-Einrichtungsansicht mit abgesenktem Rahmen und QR-Hinweis erfolgreich. Eine separat gestartete Vorschau hatte eine Überlagerung verursacht; beim Testen niemals Vorschau und normale Einrichtung gleichzeitig laufen lassen.
- Nach Wiedergabe gemeldetes Vollbild-/Fensterproblem korrigiert; vollständiger Übergang zum Widget mit aktuellem Build noch abschließend prüfen. Spotify-MPRIS meldete auch iPhone-Wiedergabe als Playing.
- Hell/Dunkel und Akzent im angemeldeten Konto ändern, abmelden und Greeter-Kontrast/Übernahme prüfen. Öffentliche Datei wurde in Test-VM mit dark=false und accent=#bf69a4 bestätigt. Keine fremde Sitzung für diesen Test ungefragt beenden.
- Einstellungen/Dateien im Startmenü und Dock sichtbar bestätigt. Startversuch scheiterte an fehlendem gio. Installation von libglib2.0-bin in Test-VM wurde angestoßen, aber durch Unterbrechung noch nicht verifiziert: /tmp/miracleos-gio-install.log prüfen, danach beide Apps tatsächlich starten.
- Neue Spotify-Buttons visuell mit aktiver und ohne Wiedergabe prüfen; Build erfolgreich, abschließende Sichtprüfung steht aus.
- WLAN-Verbindung auf echter Hardware noch nicht getestet (VM hat nur Ethernet). Parser-/Erkennungstests vorhanden.
- Vollständigen Reset mit ISO, Passwort und zweiter Bestätigung noch Ende-zu-Ende testen. Nur Zieldisk neu installieren; zusätzliche Sentinel-Disk erhalten. BIOS-Installation früher getestet, UEFI noch vollständig prüfen.
- Vor späterem Release: aktuelle Quellen vollständig auf Builder übertragen, sauberes Image bauen, Manifest/Prüfsummen aktualisieren, frische zweite Testinstallation und Reset prüfen. Keine temporären Root-Seriell-Dienste, Testprofile, Zugangsdaten oder Builder-Zustände ins Image übernehmen.

## Aktuelle Artefakte

- Letzte veröffentlichte Datei outputs/MiracleOS-0.1-spotify-setup-amd64.iso im Chat-Arbeitsordner ist v6 (SHA256 46f1ad15915adf6b2531016b382434df54dd9f04c9fa5894d4f4a27fe830de94); nicht als aktueller Stand ausgeben.
- v7-Image auf Builder kann existieren, wurde nicht veröffentlicht. v9-Buildablauf wurde auf Nutzerwunsch vor ISO-Erstellung gestoppt; UI separat gebaut und in Test-VM installiert. Kleine Transfermedien sind keine neue veröffentlichte Installations-ISO.
- Builderzugang und VM-Hilfsskripte befinden sich im Chat-Arbeitsordner work/iso-build bzw. work/iso-test. Geheimnisse niemals in Dokumentation oder Release kopieren.
