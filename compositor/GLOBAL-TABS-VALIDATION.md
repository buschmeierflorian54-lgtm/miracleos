# Global app tabs: 2026-10-19

The first tab displays the application's normal name. Subsequent tabs display
their content/window titles. Native providers (Luna, Files) retain their actual
internal tabs. Other applications group real windows; plus launches another
window through its desktop entry. Single-instance apps do not get phantom tabs.

Active tab captions pass pointer presses through to Sway's floating-window move
operation. Selection of inactive tabs and tab-close/plus buttons remain handled
by the tab bridge. Initial oversized windows are fitted into the workspace.

Validation completed on the isolated builder:

- `go test ./internal/apptabs` passed.
- `compositor/check-global-tabs.py` passed: real plus, one visible window,
  selecting a tab, closing and restoring the remaining window, desktop app name.
- Native integration: Luna loads local HTTP pages, switches real tabs, creates
  tabs; Files creates and switches tabs successfully.
- Full desktop Go tests require a graphical display; an earlier run without
  DISPLAY failed at GTK initialization and is not a passing full-suite result.

Visible MiracleOS-ISO-Test VM:

- First tab shows Luna Browser; additional tabs show page titles.
- Actual VirtualBox absolute mouse dragging moved the active browser tab/window
  by 30 pixels left and up. Sway IPC cursor warps do not exercise pointer_motion.
- App/taskbar icons returned after restarting user services with the current
  SWAYSOCK. Session startup must restart the bar and tab broker, not reuse a
  process holding the previous compositor socket; configure-image.sh includes it.
- No original VM changes and no installer ISO rebuild. Transfer CDs are updates,
  not installable OS images. Source changes must be included in the next ISO build.
