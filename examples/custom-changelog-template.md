# API Changes {{ .GetVersionTitle }}

{{ range $group, $changes := .GroupedChanges }}
## {{ if $group.Path }}{{ $group.Operation }} {{ $group.Path }}{{ else }}{{ $group.Section }}{{ end }}

{{ range $changes }}
- {{ if .IsBreaking }}🚨 **BREAKING CHANGE**: {{ else }}📝 {{ end }}{{ .Text }}
{{ end }}

{{ end }}