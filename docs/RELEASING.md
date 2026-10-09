# Releases

1. Aktuelle Quellen vollständig auf die separate Builder-VM übertragen; keine Profile oder Schlüssel kopieren.
2. Tests aus README ausführen und ISO bauen. Bei Änderungen am Desktop nach dem ursprünglichen Build auch Image-Manifest und ISO-interne Prüfsummen aktualisieren.
3. ISO auf neutraler Test-VM starten. Teststand und Grenzen ehrlich festhalten.
4. SHA-256 der fertigen ISO berechnen und beim Transfer erneut vergleichen.
5. Quellstand committen, Tag erstellen und ein als Vorabversion markiertes GitHub Release veröffentlichen.
6. ISO und SHA256SUMS als Release-Assets hochladen. ISO/VM-Platten nie mit git add aufnehmen.
7. Repository muss privat bleiben, sofern der Eigentümer keine Veröffentlichung anfordert.

Build-Artefakte sind keine Backups der Test-VM. Persönliche Spotify-Sitzungen, Schlüsselbünde, SSH-Zugänge und Testhilfsdienste dürfen nicht enthalten sein.
