# Source Location Tracking

oasdiff can correlate breaking changes and changelog entries with the exact line and column in the OpenAPI spec file where the change occurred.  
This enables inline annotations on GitHub pull requests, pointing reviewers directly to the relevant lines.

## What it does

When source location tracking is enabled, reported changes include:

- **BaseSource**: file, line, and column in the base (old) spec
- **RevisionSource**: file, line, and column in the revised (new) spec

Changes that represent a removal (e.g., endpoint removed, property removed) will have a `BaseSource` only since the removed element doesn't exist in the revision file.  
Likewise, Changes that represent an addition (e.g., endpoint added, new enum value) will have a `RevisionSource` only since the added element doesn't exist in the base file.   
Changes that represent a modification (e.g., maxLength increased) will have both `BaseSource` and `RevisionSource`.

### Multi-file specs

When a spec uses `$ref` to import schemas from other YAML or JSON files, source locations point to the file where the changed element actually lives, with the precise line and column. The `file` field on `BaseSource` and `RevisionSource` reflects the imported file, not just the top-level spec entry point, and inline GitHub Actions annotations use that path.

## Output formats

Source locations are available in:
- **JSON** (`-f json`) and **YAML** (`-f yaml`) output as `baseSource` and `revisionSource` fields on each change
- **GitHub Actions** (`-f githubactions`) output includes `revisionSource` fields to use as inline annotations on the "Files changed" tab of PRs. Note that `baseSource` isn't displayed because GitHub can only display annotations on the latest version of the file.

### Precision levels

Source locations are reported at the most specific level available:
- **Field-level**: points to the exact field that changed (e.g., the `type:` line when a type changes)
- **Sequence item-level**: points to a specific item in a list (e.g., a particular enum value or required property name)
- **Sub-object-level**: points to the schema, parameter, or response that changed

## Example output

### GitHub Actions annotations

Additions and modifications appear inline:
```
::error title=new-required-request-property,file=openapi.yaml,line=30,col=13::in API POST /users added the new required request property 'email'
```

Removals appear without inline annotation, because GitHub only attaches Actions annotations to the revision (head) version of files, not to the base. `BaseSource` is still emitted in `-f json` and `-f yaml` output and points at the exact line in the base spec where the removed element used to live.

```
::error title=request-property-removed::in API POST /users removed the request property 'name'
```

### JSON output (`-f json`)

```json
{
  "id": "request-property-type-changed",
  "level": 3,
  "operation": "POST",
  "path": "/users",
  "baseSource": {"file": "base.yaml", "line": 42, "column": 15},
  "revisionSource": {"file": "revision.yaml", "line": 42, "column": 15}
}
```

### YAML output (`-f yaml`)

```yaml
- id: request-property-type-changed
  level: 3
  operation: POST
  path: /users
  baseSource:
    file: base.yaml
    line: 42
    column: 15
  revisionSource:
    file: revision.yaml
    line: 42
    column: 15
```

## Demo

**Try it visually:** [oasdiff.com/diff](https://oasdiff.com/diff) shows a side-by-side comparison of any two OpenAPI specs in your browser, no installation required.
