# Validate

`oasdiff validate <spec>` checks a single OpenAPI spec for per-RFC violations: invalid `type` values, missing required fields, malformed paths, bad regex patterns, unresolved `$ref`s, and similar structural problems. It validates the document against the OpenAPI and JSON Schema rules.

It is not a configurable style linter (like Spectral); it reports hard, spec-defined violations only.

This is distinct from `breaking` / `changelog`, which compare two specs. `validate` looks at one spec and answers "is this a valid OpenAPI document?".

## Usage

```bash
oasdiff validate openapi.yaml
```

The spec can be a file path, a URL, a git ref (e.g. `main:openapi.yaml`, see [Git revisions](GIT-REVISION.md)), or `-` to read from standard input:

```bash
oasdiff validate https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml
cat openapi.yaml | oasdiff validate -
```

## Output

The default output is human-readable text: a summary line followed by one block per finding, each with a stable rule ID and a `file:line:column` location (when the loader can track the origin):

```
2 findings: 2 error, 0 warning, 0 info
error	[duplicate-operation-id] at openapi.yaml:42:7
	duplicate operation id "listPets"
error	[security-scheme-type-invalid] at openapi.yaml:88:9
	...
```

Use `-f yaml` or `-f json` for structured output. Each finding carries `id`, `text`, `level`, `section`, `source` (`file` / `line` / `column`), and a stable `fingerprint`.

```bash
oasdiff validate -f json openapi.yaml
```

In CI, `-f githubactions` emits a GitHub Actions annotation per finding (anchored to its file/line/column) so violations show up inline on the pull request's Files Changed tab, and publishes `error_count` / `warning_count` / `info_count` as step outputs.

All findings are reported in one pass (multi-error), not just the first one.

## Severities

Every finding comes from the OpenAPI / JSON Schema rules, but they are classified by impact:

- **error** — the spec can't be reliably consumed: missing required fields, unresolved `$ref`s, invalid types, malformed paths, and similar structural breaks. Also `duplicate-operation-id`, since a non-unique operationId breaks code generators.
- **warning** — structurally valid but a real risk: a 3.1-only field in an older doc, `$ref` siblings that are silently ignored, conflicting paths, duplicate parameters, and a `default` value that doesn't match its schema.
- **info** — informational only: an `example` that doesn't match its schema (the contract itself is valid).

`--fail-on` decides which severities fail the command (see Exit codes). The classification is currently fixed; per-rule customization may be added later.

## Flags

| Flag | Default | Description |
|---|---|---|
| `-f, --format` | `text` | output format: `text`, `yaml`, `json`, or `githubactions` |
| `-o, --fail-on` | `ERR` | exit with code 1 when a finding has this severity or higher: `ERR`, `WARN`, or `INFO` |
| `--color` | `auto` | when to colorize text output: `auto`, `always`, `never` |
| `--allow-external-refs` | `true` | resolve external `$ref`s; set to `false` to prevent SSRF when validating untrusted specs |

## Exit codes

| Code | Meaning |
|---|---|
| `0` | no findings at or above the `--fail-on` severity |
| `1` | at least one finding at or above the `--fail-on` severity |
| `102` | failed to load the spec |

This makes it usable as a CI gate: by default `oasdiff validate openapi.yaml` fails the step on any error (warnings and info still print but don't fail). Use `--fail-on WARN` to also fail on warnings, or `--fail-on INFO` to fail on any finding.

## Rule IDs

Each finding has a stable, kebab-case rule ID derived from the violation, for example `path-parameters-mismatch`, `duplicate-operation-id`, `unresolved-ref`, `schema-pattern-regex-invalid`, and `security-scheme-type-invalid`. Violations that do not yet have a dedicated ID surface under the catchall `spec-validation-error`.

## Feedback

Found an issue? [Open one here](https://github.com/oasdiff/oasdiff/issues/new?template=bug_report.md&title=[validate]%20) with `[validate]` in the title.
