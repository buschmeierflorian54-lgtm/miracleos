# Planned automatic updates from GitHub

Automatic updates are a requirement for future work. They are not implemented yet.

## Release contract

Publish versioned bundles through GitHub Releases, rather than downloading loose files from a branch. Each bundle must include the desktop executable and its matching `runtime/` helpers. Record the version, Linux architecture, supported system dependencies and checksums in a manifest. Authenticate release metadata and payloads with a trusted signing key; a checksum alone does not prove who published an update.

Build and test the whole bundle before publishing. Keep user preferences, passwords, files and VM connection credentials out of release packages.

## Installation and recovery

Install versions into separate version directories and point the service at the active version. Download and verify into staging, then switch the active version atomically. Never replace only the executable while leaving incompatible runtime helpers behind. Retain the previous version for rollback if startup or a health check fails.

The taskbar service must have a stable installation path independent of a developer checkout. Its current `/home/test/desktop` path is a development setup and must be migrated before enabling automatic updates.

Keep `/home/<user>/.config/miracleos` and user documents outside installation directories. Handle settings migrations explicitly and keep rollback-compatible backups. System-owned wallpapers, the protected login executable, greetd configuration and other privileged components require a separate authenticated privileged installation step. Never ask an unprivileged updater to overwrite those directly.

Use the same verified release package for manual installation and automatic updates. The updater should download in the background and apply according to the user's update policy. Authentication and active file transfers must survive or finish before affected processes restart. A failed update must leave a working desktop available.

## Follow-up implementation

1. Version/build metadata and reproducible release bundles.
2. Signed GitHub Release publication with CI tests.
3. Stable installation paths and a privileged installer where required.
4. Background checking, staging, verification, activation and health checks.
5. Rollback tests, settings migrations and update UI.

Central window controls now also require the version-matched compositor bundle. Include its corresponding upstream source and patch to meet the Sway GPL requirements. Compositor updates activate on the next desktop login; restart only after the user's session policy permits it.
