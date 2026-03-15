---
name: commit
description: Commit changes in the oasdiff repo, running required pre-commit steps first
allowed-tools: Bash, Read, Glob, Grep
---

Before committing in this repo, you MUST run these steps in order:

1. **`go fmt ./...`** — formats all Go code. The CI checks for a clean diff after `go fmt`, so any formatting changes must be included in the commit.

2. **`go vet ./...`** — runs static analysis. Fix any issues before committing.

3. **`make doc-breaking-changes`** — regenerates `docs/BREAKING-CHANGES-EXAMPLES.md` from the current test files. The CI build will fail if this file is out of date. Include any changes to this file in the commit.

After running those steps, follow the standard commit process: stage files, write a concise commit message, and commit.

Note: `make lint` runs golangci-lint in addition to `go fmt` and `go vet`, but may fail locally if the installed golangci-lint version doesn't match the Go version. Running the three steps above is sufficient.
