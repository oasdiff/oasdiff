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

Each comparison runs the same checks as `oasdiff breaking`, so what the docs say about `breaking` applies to `breaking-files` too, unless this page says otherwise.

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

oasdiff ships a [pre-commit](https://pre-commit.com/) hook built on `breaking-files`. See [Pre-commit Hook](PRE-COMMIT-HOOK.md).
