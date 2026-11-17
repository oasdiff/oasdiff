# Output Formats for Breaking Changes and Changelog
The `breaking` and `changelog` commands share these output options.

## Formats
By default, oasdiff displays changes in a human-readable [colorized](#color) text format.  
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

## Color
When outputting changes to a Unix terminal, oasdiff automatically adds colors with ANSI color escape sequences.  
If output is piped into another process or redirected to a file, oasdiff disables color.  
To control color manually, use the `--color` flag with `always` or `never`.

## Localization
To display changes in other languages, use the `--lang` flag.  
Currently English, Russian and Brazilian Portuguese are supported.  
[Please improve oasdiff by adding your own language](https://github.com/oasdiff/oasdiff/issues/383).
