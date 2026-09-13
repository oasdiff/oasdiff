# How to Add Breaking-Changes Checks

## First: Is the Check Generated?

Checks for setting, unsetting, increasing, and decreasing the ordered constraint keywords (`maximum`, `minimum`, `maxLength`, `minLength`, `maxItems`, `minItems`, `maxProperties`, `minProperties`, `minContains`, `maxContains`, `exclusiveMinimum`, `exclusiveMaximum`, and `multipleOf` for set/unset) are **generated from a table**, not written as functions: one row in `boundSpecs` in [checker/bound_rules.go](../checker/bound_rules.go) produces the rules for every direction (request/response), scope (body, property, parameter, response header), and action, with levels derived from the severity law and messages derived from per-locale templates.

To cover a new constraint keyword:

1. Add a row to `diff.SchemaBounds` in [diff/schema_bounds.go](../diff/schema_bounds.go) so the diff exposes the keyword with its absence encoding (nil, zero, or false as absent).
2. Add a row to `boundSpecs` with the id segment, the keyword, and its polarity (`lowerBound`: increasing narrows; `upperBound`: decreasing narrows; `unordered`: no increase/decrease rules).
3. Run `make bound-messages` (message templates fill in the keyword per locale) and `make localize`.
4. Update the pinned counts the tests name when they fail: the rule count in [checker/bound_rules_test.go](../checker/bound_rules_test.go), the cell count in [checker/bound_fire_test.go](../checker/bound_fire_test.go), and `numOfIds` in [checker/config_test.go](../checker/config_test.go).

Every generated cell is exercised by `TestBoundCellsFire`, which builds a spec pair per cell and requires exactly one change with the right id and level. If you hand-write a check whose id covers a cell the generator would produce, it must use the generated id format so the generator skips that cell; `TestHandWrittenBoundIdsMatchTheGrammar` fails otherwise.

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
