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
- `effect` — the check's verdict about the set of payloads the contract accepts: `widens`, `narrows`, `incomparable` (the change both rejects payloads that were valid and accepts payloads that were not), `unknown` (the check cannot tell), `violation` (breaks the deprecation/stability contract rather than the wire contract), or `none` (metadata with no effect on accepted payloads). Together with `direction`, the effect determines the default severity: narrowing requests and widening responses break clients.
- `direction` — `request`, `response`, or `none`.

## Filtering by Tag
Use `--tags` to show only checks in a specific area, kind, action, effect, or direction:
```
oasdiff checks changelog --tags request,parameters
oasdiff checks changelog --tags schema,constraints
```

Available tags, by dimension:

- direction: `request`, `response`
- action: `add`, `remove`, `change`, `increase`, `decrease`, `set`, `unset`
- effect: `widens`, `narrows`
- area: `schema`, `parameters`, `requestBody`, `responses`, `paths`, `headers`, `security`, `tags`, `components`
- kind: `existence`, `requiredness`, `mutability`, `type`, `constraints`, `values`, `structure`, `lifecycle`

Values of the same dimension are combined with OR, different dimensions with AND: `--tags request,response,add` selects checks that are (request or response) and add.

## Coverage Map
`oasdiff checks changelog coverage` lists every possible edit of an OpenAPI document with what the audit decided about it, one row per edit:

- `covered` — the checks that claim the edit
- `waived` — no check; the `category` field says why: `open` (a missing check, with its reason and a suggested id), `resolved-at-usage` (component definitions are compared at their referencing operations, which have their own rows), or `covered-as` (the same document edit is reported under another action)
- `uncovered` — no check and no waiver (the build fails in this state, so the listing is normally empty)
- `non-contract` — the edit cannot change which payloads are valid, so no check is expected

Available tags, by dimension:

- status: `covered`, `uncovered`, `waived`, `non-contract`
- polarity: `request`, `response`, `document` (neither wire direction), `shared` (a component, whose direction depends on the referencing site)
- action: `add`, `remove`, `change`, `increase`, `decrease`, `set`, `unset`

Values of the same dimension are combined with OR, different dimensions with AND, as in the changelog listing. `--format text|json|yaml` picks the output.

```
oasdiff checks changelog coverage --tags waived,request
oasdiff checks changelog coverage --tags covered,add,remove
```

`--patterns` summarizes the same accounting instead of listing it: one row per waiver or non-contract entry, with the number of edits it accounts for, which answers which reasons the unchecked edits fall under without reading thousands of rows. It takes no `--tags`, having no edits to filter.

```
oasdiff checks changelog coverage --patterns

KIND         PATTERN                                       EDITS REASON
waiver       webhooks.**                                   2353  webhooks are diffed (WebhooksDiff) but checkers only ...
non-contract **.servers.**                                 84    server URLs are deployment metadata, not part of the ...
```

The same accounting is enforced by tests, so the listing is always current. The `locations` field in `checks changelog --format json|yaml` shows the inverse mapping: each check's claimed edits.

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

