# Nullability Changes

A schema can allow the value `null` in three equivalent ways:

1. The `nullable` keyword (OpenAPI 3.0):
   ```yaml
   type: string
   nullable: true
   ```
2. A `"null"` entry in the type array (OpenAPI 3.1):
   ```yaml
   type: [string, "null"]
   ```
3. A `oneOf` with a null branch. This is the common way to make a `$ref`'d schema nullable, since a `$ref` cannot carry a type array:
   ```yaml
   oneOf:
     - type: "null"
     - $ref: '#/components/schemas/Address'
   ```

oasdiff treats all three forms as the same thing: adding any of them is reported as `became-nullable` and removing any of them as `became-not-nullable`.

This matters most for the `oneOf` form. Wrapping a schema in a `oneOf` moves it one level down, so a naive comparison sees a type change, an enum removal and a `oneOf` addition, and reports some of them as breaking. oasdiff recognizes that the wrapped schema is equivalent to the original and reports a single nullability change instead:

```
oasdiff changelog data/checker/nullable_wrap_base.yaml data/checker/nullable_wrap_revision.yaml

info    [request-property-became-nullable]
        in API POST /test
                the request property `optionalEnum` became nullable
```

## A nullability change inside a `oneOf` branch

When the `oneOf` is in both specs and one of its branches changes nullability, the branch does not decide what the schema around it accepts. A `oneOf` matches only when exactly one of its branches matches, so the schema accepts `null` only while exactly one branch does:

```yaml
address:
  oneOf:
    - type: "null"
    - $ref: '#/components/schemas/Address'
```

Making `Address` nullable gives this property two branches that accept `null`, so a `null` value matches both and is rejected. The property stopped accepting `null`, the opposite of what the edited branch did, so oasdiff reports the property rather than the branch:

```
oasdiff changelog data/checker/one_of_null_branch_base.yaml data/checker/one_of_null_branch_revision.yaml

error   [request-property-became-not-nullable]
        in API POST /products
                the request property `address` became not nullable
                A `oneOf` matches only when exactly one of its branches matches, so it accepts null only while exactly one branch does. Adding null to a second branch takes null out of the composition, and removing it from one of two puts it back.
```

An `anyOf` matches when any of its branches matches, so a null branch keeps the schema nullable whatever the other branches do. A nullability change in one of them leaves the schema accepting `null` as before and is not reported.

## Breaking or not?
Whether a nullability change is breaking depends on the direction of the data:

- A **request** is sent by the client. A request schema that becomes nullable accepts everything it accepted before, plus `null`: not breaking. A request schema that becomes not nullable rejects `null` values that were valid before: breaking.
- A **response** is returned by the server. A response schema that becomes nullable may return `null` to clients that don't expect it: breaking. A response schema that becomes not nullable only guarantees more than before: not breaking.

|                     | Request | Response |
|---------------------|---------|----------|
| became nullable     | info    | error    |
| became not nullable | error   | info     |

For example, removing the wrapper from the request schemas above is breaking:

```
oasdiff changelog data/checker/nullable_wrap_revision.yaml data/checker/nullable_wrap_base.yaml

error   [request-property-became-not-nullable]
        in API POST /test
                the request property `optionalEnum` became not nullable
```

Parameters follow the request column: `request-parameter-became-nullable` (info) and `request-parameter-became-not-nullable` (error), with `request-parameter-property-*` variants for properties of object parameters.

Response headers follow the response column: `response-header-became-nullable` (error) and `response-header-became-not-nullable` (info).

## Limitations
oasdiff recognizes the `oneOf` form only when it is a pure wrap: exactly two branches, one of them just `{type: "null"}`, and the other equivalent to the original schema. When the edit is more than a nullability change, oasdiff falls back to reporting the underlying changes, erring on the side of reporting too much rather than missing a breaking change:

- **Wrapping an already-nullable schema** is not a nullability change. Under `oneOf`, a value must match exactly one branch; if the schema already allows `null`, a `null` value now matches two branches and is rejected.
- **A wrap that also changes the schema**, for example removing an enum value while adding the wrapper, is reported by the underlying changes (the enum removal), not as a nullability change.
- **Adding a null branch to a `oneOf` that already has several branches** is reported as a list-of-types change (for simple types) or a `oneOf` change (for objects), not as a nullability change.
- **A composition oasdiff cannot decide**: an `allOf` or a `not` alongside the `oneOf` or `anyOf` constrains the value too, and a branch that composes further has no nullability of its own. In both cases oasdiff cannot tell what the composition accepts, and reports the branch's own nullability change instead.
