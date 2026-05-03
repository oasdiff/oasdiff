# API Changelog {{ .GetVersionTitle }}
{{ if .GroupedChanges }}
{{ with pathGroups .GroupedChanges }}
## API Changes
{{ range . }}
### {{ .Group.Operation }} {{ .Group.Path }}
{{ range .Changes }}- {{ if .IsBreaking }}:warning:{{ end }} {{ .Text }}
{{ end }}
{{ end }}
{{ end }}
{{ range sectionGroups .GroupedChanges }}
## {{ capitalize .Group.Section }}
{{ range .Changes }}- {{ if .IsBreaking }}:warning:{{ end }} {{ .Text }}
{{ end }}
{{ end }}
{{ else }}
{{ if .DiffEmpty }}No changes detected{{ else if .IsBreaking }}No breaking changes to report, but the specs are different{{ else }}No changes to report, but the specs are different{{ end }}
{{ end }}
