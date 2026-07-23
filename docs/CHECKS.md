# Checks
The `oasdiff checks` command displays all checks that oasdiff uses to detect API changes.
Checks are the individual rules that `oasdiff breaking` and `oasdiff changelog` apply when comparing two OpenAPI specs.
This command is typically used to explore what oasdiff can detect or to identify check IDs for ignoring or customizing specific rules.

## Example: display all checks
```
oasdiff checks
```

## Output Formats
The default output format is `text`.
Additional formats can be generated using the `--format` flag:
- text: human-readable table with ID, description, and severity level (default)
- yaml: machine-readable output, suitable for further processing
- json: machine-readable output, suitable for further processing

## Filtering by Severity
Use `--severity` to show only checks at a given level:
```
oasdiff checks --severity error
oasdiff checks --severity warn
oasdiff checks --severity info
```

Checks are categorized into three severity levels:
- `error` — definite breaking changes which should be avoided
- `warn` — potential breaking changes which cannot be confirmed programmatically
- `info` — non-breaking changes

Run `oasdiff checks` for the current list, or browse the full catalog at [oasdiff.com/docs/breaking-changes](https://www.oasdiff.com/docs/breaking-changes).

## Categorization
Every check is categorized along independent axes, emitted as fields in the `json` and `yaml` output:

- `area` — the OpenAPI object the check concerns, aligned with the OpenAPI specification's object model: `schema`, `parameters`, `requestBody`, `responses`, `paths`, `headers`, `security`, `tags`, `components`.
- `kind` — the aspect of the API contract that changed: `existence` (an element added or removed), `requiredness` (required / optional / nullable), `mutability` (read-only / write-only), `type` (data type or format), `constraints` (bounds such as min/max, length, pattern, items), `values` (enum, const, default), `structure` (composition and applicator keywords such as allOf/anyOf/oneOf, discriminator, if/then/else, contains), and `lifecycle` (deprecation, sunset, stability).
- `action` — the verb: `add`, `remove`, `change`, `generalize`, `specialize`, `increase`, `decrease`, `set`.
- `direction` — `request`, `response`, or `none`.

## Filtering by Tag
Use `--tags` to show only checks in a specific area, kind, action, or direction:
```
oasdiff checks --tags request,parameters
oasdiff checks --tags schema,constraints
```

Available tags: `request`, `response`, `add`, `remove`, `change`, `generalize`, `specialize`, `increase`, `decrease`, `set`, `schema`, `parameters`, `requestBody`, `responses`, `paths`, `headers`, `security`, `tags`, `components`, `existence`, `requiredness`, `mutability`, `type`, `constraints`, `values`, `structure`, `lifecycle`.

Multiple tags are combined with AND — only checks that match all specified tags are shown.

## Localization
Use `--lang` to view check descriptions in a supported language:
```
oasdiff checks --lang ru
```
Supported languages: `en` (default), `ru`, `pt-br`, `es`.

## Using Check IDs
Each check has a unique ID (e.g. `api-path-removed-without-deprecation`) which can be used to:
- [Ignore specific changes](BREAKING-CHANGES.md#ignoring-specific-breaking-changes)
- [Customize severity levels](BREAKING-CHANGES.md#customizing-severity-levels)
- [Write custom checks](CUSTOMIZING-CHECKS.md)

