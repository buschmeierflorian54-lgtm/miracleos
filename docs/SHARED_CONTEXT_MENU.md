# Shared MiracleOS context menu (version 1)

Applications explicitly opt in by starting `miracleos-desktop --context-menu`.
Send one UTF-8 JSON object to standard input, then close standard input. Read one
JSON response from standard output. Diagnostics are written to standard error.
This adapter does not intercept other applications' right-click events.
Escape or the automatically added **Abbrechen** action cancels the request.

```json
{"version":1,"title":"Image","items":[{"id":"copy","label":"Kopieren"},{"id":"paste","label":"Einfügen","enabled":false},{"separator":true},{"id":"properties","label":"Eigenschaften"}]}
```

`title` is optional. Each action has a unique non-empty `id` and `label`.
`enabled` defaults to true. A separator contains only `separator: true`.
Requests are limited to 1 MiB and 100 items. Unknown fields, invalid versions,
control characters, duplicate IDs and empty action lists are rejected.

Responses:

```json
{"version":1,"status":"selected","id":"copy"}
{"version":1,"status":"cancelled"}
{"version":1,"status":"error","error":"unsupported protocol version"}
```

The application computes which actions apply and their enabled state. It handles
the returned ID itself. MiracleOS never executes IDs, commands, URLs or file
operations supplied through this protocol. Labels use plain text, not markup.

```python
import json
import subprocess

request = {
    "version": 1,
    "items": [
        {"id": "copy", "label": "Kopieren", "enabled": has_selection},
        {"id": "paste", "label": "Einfügen", "enabled": can_paste},
    ],
}
completed = subprocess.run(
    ["/home/test/desktop/miracleos-desktop", "--context-menu"],
    input=json.dumps(request), text=True, capture_output=True, check=True,
)
response = json.loads(completed.stdout)
if response["status"] == "selected":
    handlers = {"copy": copy_selection, "paste": paste_selection}
    handler = handlers.get(response["id"])
    if handler is not None:
        handler()
```

The native GTK menu uses the same `buildSharedMenu` renderer and system palette
as built-in menus. The standalone process does not have the originating app's
Wayland surface and input serial: under MiracleOS/Sway it uses a tiny centered
layer-shell anchor, so it cannot promise placement at the originating click.
The adapter temporarily receives keyboard focus while open. Outside-click
dismissal depends on the compositor; Escape and **Abbrechen** are supported.
Built-in GTK integration can call the shared renderer with its own parent and
input event for correctly anchored popups. Browsers and third-party applications
require their own explicit adapter; this is not a browser extension or a global
replacement of all context menus.
