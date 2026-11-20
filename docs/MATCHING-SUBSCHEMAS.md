# How oasdiff compares subschemas

A subschema under `allOf`, `anyOf` or `oneOf` has no name to match on, so oasdiff pairs the base and revision lists in steps, each one over whatever the earlier steps left unmatched:

1. **By `$ref`** - two subschemas that reference the same component.
2. **By identical content** - two inline subschemas with no difference at all between them, annotations included.
3. **By title** - two inline subschemas with the same non-empty `title`.
4. **By validation contract** - two untitled inline subschemas that accept the same payloads.
5. **By elimination** - one untitled inline subschema left on each side, paired whatever their contracts, on the assumption that one was edited into the other.
6. **Across the inline/`$ref` boundary** - an inline subschema and a `$ref` to a component that accept the same payloads. `anyOf` and `oneOf` only, and `--match-inline-refs=false` turns it off.

Whatever differs inside a matched pair is reported as a modified subschema. Anything left unmatched is reported as added or deleted. So an edit confined to annotations, a `description` inside several branches at once for instance, is reported where it happened rather than as branches removed and added.

Steps 4 and 6 compare validation contracts: annotation-only differences (`title`, `description`, `default`, `example`, `examples`, external docs, `$comment`) do not prevent the match and are reported inside the pair. Differences that affect validation, `deprecated` included, prevent it.

## Inline and `$ref` Subschemas

Under `anyOf` and `oneOf`, oasdiff treats an inline subschema and a `$ref` to a validation-equivalent component as the same branch by default. This catches a common codegen pattern where a tool extracts an inline schema (often an enum) into a named component without changing the accepted payloads:

```yaml
# base
role:
  anyOf:
    - type: string
      enum: [user, admin]
    - type: "null"

# revision (UserRole is { type: string, enum: [user, admin] })
role:
  anyOf:
    - $ref: "#/components/schemas/UserRole"
    - type: "null"
```

The diff reports no change for the `anyOf` list because the union of accepted payloads is identical. The same behaviour applies to `breaking` and `changelog`, which run on the same diff engine.

To restore the previous behaviour where the inline branch is reported as removed and the `$ref` branch as added, pass `--match-inline-refs=false`.

Component `$ref` renames (both sides `$ref`, e.g. `UserRoleV1` to `UserRole`) are out of scope by design: only the inline-to-`$ref` form change is matched this way.
