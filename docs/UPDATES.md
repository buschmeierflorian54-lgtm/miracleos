# Automatische Updates

MiracleOS prüft täglich die Releases von [buschmeierflorian54-lgtm/miracleos](https://github.com/buschmeierflorian54-lgtm/miracleos). Das öffentliche Repository erfordert keinen Token. Nach dem Start wartet der Dienst 10 bis 20 Minuten; ausgeschaltete automatische Updates werden respektiert.

## Nutzung

In **Einstellungen → Updates** stehen Status, manuelle Aktualisierung, Aktivieren/Pausieren und Wiederherstellung der vorherigen Version zur Verfügung. Manuelle Systemänderungen benötigen das MiracleOS-Passwort. Der Hintergrunddienst installiert ohne weitere Abfrage.

Aktualisiert werden der MiracleOS-Desktop, Luna Browser, der angepasste Sway-Compositor, der Tab-Dienst und die mitgelieferten Python-Laufzeitdateien. Laufende Prozesse werden nicht beendet; nach erneuter Anmeldung wird die neue Version verwendet. Debian-Pakete, Kernel, persönliche Daten und Benutzerprofile sind nicht Bestandteil dieses Update-Kanals.

Für bereits installierte Systeme aus der ISO vom 19.10.2026 muss einmal das Paket `miracleos-updater_0.1.20261020_amd64.deb` aus dem Update-Release installiert werden:

```sh
sudo apt install ./miracleos-updater_0.1.20261020_amd64.deb
sudo /usr/local/lib/miracleos/update-manager.py apply
```

Künftige mit dem aktuellen `scripts/iso/build.sh` gebaute ISOs enthalten den Dienst. Die bereits veröffentlichte ISO vom 19.10.2026 wird durch dieses Release nicht ersetzt.

## Prüfung und Wiederherstellung

Der Client vertraut dem installierten öffentlichen Ed25519-Schlüssel unter `/etc/miracleos/update-public.pem`. Er prüft die Signatur des Manifests sowie Repository, Architektur, Version, Größe und SHA-256 des Pakets. Archive dürfen nur bekannte Programme und Laufzeitdateien enthalten; Pfadwechsel, Links und doppelte Dateien werden abgewiesen. Es werden keine Skripte aus Releases ausgeführt.

Jede Version liegt in einem eigenen Verzeichnis unter `/opt/miracleos/updates/releases`. Ein atomarer Wechsel von `current` aktiviert den vollständigen Satz. Die vorherige Installation bleibt erhalten. Wiederherstellen pausiert automatische Updates, damit die zurückgenommene Version nicht sofort erneut installiert wird. Erneutes Aktivieren ist eine bewusste Entscheidung in den Einstellungen. Alte Versionen werden zunächst nicht automatisch gelöscht; bei weniger als 2 GB freiem Speicher stoppt der Updater.

Netzwerk-, Signatur- und Downloadfehler belassen die aktive Version. Status: `/var/lib/miracleos/updates/status.json`; Protokoll: `journalctl -u miracleos-update.service`.

## Updates veröffentlichen

Der aktuelle Testkanal berücksichtigt auch Vorab-Releases. Ein Release wird nur als Update erkannt, wenn alle drei Assets vorhanden sind:

- `miracleos-update.json`
- `miracleos-update.json.sig`
- `miracleos-update-amd64.tar.gz`

`scripts/iso/package-update.py` erstellt diese Assets aus getesteten amd64-Binärdateien. `--sequence` muss für jedes neue Update streng steigen; `--tag` muss dem GitHub-Release-Tag entsprechen. Beispiel:

```sh
python3 scripts/iso/package-update.py \
  --desktop /build/miracleos-desktop --luna /build/luna-browser \
  --sway /build/miracleos-sway --broker /build/miracleos-app-tabs \
  --key /secure/private-update-key.pem --output /build/release \
  --tag v0.1.0-update.20261020 --version 0.1.0-update.20261020 \
  --sequence 2026102001
```

Den privaten Schlüssel außerhalb von Repository, Release, Update-Paket und ISO aufbewahren. Der öffentliche Schlüssel ist versioniert. Veröffentlichte Assets nicht durch anders signierte Inhalte mit derselben Versionsnummer ersetzen.

## Tests

`python3 -m unittest discover -s scripts/iso -p 'test_*.py'` prüft Installer und Updater, darunter manipulierte Signaturen, fremde Release-Tags, Dateipfade, Links, doppelte Archivdateien, Prüfsummenfehler, Versionsvergleich und Wiederherstellung. Go/GTK-Tests laufen zusätzlich unter Xvfb. Ein Dienstneustart oder neuer ISO-Build ersetzt keinen Test in einer installierten VM.
