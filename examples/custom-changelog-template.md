# API Changes {{ .GetVersionTitle }}

{{ range $group, $changes := .GroupedChanges }}
## {{ if $group.Operation }}{{ $group.Operation }} {{ end }}{{ $group.Path }}

{{ range $changes }}
- {{ if .IsBreaking }}🚨 **BREAKING CHANGE**: {{ else }}📝 {{ end }}{{ .Text }}
{{ end }}

{{ end }}