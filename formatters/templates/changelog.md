# API Changelog {{ .GetVersionTitle }}
{{ range $group, $changes := .GroupedChanges }}
## {{ if $group.Operation }}{{ $group.Operation }} {{ end }}{{ $group.Path }}
{{ range $changes }}- {{ if .IsBreaking }}:warning:{{ end }} {{ .Text }}
{{ end }}
{{ end }}
