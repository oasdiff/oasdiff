# Output Formats for Breaking Changes and Changelog
By default, oasdiff displays changes in a human-readable [colorized](COLOR.md) text format.  
Additional formats can be generated using the `--format` flag:
- json
- yaml
- githubactions: suitable for integration with github
- junit: suitable for integration with gitlab
- html: [see example](https://html-preview.github.io/?url=https://github.com/oasdiff/oasdiff/blob/main/examples/changelog.html)
- markdown: [see example](../examples/changelog.md)
- text: the default, human-readable, format
- singleline: displays each change on a single line, this can be useful to prepare [ignore files](IGNORING-CHANGES.md)

For example:
```
oasdiff breaking -f yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

## JSON Schema
`oasdiff schema` prints a [JSON Schema](https://json-schema.org/) for the `json` output of `breaking` and `changelog` (the `yaml` output has the same structure):
```
oasdiff schema
```
