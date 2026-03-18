# API Changelog {{ .GetVersionTitle }}
{{ range $group, $changes := .GroupedChanges }}
## {{ if $group.Path }}{{ $group.Operation }} {{ $group.Path }}{{ else }}{{ $group.Section }}{{ end }}
{{ range $changes }}- {{ if .IsBreaking }}:warning:{{ end }} {{ .Text }}
{{ end }}
{{ end }}
