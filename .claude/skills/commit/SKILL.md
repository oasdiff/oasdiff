---
name: commit
description: Commit changes in the oasdiff repo, running required pre-commit steps first
allowed-tools: Bash, Read, Glob, Grep
---

Before committing in this repo, you MUST run:

```
make doc-breaking-changes
```

This regenerates `docs/BREAKING-CHANGES-EXAMPLES.md` from the current test files. The CI build will fail if this file is out of date. Include any changes to this file in the commit.

After running that command, follow the standard commit process: stage files, write a concise commit message, and commit.
