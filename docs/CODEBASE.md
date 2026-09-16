# Codebase guide

Where things live and the conventions each package follows. For ways to contribute see [CONTRIB.md](CONTRIB.md), and run `make help` for the common commands.

Every Go directory has a `doc.go` stating its purpose, so start there when you open one.

## Packages

| Package | Purpose |
|---------|---------|
| `checker/` | The breaking-change and changelog checks |
| `diff/` | The diff engine. It reports how two documents differ, not whether a change breaks clients, which is the checker's job |
| `data/` | OpenAPI specs used by the unit tests |

## Adding a check

- Each check is a function in `checker/` returning `Changes`, a slice of `Change`.
- Register the function in `GetAllRules()` in `checker/rules.go`. A check that is not registered never runs.
- Every `Change` carries a string ID that is unique across the catalog. Define it as a constant at the top of the file holding the check.
- Write the message for that ID under `checker/localizations_src/<locale>/`, in every locale (`en`, `es`, `pt-br`, `ru`). An English-only entry falls back silently.
- Run `make localize` to compile the messages into `checker/localizations/localizations.go`, and commit the result. That file is generated, so never edit it by hand, and CI fails when it is out of date.

## Before opening a pull request

Run `make lint` and `make test`. CI runs the same, plus govulncheck, CodeQL, and shellcheck and actionlint over the workflows.
