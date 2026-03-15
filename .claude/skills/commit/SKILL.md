---
name: commit
description: Commit changes in the oasdiff repo, running required pre-commit steps first
allowed-tools: Bash, Read, Glob, Grep
---

## Pre-commit steps (REQUIRED before every commit)

Run this command first:

1. **`go fmt ./...`** — formats all Go code. The CI checks for a clean diff after `go fmt`, so any formatting changes must be included in the commit.

2. **`go vet ./...`** — runs static analysis. Fix any issues before committing.

3. **`make doc-breaking-changes`** — regenerates `docs/BREAKING-CHANGES-EXAMPLES.md` from the current test files. The CI build will fail if this file is out of date. Include any changes to this file in the commit.

Note: `make lint` runs golangci-lint in addition to `go fmt` and `go vet`, but may fail locally if the installed golangci-lint version doesn't match the Go version. Running the three steps above is sufficient.

## Git workflow

NEVER push directly to main. Always:
1. Create a feature/fix branch
2. Commit changes there
3. Push the branch
4. Open a PR

After the pre-commit step and staging files, write a concise commit message and commit.
