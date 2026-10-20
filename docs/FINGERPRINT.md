# Change Fingerprints

Each changelog entry reported by oasdiff includes a `fingerprint` — a short, stable identifier derived from the content of the change.

## What it is

A fingerprint is a 12-character hex string computed as:

```
SHA256("{id}:{operation}:{path}:{args}")[:12]
```

Where `id` is the rule ID (e.g. `response-success-status-removed`), `operation` is the HTTP method, `path` is the API path, and `args` are the values the change's message is built from, joined by `;`. The rendered text is not an input, so a change keeps its fingerprint when its message is reworded or translated.

## Why it's useful

The fingerprint is **stable across commits** — the same breaking change in a PR gets the same fingerprint regardless of which commit introduced it. This makes it possible to:

- **Track review decisions** across commits: if a reviewer approves a breaking change on commit A, the approval can be carried forward when commit B is pushed and the same change is still present
- **Deduplicate changes** when comparing results from multiple runs
- **Reference specific changes** in external systems (CI, review tools, audit logs) without storing the full change text

## When the base changes

A fingerprint stays the same as long as the base spec does. When the base changes, for example when a pull request is updated from a `main` that has moved on, the same change can get a new fingerprint, and a review decision stored against the old one no longer matches:

- **Arguments taken from the base.** Some messages name a value from the base spec. A type change reads `changed from integer to string`, so if the base type changes, the fingerprint does too, although the pull request still changes the type to `string`.
- **Positions.** A `oneOf`, `anyOf` or `allOf` branch that is not a `$ref` is named by its position, as in `oneOf[subschema #4]`. If the base gains a branch before it, a change in that branch gets a new path.
- **Renames.** If the base renames an API path or a property, every change below it gets a new fingerprint.
- **Schemas used by several properties.** A change in such a schema is reported at the first property in alphabetical order. If the base gains a property that sorts earlier and uses the same schema, the change is reported there instead.

In each case the change returns to unreviewed rather than keeping a decision that may no longer apply.

## Output formats

Fingerprints appear in JSON and YAML output:

### JSON (`-f json`)

```json
{
  "id": "response-success-status-removed",
  "text": "removed the success response with status '200'",
  "level": 3,
  "operation": "GET",
  "path": "/users/{id}",
  "section": "paths",
  "fingerprint": "a3f8c21b9d04"
}
```

### YAML (`-f yaml`)

```yaml
- id: response-success-status-removed
  text: removed the success response with status '200'
  level: 3
  operation: GET
  path: /users/{id}
  section: paths
  fingerprint: a3f8c21b9d04
```

## Usage in Go

When using oasdiff as a Go library, the fingerprint is available on each `formatters.Change`:

```go
import (
    "github.com/oasdiff/oasdiff/formatters"
    "github.com/oasdiff/oasdiff/checker"
)

changes := formatters.NewChanges(checkerChanges, localizer)
for _, c := range changes {
    fmt.Printf("change %s fingerprint: %s\n", c.Id, c.Fingerprint)
}
```

## Collision probability

The fingerprint is 12 hex characters (48 bits). For a typical PR with tens of breaking changes, the collision probability is negligible. If two changes happen to share a fingerprint, they are treated as the same change — in practice this will not occur for realistic API diffs.
