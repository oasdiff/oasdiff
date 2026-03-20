# OpenAPI Changelog with Custom Template

You can customize the changelog output format by providing a custom template file when using `markdown` or `html` format:

```bash
oasdiff changelog base.yaml revision.yaml --template my-template.md -f markdown
oasdiff changelog base.yaml revision.yaml --template my-template.html -f html
```

## Template Data

Templates have access to the following data and functions:

### Data Fields

| Field | Description |
|---|---|
| `.GroupedChanges` | Changes grouped by endpoint (for API changes) or by section (for security/component changes) |
| `.BaseVersion` | Base spec version string |
| `.RevisionVersion` | Revision spec version string |
| `.GetVersionTitle()` | Formatted version comparison string, e.g. `1.0.0 vs. 2.0.0` |

### Template Functions

| Function | Description |
|---|---|
| `pathGroups .GroupedChanges` | Returns API path changes sorted by path and operation |
| `sectionGroups .GroupedChanges` | Returns security and component changes sorted by section name |
| `capitalize "string"` | Capitalizes the first letter of a string |

Each entry returned by `pathGroups` and `sectionGroups` has:
- `.Group.Path` — API path (e.g. `/pets`)
- `.Group.Operation` — HTTP method (e.g. `GET`)
- `.Group.Section` — Section name for non-path groups (e.g. `security`, `components`)
- `.Changes` — Slice of changes, each with `.Text`, `.IsBreaking`, and `.Comment`

## Example Template

A full example template is available at [`examples/custom-changelog-template.md`](../examples/custom-changelog-template.md).

Here is a minimal example:

```markdown
# API Changelog {{ .GetVersionTitle }}

{{ if .GroupedChanges }}
{{ with pathGroups .GroupedChanges }}
## API Changes
{{ range . }}
### {{ .Group.Operation }} {{ .Group.Path }}
{{ range .Changes }}- {{ if .IsBreaking }}:warning: **BREAKING**: {{ end }}{{ .Text }}
{{ end }}
{{ end }}
{{ end }}

{{ range sectionGroups .GroupedChanges }}
## {{ capitalize .Group.Section }}
{{ range .Changes }}- {{ if .IsBreaking }}:warning: **BREAKING**: {{ end }}{{ .Text }}
{{ end }}
{{ end }}
{{ else }}
No changes
{{ end }}
```

## Notes

- `pathGroups` and `sectionGroups` return sorted slices, ensuring consistent output order across runs.
- If there are no path changes, the `API Changes` section is omitted entirely.
- Each section group (e.g. Security, Components) only appears if it has at least one change.
- Use `{{ if .GroupedChanges }}` to display a "No changes" message when the changelog is empty.
