## Usage Examples

### OpenAPI diff of local files in YAML
```bash
oasdiff diff data/openapi-test1.yaml data/openapi-test2.yaml
```
The default diff output format is YAML.  
No output means that the diff is empty, or, in other words, there are no changes.

### OpenAPI diff of local files in Text/Markdown 
```bash
oasdiff diff data/openapi-test1.yaml data/openapi-test2.yaml -f text
```
The Text/Markdown diff report provides a simplified and partial view of the changes.  
To view all details, use the default format: YAML.  
If you'd like to see additional details in the text/markdown report, please submit a [feature request](https://github.com/oasdiff/oasdiff/issues/new?assignees=&labels=&template=feature_request.md&title=).

### OpenAPI diff of local files in HTML
```bash
oasdiff diff data/openapi-test1.yaml data/openapi-test2.yaml -f html 
```
The HTML diff report provides a simplified and partial view of the changes.  
To view all details, use the default format: YAML.  
If you'd like to see additional details in the HTML report, please submit a [feature request](https://github.com/oasdiff/oasdiff/issues/new?assignees=&labels=&template=feature_request.md&title=).


### OpenAPI diff for remote files over http/s
```bash
oasdiff diff https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text
```

### Breaking changes
```bash
oasdiff breaking https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

### Breaking changes as YAML
```bash
oasdiff breaking https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f yaml
```

### Breaking changes across multiple specs with globs
```bash
oasdiff breaking "data/composed/base/*.yaml" "data/composed/revision/*.yaml" -c
```

### Fail with exit code 1 if any ERR-level breaking changes are found
```bash
oasdiff breaking "data/composed/base/*.yaml" "data/composed/revision/*.yaml" -c -o ERR
```

### Fail with exit code 1 if any change is found
```bash
oasdiff diff https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text -o
```

### OpenAPI changelog
```bash
oasdiff changelog https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

### OpenAPI changelog with custom template
```bash
oasdiff changelog base.yaml revision.yaml --template my-template.md -f markdown
oasdiff changelog base.yaml revision.yaml --template my-template.html -f html
```
You can customize the changelog output format by providing a custom template file when using markdown or html format.  
The template uses Go's text/template format and has access to the following data:
- `.APIChanges` - map of endpoints to their changes
- `.BaseVersion` - base spec version
- `.RevisionVersion` - revision spec version  
- `.GetVersionTitle()` - formatted version comparison string

Example custom template:
```markdown
### Changes {{ .GetVersionTitle }}
{{ range $endpoint, $changes := .APIChanges }}
#### {{ $endpoint.Operation }} {{ $endpoint.Path }}
{{ range $changes }}* {{ if .IsBreaking }}**BREAKING**: {{ end }}{{ .Text }}
{{ end }}
{{ end }}
```

### OpenAPI diff for endpoints containing "/api" in the path
```bash
oasdiff diff https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text -p "/api"
```
Filters are applied recursively at all levels. For example, if a path contains a [callback](https://swagger.io/docs/specification/callbacks/), the filter will be applied both to the path itself and to the callback path. To include such a nested change, use a regular expression that contains both paths, for example ```-filter "path|callback-path"```

### Exclude paths and operations with extension "x-beta"
```bash
oasdiff diff https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text --filter-extension "x-beta"
``` 
Notes:
1. [OpenAPI Extensions](https://swagger.io/docs/specification/openapi-extensions/) can be defined both at the [path](https://swagger.io/docs/specification/paths-and-operations/) level and at the [operation](https://swagger.io/docs/specification/paths-and-operations/) level. Both are matched and excluded with this flag.
2. If a path or operation has a matching extension only in one of the specs, but not in the other, it will appear as Added or Deleted.

### Ignore changes to description and examples
```bash
oasdiff diff data/openapi-test1.yaml data/openapi-test3.yaml --exclude-elements description,examples -f text
``` 

### Display change summary
```bash
oasdiff summary https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

### OpenAPI Diff with Docker
To run with docker just replace the `oasdiff` command by `docker run --rm -t oasdiff/oasdiff`, for example:

```bash
docker run --rm -t tufin/oasdiff diff https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text
```

### Breaking changes with Docker
```bash
docker run --rm -t tufin/oasdiff breaking https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

### Comparing local files with Docker
```bash
docker run --rm -t -v $(pwd)/data:/data:ro tufin/oasdiff diff /data/openapi-test1.yaml /data/openapi-test3.yaml
```

Replace `$(pwd)/data` by the path that contains your files.  
Note that the spec paths must begin with `/`.  