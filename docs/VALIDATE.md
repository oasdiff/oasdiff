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

Use `-f yaml` or `-f json` for structured output. Each finding carries `id`, `text`, `level`, `section`, `source` (`file` / `line` / `column`), and a stable `fingerprint`. The field names match the `changelog` command's output, so a single CI script can parse both.

```bash
oasdiff validate -f json openapi.yaml
```

All findings are reported in one pass (multi-error), not just the first one.

## Flags

| Flag | Default | Description |
|---|---|---|
| `-f, --format` | `text` | output format: `text`, `yaml`, or `json` |
| `--color` | `auto` | when to colorize text output: `auto`, `always`, `never` |
| `--allow-external-refs` | `true` | resolve external `$ref`s; set to `false` to prevent SSRF when validating untrusted specs |

## Exit codes

| Code | Meaning |
|---|---|
| `0` | no findings |
| `1` | at least one finding |
| `102` | failed to load the spec |

This makes it usable as a CI gate: `oasdiff validate openapi.yaml` fails the step on any finding.

## Rule IDs

Each finding has a stable, kebab-case rule ID derived from the violation, for example `path-parameters-mismatch`, `duplicate-operation-id`, `unresolved-ref`, `schema-pattern-regex-invalid`, and `security-scheme-type-invalid`. Violations that do not yet have a dedicated ID surface under the catchall `spec-validation-error`.
