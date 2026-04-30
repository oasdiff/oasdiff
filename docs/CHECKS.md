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
- `error` — definite breaking changes which should be avoided (~111 checks)
- `warn` — potential breaking changes which cannot be confirmed programmatically (~24 checks)
- `info` — non-breaking changes (~170 checks)

## Filtering by Tag
Use `--tags` to show only checks related to a specific area:
```
oasdiff checks --tags request,parameters
```

Available tags include: `add`, `body`, `change`, `components`, `decrease`, `generalize`, `headers`, `increase`, `parameters`, `properties`, `remove`, `request`, `response`, `security`, `set`, `specialize`.

Multiple tags are combined with AND — only checks that have all specified tags are shown.

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

