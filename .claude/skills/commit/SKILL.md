---
name: commit
description: Commit changes in the oasdiff repo, running required pre-commit steps first
allowed-tools: Bash, Read, Glob, Grep
---

## Pre-commit steps (REQUIRED before every commit)

Run this command first:

```
make doc-breaking-changes
```

This regenerates `docs/BREAKING-CHANGES-EXAMPLES.md` from the current test files. The CI build will fail if this file is out of date. Include any changes to this file in the commit.

## Git workflow

NEVER push directly to main. Always:
1. Create a feature/fix branch
2. Commit changes there
3. Push the branch
4. Open a PR

After the pre-commit step and staging files, write a concise commit message and commit.
