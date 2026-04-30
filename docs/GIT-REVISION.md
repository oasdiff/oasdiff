# Loading Specs from Git Revisions

oasdiff can load OpenAPI specs directly from any git revision using the standard `<ref>:<path>` syntax — no extra flags, no temp files, no manual `git show` steps.

## Syntax

```
<git-ref>:<path-to-spec>
```

Examples:

| Expression | Meaning |
|---|---|
| `HEAD:openapi.yaml` | Spec at `openapi.yaml` in the current commit |
| `HEAD~1:openapi.yaml` | Spec one commit before the current HEAD |
| `origin/main:openapi.yaml` | Spec at `openapi.yaml` on the remote `main` branch |
| `v2.3.0:api/openapi.yaml` | Spec at `api/openapi.yaml` at tag `v2.3.0` |
| `abc1234:openapi.yaml` | Spec at a specific commit SHA |

## Usage

Compare the current working-tree spec against the version on `origin/main`:

```bash
oasdiff breaking origin/main:openapi.yaml openapi.yaml
```

Compare across two tags:

```bash
oasdiff changelog v1.0.0:openapi.yaml v2.0.0:openapi.yaml
```

Compare yesterday's `HEAD` against today's:

```bash
oasdiff diff HEAD~1:openapi.yaml HEAD:openapi.yaml
```

## How it works

oasdiff detects the `<ref>:<path>` pattern and runs `git show <ref>:<path>` internally to retrieve the spec content. Relative `$ref`s in the spec are resolved against the spec's path, matching the behaviour of loading from a local file.

The command must be run from within the git repository that contains the spec.

## GitHub Actions

Git revision syntax is particularly useful in CI/CD. The following workflow detects breaking changes between the base branch and the PR branch without needing a separate checkout step or temp files:

```yaml
name: API breaking changes

on:
  pull_request:

jobs:
  oasdiff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0          # full history needed for git refs

      - uses: oasdiff/oasdiff-action/breaking@v0.0.44
        with:
          base: 'origin/${{ github.base_ref }}:openapi.yaml'
          revision: 'HEAD:openapi.yaml'
```

> **Note:** `fetch-depth: 0` is required. The default shallow clone used by `actions/checkout` does not contain the history or remote refs that git revision syntax relies on.
