# Checks
The `oasdiff checks` command displays the rules oasdiff applies. There is one listing per rule set:

- `oasdiff checks changelog` — the checks that `oasdiff breaking` and `oasdiff changelog` apply when comparing two specs. The rest of this page describes these.
- `oasdiff checks validate` — the rules `oasdiff validate` reports for a single spec.

This command is typically used to explore what oasdiff can detect or to identify check IDs for ignoring or customizing specific rules.

`oasdiff checks` on its own prints the two subcommands: a listing always names its rule set.

## Example: display all checks
```
oasdiff checks changelog
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
oasdiff checks changelog --severity error
oasdiff checks changelog --severity warn
oasdiff checks changelog --severity info
```

Checks are categorized into three severity levels:
- `error` — definite breaking changes which should be avoided
- `warn` — potential breaking changes which cannot be confirmed programmatically
- `info` — non-breaking changes

Run `oasdiff checks changelog` for the current list, or browse the full catalog at [oasdiff.com/docs/breaking-changes](https://www.oasdiff.com/docs/breaking-changes).

## Categorization
Every check is categorized along independent axes, emitted as fields in the `json` and `yaml` output:

- `area` — the OpenAPI object the check concerns, aligned with the OpenAPI specification's object model: `schema`, `parameters`, `requestBody`, `responses`, `paths`, `headers`, `security`, `tags`, `components`.
- `kind` — the aspect of the API contract that changed: `existence` (an element added or removed), `requiredness` (required / optional / nullable), `mutability` (read-only / write-only), `type` (data type or format), `constraints` (bounds such as min/max, length, pattern, items), `values` (enum, const, default), `structure` (composition and applicator keywords such as allOf/anyOf/oneOf, discriminator, if/then/else, contains), and `lifecycle` (deprecation, sunset, stability).
- `actions` — the syntactic edits the check covers, derived from its position in the OpenAPI object model: `add`, `remove`, `change`, `increase`, `decrease`, `set`, `unset`.
- `effect` — the check's verdict about the set of payloads the contract accepts: `widens`, `narrows`, `incomparable` (provably neither), `unknown` (cannot be decided), `violation` (breaks the deprecation/stability contract rather than the wire contract), or `none` (metadata with no effect on accepted payloads). Together with `direction`, the effect determines the default severity: narrowing requests and widening responses break clients.
- `direction` — `request`, `response`, or `none`.

## Filtering by Tag
Use `--tags` to show only checks in a specific area, kind, action, effect, or direction:
```
oasdiff checks changelog --tags request,parameters
oasdiff checks changelog --tags schema,constraints
```

Available tags: `request`, `response`, `add`, `remove`, `change`, `increase`, `decrease`, `set`, `unset`, `widens`, `narrows`, `schema`, `parameters`, `requestBody`, `responses`, `paths`, `headers`, `security`, `tags`, `components`, `existence`, `requiredness`, `mutability`, `type`, `constraints`, `values`, `structure`, `lifecycle`. The retired action names `generalize` and `specialize` are kept as aliases for `widens` and `narrows`.

Multiple tags are combined with AND — only checks that match all specified tags are shown.

## Coverage Map
[COVERAGE.md](COVERAGE.md) shows the inverse view: every field location and edit in the OpenAPI object model, with the checks that cover it, and a reasoned account of every edit that has no check.

## Localization
Use `--lang` to view check descriptions in a supported language:
```
oasdiff checks changelog --lang ru
```
Supported languages: `en` (default), `ru`, `pt-br`, `es`.

Only the changelog listing takes `--lang`. The validate rule descriptions are not localized.

## Validate checks
`oasdiff checks validate` lists the rules `oasdiff validate` reports, with the same `--format` and `--severity` flags:
```
oasdiff checks validate
oasdiff checks validate --severity error --format json
```
It takes no `--tags` (validate rules carry none) and no `--lang`.

## Using Check IDs
Each check has a unique ID (e.g. `api-path-removed-without-deprecation`) which can be used to:
- [Ignore specific changes](BREAKING-CHANGES.md#ignoring-specific-breaking-changes)
- [Customize severity levels](BREAKING-CHANGES.md#customizing-severity-levels)
- [Write custom checks](CUSTOMIZING-CHECKS.md)

