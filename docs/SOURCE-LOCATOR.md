# Source Location Tracking (Beta)

oasdiff can now correlate each breaking change and changelog entry with the exact line and column in the OpenAPI spec file where the change occurred.

This enables inline annotations on GitHub pull requests, pointing reviewers directly to the relevant lines.

This feature is currently in **beta** and available for community testing before general availability.

## What it does

When source location tracking is enabled, each reported change includes:
- **RevisionSource**: file, line, and column in the revised (new) spec
- **BaseSource**: file, line, and column in the base (old) spec

Source locations are available in:
- **JSON** (`-f json`) and **YAML** (`-f yaml`) output as `baseSource` and `revisionSource` fields on each change
- **GitHub Actions** (`-f githubactions`) as inline annotations on the "Files changed" tab of PRs, using `RevisionSource` for the annotation location

### Precision levels

Source locations are reported at the most specific level available:
- **Field-level**: points to the exact field that changed (e.g., the `type:` line when a type changes)
- **Sequence item-level**: points to a specific item in a list (e.g., a particular enum value or required property name)
- **Sub-object-level**: points to the schema, parameter, or response that changed

### Removal-type changes

Changes that represent a removal (e.g., endpoint removed, property removed) only have a `BaseSource` since the removed element doesn't exist in the revision file. These appear in the GitHub Actions log and check summary but not as inline annotations on the diff.

## How to try it

### Docker

```bash
docker pull tufin/oasdiff:source-location-tracking

# Changelog with GitHub Actions annotations
docker run --rm -v $(pwd):/data:ro -w /data tufin/oasdiff:source-location-tracking changelog base.yaml revision.yaml -f githubactions

# Changelog with JSON output (includes baseSource/revisionSource fields)
docker run --rm -v $(pwd):/data:ro -w /data tufin/oasdiff:source-location-tracking changelog base.yaml revision.yaml -f json
```

### GitHub Actions workflow

```yaml
name: API breaking changes
on: pull_request

jobs:
  oasdiff:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 2

    - name: Get previous spec
      run: |
        PREV=$(git rev-parse HEAD~1)
        echo "PREV_URL=https://raw.githubusercontent.com/${{ github.repository }}/$PREV/openapi.yaml" >> $GITHUB_ENV

    - name: Run oasdiff changelog
      run: |
        docker run --rm -v $(pwd):/data:ro -w /data \
          tufin/oasdiff:source-location-tracking \
          changelog ${{ env.PREV_URL }} openapi.yaml -f githubactions
```

Annotations appear inline on the "Files changed" tab of the pull request.

### Build from source

```bash
git clone -b source-location-tracking https://github.com/oasdiff/oasdiff.git
cd oasdiff
go build
./oasdiff changelog base.yaml revision.yaml -f githubactions
```

## Example output

### GitHub Actions annotations (inline on PR diff)

Additions and modifications appear inline:
```
::error title=new-required-request-property,file=openapi.yaml,line=30,col=13::in API POST /users added the new required request property 'email'
```

Removals appear in the Actions log (no inline annotation):
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

See [oasdiff/github-demo](https://github.com/oasdiff/github-demo) for a working example with inline PR annotations.

## Provide feedback

We need your feedback to move this to general availability.

### Found an issue?
[Open an issue](https://github.com/oasdiff/oasdiff/issues/new?template=bug_report.md&title=[source-location]%20) with `[source-location]` in the title.

### It works for you?
Please add a thumbs-up reaction to [this tracking issue](https://github.com/oasdiff/oasdiff/issues/574) and optionally leave a comment describing your use case.

## Status

| Milestone | Status |
|-----------|--------|
| Source tracking infrastructure | Done |
| All 93 checkers report locations | Done |
| GitHub Actions formatter integration | Done |
| Community testing | **In progress** |
| General availability | Pending testing feedback |
