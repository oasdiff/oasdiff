# Schemas Used in Several Places
A schema referenced from more than one property of the same payload is reported once per check, at one of the properties that reach it. The finding names the shared schema and lists the other properties the change is at:

```
the `customerId` response's property pattern `^[0-9]+$` was added for the status `200` (shared schema: `Id`, also at `userId`)
```

So the change is under `components/schemas/Id` and it affects both properties: `customerId` and `userId`. The same comparison always reports the change at the same property.

The text lists up to three properties and counts the rest. When shared schemas are nested, every combination counts. Say `Inner` has a property `id`, `Outer` references `Inner` from its properties `a` and `b`, and the payload references `Outer` from its properties `p` and `q`. A pattern added to `id` is at four properties, `p/a/id`, `q/a/id`, `p/b/id` and `q/b/id`:

```
the `p/a/id` response's property pattern `^[0-9]+$` was added for the status `200` (shared schema: `Inner`, also at `q/a/id`, `p/b/id` and 1 more)
```

JSON and YAML output carry the same list and the full count:

```json
"sharedSchema": {
  "name": "Inner",
  "properties": ["p/a/id", "q/a/id", "p/b/id"],
  "count": 4
}
```

Why one finding: the number of properties a change is at multiplies with each level of nesting. In the Microsoft Graph API description (v1.0, from the APIs.guru directory), a single schema is reached from 164,769 properties of one request or response body. A finding per property would make the changelog impossible to read and the comparison slow, so oasdiff reports the change once and counts the rest.

When shared schemas are nested, the finding names the one closest to the change: a change to `id` names `Inner`, and a property added to `Outer` names `Outer`.

The shared schema does not change the change's fingerprint, which is computed from its arguments and not from its text.

Three things stay separate, and each still gets its own line:

- **Each check.** Every check reports on its own, so a shared schema that both added a property and gained a `pattern` produces one `response-optional-property-added` and one `response-property-pattern-added`.
- **Each change.** Two properties added to the same shared schema are two changes, both reported at the same property.
- **Each operation.** An operation is a separate contract, so an operation that reaches the same changed schema reports it too.
