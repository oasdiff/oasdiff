---
name: commit
description: Commit changes in the oasdiff repo, running required pre-commit steps first
allowed-tools: Bash, Read, Glob, Grep
---

## Git workflow

NEVER push directly to main. Always:
1. Create a feature/fix branch
2. Commit changes there
3. Push the branch
4. Open a PR

Before committing, run `make lint` to ensure formatting and generated files are up to date.

After the pre-commit step and staging files, write a concise commit message and commit.
