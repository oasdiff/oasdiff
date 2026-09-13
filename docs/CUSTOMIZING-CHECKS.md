# How to Add Breaking-Changes Checks

## First: Ask the Coverage Map

Before writing anything, ask the audit what it already knows about the edit you want to check:

```
oasdiff checks changelog coverage
```

One row per possible edit of an OpenAPI document (filter with `--tags`, see [CHECKS.md](CHECKS.md#coverage-map)):

- `covered`: checks already claim this edit; nothing to add.
- `waived`: no check yet, and the reason why. An `open` waiver is an invitation: it carries a suggested id in the house grammar, and its reason in [checker/coverage/waivers.go](../checker/coverage/waivers.go) often points at the tracking issue.
- `non-contract`: the edit cannot change which payloads are valid; no check is expected.

### Example: locating a specific edit

Say you want to know whether changing a query parameter's `style` is checked. Every edit is a location in the OpenAPI document plus an action, so grep the map for the location:

```
$ oasdiff checks changelog coverage | grep 'paths.*.parameters.*.style'
paths.*.*.parameters.*.style    change   waived   request-parameter-style-changed
paths.*.*.parameters.*.style    set      waived   request-parameter-style-set
paths.*.*.parameters.*.style    unset    waived   request-parameter-style-unset
```

`waived` means not implemented, and the last column is the suggested id for whoever implements it. The `json` format adds the reason:

```
$ oasdiff checks changelog coverage --format json | \
    jq '.[] | select(.location == "paths.*.*.parameters.*.style" and .action == "change")'
{
  "location": "paths.*.*.parameters.*.style",
  "action": "change",
  "polarity": "request",
  "status": "waived",
  "category": "open",
  "reason": "parameter serialization style changes the wire format but is unchecked (tracked in #1164)",
  "suggestedId": "request-parameter-style-changed"
}
```

So: not implemented, deliberately recorded as a gap, tracked in #1164, and the id to use is already chosen. Compare an edit that is implemented:

```
$ oasdiff checks changelog coverage | grep 'requestBody.content.*.schema.maximum '
paths.*.*.requestBody.content.*.schema.maximum   set   covered   request-body-max-set,request-property-max-set
```

`covered` names the checks that claim the edit; run `oasdiff checks changelog` and look them up, or grep the [checker](../checker) package for the id, to see how they behave.

## Second: Is the Check Generated?

Some check families are **generated from tables**, not written as functions, and the set grows over time. A check whose id a generator produces must not be written by hand: extend the generating table instead, and the generated rules pick up every direction, scope, and action at once, with levels derived from the severity law and messages from per-locale templates.

Currently generated: setting, unsetting, increasing, and decreasing the ordered constraint keywords (`boundSpecs` in [checker/bound_rules.go](../checker/bound_rules.go)). To cover a new constraint keyword: add a row to `diff.SchemaBounds` in [diff/schema_bounds.go](../diff/schema_bounds.go) (the keyword with its absence encoding), add a `boundSpecs` row (id segment, keyword, polarity: `lowerBound` narrows on increase, `upperBound` on decrease, `unordered` gets no increase/decrease rules), run `make bound-messages` and `make localize`, and update the pinned counts the failing tests name.

The gates enforce the boundary in both directions: `TestBoundCellsFire` builds a spec pair for every generated cell and requires exactly one change at the registered level, and `TestHandWrittenBoundIdsMatchTheGrammar` fails a hand-written check whose id covers a generated cell under a different format.

Everything else, checks that need a judgment call, stays hand-written as follows.

## Write the Check Function
1. Create a new go file under [checker](../checker), named after the use case, for example `check_request_property_became_nullable.go`.
2. Define the check ids as constants at the top of the file. Each id is a unique kebab-case string, for example `request-property-became-nullable`. Related ids (request/response, body/property, added/removed) live in the same file.
3. Write the check function. Use the existing plumbing rather than walking the diff by hand:
   - `walkModifiedRequestBodySchemas` / `walkModifiedResponseSchemas` call your function once per modified media type with a `mediaTypeInfo`; its `walkProperties` method visits every modified property under that media type (including inside allOf/oneOf/anyOf/items). Build changes with `info.newChange` at the body level and `p.newChange` at the property level.
   - Operation-level checks iterate the diff directly and build changes with `opInfo.NewApiChange`.
   - A change computed from a schema node should carry that node: the walkers do it automatically; parameter checks chain `.WithSchema(schemaDiff)`. This lets a recognized schema transition (for example, a schema wrapped in a nullable `oneOf`) claim the change so one transition is reported once, not once per raw field it touches. See [checker/transition_claims.go](../checker/transition_claims.go).

## Register the Rules
Add one rule per id to `GetAllRules()` in [checker/rules.go](../checker/rules.go):

```go
newBackwardCompatibilityRule(RequestParameterBecameNotNullableId, ERR, RequestParameterBecameNullableCheck, DirectionRequest, AreaParameters, KindRequiredness, ActionChange),
```

- **Level**: `ERR` for breaking changes, `WARN` for potentially breaking, `INFO` for backward compatible. Requests and responses are usually contravariant: what is breaking on one side is often the opposite action on the other (adding a required request property breaks clients; on the response side it is removing a property that does).
- **Direction / Area / Kind / Action** classify the rule in the taxonomy. `TestRuleSymmetry` audits it: a rule with no mirror across an axis fails the build unless the asymmetry is waived with a reason in [checker/rule_symmetry_test.go](../checker/rule_symmetry_test.go). When you add a check, add its mirror too, or record why it is intentionally absent.

## Localized Messages
1. For each id, add two keys in **all four locales** under [checker/localizations_src](../checker/localizations_src) (en, es, pt-br, ru): the message (`request-parameter-became-not-nullable: the %s request parameter %s became not nullable`) and the description (`request-parameter-became-not-nullable-description: ...`), which the `oasdiff checks changelog` command and the rule catalog display. If a translation isn't ready, copy the English text as a placeholder so the key exists.
2. Run `make localize` to regenerate [checker/localizations/localizations.go](../checker/localizations/localizations.go).

## Tests
1. Add a unit test in a file mirroring your check file's name, with OpenAPI spec fixtures under [data](../data). Assert the exact set of reported ids, not just presence, so unrelated findings fail the test.
2. Bump `numOfChecks` and `numOfIds` in [checker/config_test.go](../checker/config_test.go).

## Example
A complete, current example, new ids, registration, four-locale messages and descriptions, fixtures and tests: [#1093](https://github.com/oasdiff/oasdiff/pull/1093).
