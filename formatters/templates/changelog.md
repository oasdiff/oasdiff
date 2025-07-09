{{ range $endpoint, $changes := .APIChanges }}
#### `{{ $endpoint.Operation }}` {{ $endpoint.Path }}
{{ range $changes }}- {{ if .IsBreaking }}`!WARNING!`{{ end }} {{ .Text }}
{{ end }}
{{ end }}
