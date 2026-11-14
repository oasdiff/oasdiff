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

When a spec uses `$ref` to import schemas from other files, source locations point to the file where the changed element actually lives, with the precise line and column. The `file` field on `BaseSource` and `RevisionSource` reflects the imported file, not just the top-level spec entry point, and inline GitHub Actions annotations use that path.

## Limitations

JSON and YAML specs both carry source locations. A JSON spec reports the line and column of the changed field like a YAML one, including fields on a single line, where the column is what tells them apart.

### Minified JSON

Positions are only as fine-grained as the document's line breaks. A pretty-printed JSON spec, one element per line, reports each change on the element's own line, like a YAML spec. A minified spec, the whole document on one line, reports every change on line 1 with a column that is still exact, so GitHub Actions annotations all land on the first line and a side-by-side review shows the whole document as the block for every change. Run the spec through `jq .` before comparing it, which is the pretty-printed form.

### Changes located at the enclosing element

A change oasdiff cannot tie to one field reports the nearest enclosing element that has a location, which is the schema, parameter or operation containing it. The line is then the start of that element rather than the line that changed.

## Output formats

`-f` selects the format oasdiff writes its report in. It is unrelated to the format the spec itself is written in.

Source locations are available in:
- **JSON** (`-f json`) and **YAML** (`-f yaml`) output as `baseSource` and `revisionSource` fields on each change
- **GitHub Actions** (`-f githubactions`) output includes `revisionSource` fields to use as inline annotations on the "Files changed" tab of PRs. Note that `baseSource` isn't displayed because GitHub can only display annotations on the latest version of the file.

See [Limitations](#limitations) for the changes that have no line to report.

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
