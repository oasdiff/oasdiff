# Contributing to oasdiff

Thank you for your interest in contributing! Here are the ways you can help.

---

## Use it and spread the word

The simplest contribution is to use oasdiff in your projects and tell others about it.

- ⭐ [Star the repo](https://github.com/oasdiff/oasdiff) — helps others discover the project
- Share it with your team or write about your experience

## Set up automated PR checks

If your team uses GitHub, install the [oasdiff GitHub Action](https://github.com/oasdiff/oasdiff-action) to automatically detect breaking changes on every pull request:

```yaml
- uses: oasdiff/oasdiff-action/changelog@v0.0.37
  with:
    base: 'origin/${{ github.base_ref }}:openapi.yaml'
    revision: 'HEAD:openapi.yaml'
```

The [Pro plan at oasdiff.com](https://www.oasdiff.com/pricing) adds a review workflow: each breaking change must be approved or rejected before the PR can merge.

## Test OpenAPI 3.1 (beta)

OpenAPI 3.1 support is in beta and needs real-world testing before general availability. See [OPENAPI-31.md](OPENAPI-31.md) for installation instructions and what's covered.

- Found an issue? [Open a bug report](https://github.com/oasdiff/oasdiff/issues/new?template=bug_report.md&title=[3.1]%20) with `[3.1]` in the title.
- Working well? Add a 👍 to the [tracking issue](https://github.com/oasdiff/oasdiff/issues/52) and describe your use case — this directly informs the GA decision.

## Improve the project

| Area | How to contribute |
|------|-------------------|
| **Documentation** | Edit files under [`docs/`](../docs) and open a PR |
| **Message text** | Review and improve texts in [`checker/localizations_src`](../checker/localizations_src) |
| **Translations** | Add messages in your language under [`checker/localizations_src`](../checker/localizations_src) — run `make localize` to regenerate |
| **Bug fixes & features** | Pick up an [open issue](https://github.com/oasdiff/oasdiff/issues) or propose your own |

For non-trivial changes, open an issue first to discuss the approach before writing code.
