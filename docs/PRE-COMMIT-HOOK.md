# Pre-commit Hook

oasdiff ships a [pre-commit](https://pre-commit.com/) hook built on [`breaking-files`](BREAKING-FILES.md), so breaking changes are caught before a commit lands, not only in CI. Add it to your `.pre-commit-config.yaml`:

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
