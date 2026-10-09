# Development and maintenance

Build from the repository root with `go build -o miracleos-desktop ./cmd/desktop` and test with `go test ./...`. Keep source and tests in the same Go package. The application is still one package: extracting real reusable internal packages should happen alongside feature work, with explicit interfaces and regression tests.

The application loads CSS and most images through `go:embed`; it must not depend on the current working directory. Python runtime helpers live in `runtime/` beside the executable. Protected default wallpapers are installed under `/usr/share/backgrounds/miracleos`.

Setup scripts in `scripts/install/` modify real session configuration. `configureBootLogin.py` is the original greeter setup; `polishBootLogin.py` applies the newer login screen. Do not run either casually on a production machine.

The Windows workspace's `work/archive/development/` stores historical screenshots, one-off debugging helpers, installer downloads and private VM connection files separately. Those are not application source and must not be committed. `outputs/` contains user deliverables and the pre-cleanup backup.

For a binary update in the VM, build to a temporary filename, replace the executable after a successful build, then restart the user service. Already-running application windows continue using their old binary until reopened.
