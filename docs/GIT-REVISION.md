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
| `e69de29:openapi.yaml` | Spec at a specific blob SHA (the form git's external-diff protocol passes — see [Git Diff Driver](GIT-DIFF-DRIVER.md)) |

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

## Fetching missing commits (`--fetch`)

By default, oasdiff reads Git revisions from the local object store. When a revision is missing, the error shows the command to fetch it. Use `--fetch` to load a missing revision from the configured `origin` remote.

```bash
oasdiff changelog --fetch <base-sha>:openapi.yaml <revision-sha>:openapi.yaml
oasdiff changelog --fetch origin/main:openapi.yaml HEAD:openapi.yaml
```

For a missing `origin/<branch>`, oasdiff fetches the branch tip with `--depth=1` into `refs/remotes/origin/<branch>`. It reads locally available revisions directly. A missing path in a local revision is reported as a file error.

## GitHub Actions

Git revision syntax is particularly useful in CI/CD. The following workflow detects breaking changes between the base branch and the PR branch without temp files:

```yaml
name: API breaking changes

on:
  pull_request:

jobs:
  oasdiff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - run: git fetch --depth=1 origin "refs/heads/$GITHUB_BASE_REF:refs/remotes/origin/$GITHUB_BASE_REF"

      - uses: oasdiff/oasdiff-action/breaking@v0
        with:
          base: 'origin/${{ github.base_ref }}:openapi.yaml'
          revision: 'HEAD:openapi.yaml'
```

`actions/checkout` makes a shallow clone that holds only the pull request's commit, so the base branch has to be fetched before `origin/<branch>` resolves. Fetching that one branch is enough, and `fetch-depth: 0` is not required.

A CLI step can compare the base branch with the checked-out pull request commit using `--fetch`.

```bash
oasdiff breaking --fetch "origin/$GITHUB_BASE_REF:openapi.yaml" HEAD:openapi.yaml
```

## See also

- [Git Diff Driver](GIT-DIFF-DRIVER.md): wire oasdiff into git as an external diff driver so `git log --patch` and `git diff` render a human-readable OpenAPI changelog inline.
