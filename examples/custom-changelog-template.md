# API Changelog {{ .GetVersionTitle }}

{{ if .GroupedChanges }}
{{ with pathGroups .GroupedChanges }}
## API Changes

{{ range . }}
### {{ .Group.Operation }} {{ .Group.Path }}

{{ range .Changes }}
- {{ if .IsBreaking }}🚨 **BREAKING CHANGE**: {{ else }}📝 {{ end }}{{ .Text }}
{{ end }}

{{ end }}
{{ end }}

{{ range sectionGroups .GroupedChanges }}
## {{ .Group.Section }}

{{ range .Changes }}
- {{ if .IsBreaking }}🚨 **BREAKING CHANGE**: {{ else }}📝 {{ end }}{{ .Text }}
{{ end }}

{{ end }}
{{ else }}
No changes
{{ end }}
