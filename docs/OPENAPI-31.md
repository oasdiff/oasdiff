# OpenAPI 3.1 Support

oasdiff supports OpenAPI 3.1 specs across all commands: `diff`, `breaking`, and `changelog`.
The [OpenAPI 3.2](#openapi-32) section below covers the 3.2 additions oasdiff handles.

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

## OpenAPI 3.2

Support is scoped to the 3.2 features listed here, not to the whole of 3.2.

### Streamed bodies (`itemSchema`)

A 3.2 media type can carry an `itemSchema` describing **one item** of a streamed body, such as a
single Server-Sent Event or one JSON Lines record, as opposed to `schema`, which describes the whole
body.

```yaml
paths:
  /events:
    get:
      responses:
        '200':
          content:
            application/jsonl:
              itemSchema:        # one record, not the whole stream
                type: object
                properties:
                  kind: { type: string, enum: [created, updated] }
```

Every check that runs against a body schema also runs against an item schema, so a change to a
streamed item's contract is reported like any other body change. Messages carry an `(item schema)`
marker to say which of the two they refer to:

```
error   [request-property-became-required]
          the request property `weight` became required (item schema)
```

An item schema appearing or disappearing is reported by its own rules, since that is not a
modification of an existing schema:

| id | level |
|---|---|
| `request-body-media-type-item-schema-added` | error (the request narrowed) |
| `request-body-media-type-item-schema-removed` | info |
| `response-body-media-type-item-schema-added` | info |
| `response-body-media-type-item-schema-removed` | warning |
| `response-body-media-type-item-schema-removed-untyped` | error |

The two removal rules differ by what is left behind. If the media type still declares a whole-body
`schema`, the response stays typed and only the per-item guarantee is gone, which is a warning: the
definition cannot say whether a consumer depended on the item shape. If nothing else types the
body, it is an error.

If you are on an earlier version, changes to a streamed item's contract are missed entirely: the
item schema is not read, so a required property added to a streamed request item, which breaks every
producer, is reported as no change at all.

## Migrating from 3.0 to 3.1

OpenAPI 3.1 changes the shape of several constructs:

| 3.0 | 3.1 |
|---|---|
| `nullable: true` on `type: string` | `type: ["string", "null"]` |
| `exclusiveMinimum: true` + `minimum: 0` | `exclusiveMinimum: 0` (numeric) |
| `example: 7` | `examples: [7]` |

A team migrating their spec from 3.0 to 3.1 has two problems oasdiff can help with: producing the migrated spec, and verifying the migration doesn't break clients.

### Converting a spec with `oasdiff upgrade`

The `upgrade` subcommand rewrites a 3.0 spec into the latest 3.x canonical form (currently 3.2.0). The transforms are idempotent, so running it on an already-canonical spec just bumps the version string.

```
oasdiff upgrade old-spec.yaml > new-spec.yaml
oasdiff upgrade old-spec.yaml --format json > new-spec.json
cat old-spec.yaml | oasdiff upgrade -
```

The output goes to stdout; redirect to a file to keep it. The default output format is `yaml`; pass `--format json` for JSON.

The walker handles 3.0 → 3.x only. Swagger 2.0 → 3.0 is out of scope.

Available since `v1.16.0`.

### Comparing across versions with `--auto-upgrade`

`diff`, `breaking`, `changelog`, and `summary` accept an `--auto-upgrade` flag that runs the same canonicalisation on both specs in-memory right after load. This makes cross-version comparisons (e.g. a 3.0 base against a 3.1 revision) produce a meaningful result instead of a noisy diff dominated by dialect-level differences.

```
# old.yaml is 3.0, new.yaml is 3.1
oasdiff breaking  old.yaml new.yaml --auto-upgrade
oasdiff changelog old.yaml new.yaml --auto-upgrade
oasdiff diff      old.yaml new.yaml --auto-upgrade
oasdiff summary   old.yaml new.yaml --auto-upgrade
```

Without the flag, the diff surfaces the 3.0→3.1 dialect rewrites (`nullable` becomes `type: ["string", "null"]`, etc.) as if they were schema changes. With the flag, both sides are canonicalised first, so only the genuine schema-level differences remain.

The flag is off by default; opt in per invocation. Safe to set even when both specs are already the same version: the walker is idempotent on already-canonical input.

Available since `v1.16.0`.

## Caveats

The following 3.1 features are not yet fully supported by [`kin-openapi`](https://github.com/getkin/kin-openapi) (the parser oasdiff uses) and therefore do not appear in oasdiff diffs:

- **`$dynamicRef` / `$dynamicAnchor`**: parsed but not resolved during loading. Schemas using dynamic references for recursive definitions will not be followed.
- **`pathItems` in `components`**: not represented in the parser's data model. Specs that declare reusable path items in components will silently drop them.

If you hit any of these, please [open an issue](https://github.com/oasdiff/oasdiff/issues/new?template=bug_report.md&title=[3.1]%20).

## Feedback

Found an issue? [Open one here](https://github.com/oasdiff/oasdiff/issues/new?template=bug_report.md&title=[3.1]%20) with `[3.1]` in the title.
