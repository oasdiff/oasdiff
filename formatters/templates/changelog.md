# API Changelog {{ .GetVersionTitle }}
{{ range $endpoint, $changes := .APIChanges }}
## {{ $endpoint.Operation }} {{ $endpoint.Path }}
{{ range $changes }}- {{ if .IsBreaking }}:warning:{{ end }} {{ .Text }}
{{ end }}
{{ end }}
