# Change Fingerprints

Each changelog entry reported by oasdiff includes a `fingerprint` — a short, stable identifier derived from the content of the change.

## What it is

A fingerprint is a 12-character hex string computed as:

```
SHA256("{id}:{operation}:{path}:{text}")[:12]
```

Where `id` is the rule ID (e.g. `response-success-status-removed`), `operation` is the HTTP method, `path` is the API path, and `text` is the human-readable change description.

## Why it's useful

The fingerprint is **stable across commits** — the same breaking change in a PR gets the same fingerprint regardless of which commit introduced it. This makes it possible to:

- **Track review decisions** across commits: if a reviewer approves a breaking change on commit A, the approval can be carried forward when commit B is pushed and the same change is still present
- **Deduplicate changes** when comparing results from multiple runs
- **Reference specific changes** in external systems (CI, review tools, audit logs) without storing the full change text

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
