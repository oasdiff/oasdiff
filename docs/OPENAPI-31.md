# OpenAPI 3.1 Support

oasdiff supports OpenAPI 3.1 specs across all commands: `diff`, `breaking`, and `changelog`.

OpenAPI 3.1 support is generally available starting with `v1.15.0`. Previous beta tags (`v1.15.0-openapi31.beta.*` and `v2.2.0-openapi31.beta.*`) are superseded.

## What's supported

### Spec loading
- OpenAPI 3.1.0 document parsing
- JSON Schema 2020-12 keywords
- Webhooks

### Diff
Changes are detected for all 3.1-specific fields:
- `exclusiveMinimum`/`exclusiveMaximum` as numeric values
- `const`, `prefixItems`, `contains`, `minContains`, `maxContains`
- `if`/`then`/`else` conditional schemas
- `dependentRequired`, `dependentSchemas`
- `patternProperties`, `propertyNames`
- `unevaluatedItems`, `unevaluatedProperties` (both boolean and schema forms)
- `contentSchema`, `contentMediaType`, `contentEncoding`
- `$defs`, `$schema`, `$comment`
- Webhooks (added/deleted/modified)
- Info `summary`, License `identifier`

### Breaking changes and changelog
162 new rule IDs covering:
- **Nullable type arrays**: `type: ["string", "null"]` correctly detected as nullable changes (not false-positive type changes)
- **Webhooks**: added/removed detection + all existing operation-level checks automatically applied to modified webhook operations
- **const**: added/removed/changed for request and response body/properties
- **exclusiveMinimum/exclusiveMaximum**: increased/decreased/set for request body/properties/parameters and response body/properties
- **prefixItems**: added/removed for request and response
- **if/then/else**: added/removed for request and response
- **contains/minContains/maxContains**: added/removed/increased/decreased
- **dependentRequired**: added/removed/changed
- **dependentSchemas**: added/removed
- **patternProperties**: added/removed
- **propertyNames**: added/removed
- **unevaluatedItems/unevaluatedProperties**: added/removed
- **contentSchema/contentMediaType/contentEncoding**: added/removed/changed

## Caveats

The following 3.1 features are not yet fully supported by [`kin-openapi`](https://github.com/getkin/kin-openapi) (the parser oasdiff uses) and therefore do not appear in oasdiff diffs:

- **`$dynamicRef` / `$dynamicAnchor`**: parsed but not resolved during loading. Schemas using dynamic references for recursive definitions will not be followed.
- **`pathItems` in `components`**: not represented in the parser's data model. Specs that declare reusable path items in components will silently drop them.

The `flatten` command has one additional caveat:

- **`flatten` of `allOf` ignores 3.1 numeric `exclusiveMinimum` / `exclusiveMaximum`**: when merging `allOf` subschemas, the merge logic considers only `minimum` / `maximum` and the 3.0-style boolean `exclusiveMinimum` / `exclusiveMaximum`. 3.1 numeric exclusive bounds are silently dropped from the merged result. Tracked in [#868](https://github.com/oasdiff/oasdiff/issues/868).

If you hit any of these, please [open an issue](https://github.com/oasdiff/oasdiff/issues/new?template=bug_report.md&title=[3.1]%20).

## Feedback

Found an issue? [Open one here](https://github.com/oasdiff/oasdiff/issues/new?template=bug_report.md&title=[3.1]%20) with `[3.1]` in the title.
