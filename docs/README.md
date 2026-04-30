
[![CI](https://github.com/oasdiff/oasdiff/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/oasdiff/oasdiff/actions)
[![codecov](https://codecov.io/gh/oasdiff/oasdiff/branch/main/graph/badge.svg?token=Y8BM6X77JY)](https://codecov.io/gh/oasdiff/oasdiff)
[![Go Report Card](https://goreportcard.com/badge/github.com/oasdiff/oasdiff)](https://goreportcard.com/report/github.com/oasdiff/oasdiff)
[![GoDoc](https://godoc.org/github.com/oasdiff/oasdiff?status.svg)](https://godoc.org/github.com/oasdiff/oasdiff)
[![Docker Image Version](https://img.shields.io/docker/v/tufin/oasdiff?sort=semver)](https://hub.docker.com/r/tufin/oasdiff/tags)

![oasdiff banner](https://github.com/yonatanmgr/oasdiff/assets/31913495/ac9b154e-72d1-4969-bc3b-f527bbe7751d)


Command-line and Go package to compare and detect breaking changes in OpenAPI specs.

Run it locally, in CI via the [GitHub Action](https://github.com/oasdiff/oasdiff-action), or use the hosted PR review workflow at [oasdiff.com](https://www.oasdiff.com) to approve or reject each change with a CI commit status.

## Get started in 30 seconds

No install needed — try it with Docker against two sample specs:

```bash
docker run --rm -t tufin/oasdiff changelog \
  https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml \
  https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test5.yaml
```

That prints a human-readable changelog of every significant change between the two specs. Swap `changelog` for `breaking` to see only breaking changes, or `diff` for the full machine-readable diff.

## Installation

### Install with Go
```bash
go install github.com/oasdiff/oasdiff@latest
```

### Install on macOS with Brew
```bash
brew install oasdiff
```

### Install on macOS and Linux using curl

```bash
curl -fsSL https://raw.githubusercontent.com/oasdiff/oasdiff/main/install.sh | sh
```

### Install with asdf

https://github.com/oasdiff/asdf-oasdiff

### Manually install on macOS, Windows and Linux
Copy binaries from [latest release](https://github.com/oasdiff/oasdiff/releases/).  

### Use install.sh
You can use the [install.sh](../install.sh) script to install oasdiff.  
The script will download the latest version, or a specific version of oasdiff and install it in /usr/local/bin.  

## Documentation

Grouped by what you're trying to do. New to oasdiff? Start with **Commands**.

### Commands
The five top-level subcommands.

- [`diff`](DIFF.md) — full diff between two OpenAPI specs (output: html, json, markdown, markup, text, or yaml — default yaml)
- [`breaking`](BREAKING-CHANGES.md) — only breaking changes
- [`changelog`](BREAKING-CHANGES.md) — every significant change, breaking or not, in human-readable form
- [`flatten`](ALLOF.md) — replace `allOf` schemas with a merged equivalent
- [`checks`](CHECKS.md) — list the rules oasdiff uses to classify changes ([customize them](CUSTOMIZING-CHECKS.md))

### Inputs
What you can compare and how oasdiff loads it.

- [Git revisions](GIT-REVISION.md) — compare against a branch, tag, or commit
- [Composed mode](COMPOSED.md) — compare two collections of specs (e.g. behind an API gateway)
- [OpenAPI 3.1](OPENAPI-31.md) — what's supported
- [`allOf` merging](ALLOF.md)
- [Common (path-level) parameter merging](COMMON-PARAMS.md)
- Local files, http/s URLs, YAML or JSON — all handled transparently

### Matching & filtering endpoints
Tell oasdiff which endpoints in the base correspond to which in the revision.

- [Endpoint matching](MATCHING-ENDPOINTS.md) (including [duplicate endpoints](MATCHING-ENDPOINTS.md#duplicate-endpoints))
- [Filter endpoints](FILTERING-ENDPOINTS.md)
- [Path prefix modification](PATH-PREFIX.md)
- [Path parameter renaming](PATH-PARAM-RENAME.md)
- [Case-insensitive header comparison](HEADER-DIFF.md)

### API lifecycle
Communicate intent across versions.

- [Deprecate APIs and parameters](DEPRECATION.md)
- [API stability levels](STABILITY.md) (draft / alpha / beta / stable)

### Output & tracking
Shape, enrich, and track changes across runs.

- [Customize HTML and Markdown changelog templates](CHANGELOG-TEMPLATE.md)
- [Add OpenAPI-extension attributes to changelog entries](ATTRIBUTES.md)
- [Source location tracking](SOURCE-LOCATOR.md)
- [Change fingerprints](FINGERPRINT.md) — stable IDs across commits
- [Exclude certain kinds of changes](DIFF.md#excluding-specific-kinds-of-changes), [exclude extension names](DIFF.md#excluding-specific-extension-names), [track OpenAPI extensions](DIFF.md#openapi-extensions)
- [Error reporting](ERRORS.md)
- Localization: en, ru, pt-br, es

### How to run
- [Docker](DOCKER.md)
- [Configuration file](CONFIG-FILES.md)
- [Embed in a Go program](GO.md)
- [GitHub Action](https://github.com/oasdiff/oasdiff-action) for CI — and [oasdiff.com](https://www.oasdiff.com) for teams, which adds a per-change PR comment with approve/reject and commit-status checks

### Reference
- [Security: control external `$ref` loading to prevent SSRF](SECURITY.md)
- [Usage examples](USAGE_EXAMPLES.md) — recipes for common scenarios
- [Contributing](CONTRIB.md)

## Demo
<img src="./demo.svg">

## Credits
This project relies on the excellent implementation of OpenAPI 3.0 and 3.1 for Go: [kin-openapi](https://github.com/getkin/kin-openapi).

## Feedback
We welcome your feedback.  
If you have ideas for improvement or additional needs around APIs, please [let us know](https://github.com/oasdiff/oasdiff/discussions/new?category=ideas).
