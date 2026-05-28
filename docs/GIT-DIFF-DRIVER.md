# Git Diff Driver

oasdiff can run as a [git external diff driver](https://git-scm.com/docs/git#Documentation/git.txt-codeGITEXTERNALDIFFcode) so that `git log --patch`, `git diff`, and `git show` render a human-readable changelog of OpenAPI changes inline, instead of a raw YAML text diff.

Inspired by [Jamie Tanna's post on using oasdiff as a git diff driver](https://www.jvt.me/posts/2026/04/11/oasdiff-driver/). The previous bash wrapper described there is no longer needed.

## Setup

Two one-time configuration lines, run from your repo:

```bash
git config diff.oasdiff.command "oasdiff git-diff-driver"
echo "openapi.yaml diff=oasdiff" >> .gitattributes
```

Replace `openapi.yaml` with the actual spec path (or a glob like `**/openapi.yaml` for multi-spec repos). Commit the `.gitattributes` change so teammates pick it up.

## Usage

The driver kicks in on any git command that emits a diff, as long as `--ext-diff` is passed. Three common cases:

```bash
# Browse the full history of a spec
git log --patch --ext-diff -- openapi.yaml

# Inspect a single commit
git show --ext-diff <commit>

# See what you're about to commit (working tree vs HEAD)
git diff --ext-diff -- openapi.yaml
```

In each case, the YAML hunks are replaced by the human-readable changelog `oasdiff changelog` would emit. For `git log` against the example above:

```
commit aed51e7b25195d82c3edd76caab75c4e1a2a2922
Author: Jane <jane@example.com>
Date:   Thu May 28 15:14:56 2026 +0300

    v2

1 changes: 0 error, 0 warning, 1 info
info  [response-non-success-status-added] at openapi.yaml
        in API GET /pets
                added the non-success response with the status `404`


commit 7dd4bdb866153dcd3e7cbc47d41953f05ff671c6
Author: Jane <jane@example.com>
Date:   Thu May 28 15:14:56 2026 +0300

    v1

Added openapi.yaml
```

Without `--ext-diff`, git falls back to the raw text diff. To make `--ext-diff` the default for a workflow, set `diff.external` in git config (see git's docs).

## What's handled

| Case | Output |
|---|---|
| Both versions exist (normal modification) | Full changelog between the two blob versions |
| File added (root commit, or new file) | `Added <path>` |
| File deleted | `Removed <path>` |
| Mode-only change (chmod) | Empty — git's own machinery already prints the mode delta |
| Load error (malformed spec, etc.) | Error surfaced inline; the diff pipeline continues for other files |

## Why `cat-file`, not git's temp files

`git-diff-driver` ignores the temp file paths git passes (`/tmp/.git-blob-xxxxx`) and re-reads each blob directly via `git cat-file blob <hex>`. The source labels in the output are the short hex of the blob plus the in-tree path (e.g. `abc1234:openapi.yaml`) rather than meaningless tempfile paths.

## Limitations

- The `--ext-diff` flag is a git design choice — there's no way for the driver to opt-in automatically.

## See also

- [Loading specs from git revisions](GIT-REVISION.md) — the `<ref>:<path>` syntax oasdiff supports generally, including blob hashes.
- [Breaking changes and changelog](BREAKING-CHANGES.md) — the underlying command whose output the git diff driver renders.
