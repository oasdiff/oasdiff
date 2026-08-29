# Checking Many Specs Against a Git Ref

`oasdiff breaking` already compares one spec to its version in a git ref, since either argument can be a git revision:

```
oasdiff breaking origin/main:openapi.yaml openapi.yaml
```

`oasdiff breaking-files` does that for a list of specs at once, deriving each one's base from its own path:

```
oasdiff breaking-files --base origin/main --fail-on ERR openapi.yaml users.yaml orders.yaml
```

The difference is the number of comparisons. `breaking` always makes one, whether the arguments are files, globs or git revisions. `breaking-files` makes one per spec and reports each separately, so it can name the spec that broke, and it exits non-zero if any of them has changes at or above `--fail-on`.

That suits anything that already knows which files changed and needs a single verdict over all of them, such as a pre-commit hook or a CI job driven by `git diff --name-only`.

## Behavior

A spec that is not in the base ref is newly added and is skipped, since a new API has no prior version to break:

```
$ oasdiff breaking-files --base origin/main --fail-on ERR new-api.yaml openapi.yaml
=== new-api.yaml ===
new file, not in base ref, skipped
=== openapi.yaml ===
1 changes: 1 error, 0 warning, 0 info
error	[api-path-removed-without-deprecation] at openapi.yaml
	in API GET /pets
		api path removed without deprecation
```

Every argument must be a spec file in the working tree. Stdin, a URL and a git revision are not paths that can be looked up in a ref, so they are refused. Being absent from the base ref is not a reason for refusal: that is the newly added case above.

Results are printed one spec at a time, so a structured `--format`, or a `--template`, is rendered once per spec rather than producing a single combined document.

## Not the same as composed mode

[Composed mode](COMPOSED.md) merges the specs matching a base glob and those matching a revision glob into a single comparison, for when several files describe one API. `breaking-files` keeps the specs separate, for when each file is its own API. Composed mode is therefore not available here, and `breaking-files` is not what you want for a spec split across files.

## Pre-commit hook

oasdiff ships a [pre-commit](https://pre-commit.com/) hook built on `breaking-files`, so breaking changes are caught before a commit lands, not only in CI. Add it to your `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/oasdiff/oasdiff
    rev: v1.30.0
    hooks:
      - id: oasdiff-breaking
```

`rev` pins the oasdiff release the hook comes from. Run `pre-commit autoupdate` to move it to the newest tag; there is no need to track oasdiff releases by hand.

The hook's `files` pattern decides which staged specs pre-commit passes to oasdiff, and by default it matches `openapi.yaml`, `openapi.yml` and `openapi.json`. Override it, and the flags, to match your setup. `args` replaces the hook's defaults rather than adding to them, so keep `--base` when you override it:

```yaml
      - id: oasdiff-breaking
        files: (^|/)api/.*\.yaml$
        args: [--base, origin/main, --fail-on, WARN]
```

pre-commit shows a hook's output only when it fails, which is when you need the report: the run that blocks the commit is the one that tells you which spec broke and how.

A fully annotated config, covering both settings above and the mistakes they invite, is at [`examples/.pre-commit-config.yaml`](../examples/.pre-commit-config.yaml).
