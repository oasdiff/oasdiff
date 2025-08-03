# API Changes {{ .GetVersionTitle }}

{{ range $endpoint, $changes := .APIChanges }}
## {{ $endpoint.Operation }} {{ $endpoint.Path }}

{{ range $changes }}
- {{ if .IsBreaking }}🚨 **BREAKING CHANGE**: {{ else }}📝 {{ end }}{{ .Text }}
{{ end }}

{{ end }}