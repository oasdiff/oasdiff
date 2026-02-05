## Diff
The `oasdiff diff` command displays the diff between OpenAPI specs.  
Output is fully detailed, typically in yaml or json but also available in text, markdown and html formats.  
This commmand is typically used to generate a structured diff report which can be consumed by other tools but it can also be viewed by humans.

### Output Formats
The default diff output format is `yaml`.  
Additional formats can be generated using the `--format` flag:
- yaml: includes all diff details
- json: includes all diff details
- text: designed to be more user-friendly and provide only the most important parts of the diff (same as markdown)
- markdown: designed to be more user-friendly and provide only the most important parts of the diff (same as text)
- html: designed to be more user-friendly and provide only the most important parts of the diff (see also [changelog with html](BREAKING-CHANGES.md#output-formats))

Notes: 
- an empty `yaml` or `json` result signifies that the diff is empty, or, in other words, there are no changes.  
- the `json` format excludes the `endpoints` section to avoid the [complex mapping keys problem](#complex-mapping-keys).

### Preventing Changes
A common way to use `oasdiff diff` is by running it as a step the CI/CD pipeline to detect changes.  
In order to prevent changes, `oasdiff diff` can be configured to return an error if changes are found.  
To exit with return code 1 if any changes are found, add the `--fail-on-diff` flag.  

### Paths vs. Endpoints
OpenAPI Specification has a hierarchical model of [Paths](https://swagger.io/specification/#paths-object) and [Operations](https://swagger.io/specification/#operation-object) (HTTP methods).  
Oasdiff follows this hierarchy and displays a hierarchical diff with path changes: added, deleted and modified, and within the latter, "modified" section, another set of operation changes: added, deleted and modified. For example:
```yaml
paths:
    deleted:
        - /register
        - /subscribe
    modified:
        /api/{domain}/{project}/badges/security-score:
            operations:
                modified:
                    GET:
```
Oasdiff also outputs an alternate simplified diff per "endpoint" which is a combination of Path + Operation, for example:
```yaml
endpoints:
    deleted:
        - method: POST
          path: /subscribe
        - method: POST
          path: /register
    modified:
        ?   method: GET
            path: /api/{domain}/{project}/badges/security-score
        :   tags:
                deleted:
                    - security
```

### Complex Mapping Keys
The modified endpoints section has two items per key, method and path, this is called a [complex mapping key](https://stackoverflow.com/questions/33987316/what-is-a-complex-mapping-key-in-yaml) in YAML.  
Some YAML libraries don't support complex mapping keys, for exampple:
- python PyYAML: see https://github.com/oasdiff/oasdiff/issues/94#issuecomment-1087468450
- golang gopkg.in/yaml.v3 fails to unmarshal the oasdiff output. This package offers a solution: https://github.com/tliron/yamlkeys

To overcome this limitation, oasdiff allows you to exclude the endpoints section by adding the following flag: `--exclude-elements=endpoints`.  
When using `json` output format, oasdiff excludes `endpoints` automatically.

### OpenAPI Extensions
Oasdiff tracks changes to [OpenAPI extensions](https://swagger.io/docs/specification/openapi-extensions/) by default. To disable this, see [Excluding Specific Kinds of Changes](#excluding-specific-kinds-of-changes).  
The diff format for OpenAPI extensions conforms with [JavaScript Object Notation (JSON) Patch](https://datatracker.ietf.org/doc/html/rfc6902#section-4.4f), for example:
```
endpoints:
    modified:
        ?   method: POST
            path: /example/callback
        :   extensions:
                modified:
                    x-amazon-apigateway-integration:
                        - oldValue: "201"
                          value: "200"
                          op: replace
                          from: ""
                          path: /responses/default/statusCode
                        - oldValue: http://api.example.com/v1/example/callback
                          value: http://api.example.com/v1/example/calllllllllback
                          op: replace
                          from: ""
                          path: /uri
```

### Excluding Specific Kinds of Changes
You can use the `--exclude-elements` flag with to exclude one or more of the following:
- Use `--exclude-elements examples` to exclude [Examples](https://swagger.io/specification/#example-object)
- Use `--exclude-elements extensions` to exclude [Extensions](https://swagger.io/specification/#specification-extensions)
- Use `--exclude-elements description` to exclude description fields
- Use `--exclude-elements title` to exclude title fields
- Use `--exclude-elements summary` to exclude summary fields
- Use `--exclude-elements endpoints` to exclude the [endpoints section of the diff](#paths-vs-endpoints)

For example, this diff excludes descriptions and examples:
```
oasdiff diff data/openapi-test1.yaml data/openapi-test3.yaml --exclude-elements description,examples -f text
```

### Excluding Specific Extension Names
If you want to exclude specific [OpenAPI extensions](https://swagger.io/docs/specification/openapi-extensions/) by name while keeping others, use the `--exclude-extensions` flag.
This is different from `--exclude-elements extensions` which excludes ALL extensions.

For example, to exclude only `x-internal` and `x-ignore` extensions while keeping all others:
```
oasdiff diff base.yaml revision.yaml --exclude-extensions x-internal,x-ignore
```

This is useful when you have extensions that are only used internally or for tooling purposes (e.g., `x-codegen-ignore`, `x-internal`) and you don't want changes to these extensions to appear in the diff.

### Additional Options
- [Merging AllOf Schemas](ALLOF.md)
- [Merging common parameters from the path level into the operation level](COMMON-PARAMS.md)
- [Filtering endpoints](FILTERING-ENDPOINTS.md)
- [Path parameter renaming](PATH-PARAM-RENAME.md)
- [Case-insensitive header comparison](HEADER-DIFF.md)
- [Comparing multiple specs](COMPOSED.md)
- [Customize with configuration files](CONFIG-FILES.md)
- [Running from docker](DOCKER.md)
- [Embedding in your go program](GO.md)

### Usage Examples

#### Diff as YAML
```
oasdiff diff data/openapi-test1.yaml data/openapi-test2.yaml
```
The default diff output format is `yaml`.  
No output means that the diff is empty, or, in other words, there are no changes.

#### Text/Markdown Diff Report
```
oasdiff diff data/openapi-test1.yaml data/openapi-test2.yaml -f text
```
The text diff report provides a simplified and partial view of the changes. It is also compatible with markdown.  
To view all diff details, use `yaml` or `json` formats.

#### HTML Diff Report
```
oasdiff diff data/openapi-test1.yaml data/openapi-test2.yaml -f html 
```
The html diff report provides a simplified and partial view of the changes.  
To view all diff details, use `yaml` or `json` formats.

#### Comparing remote files over http/s
```
oasdiff diff https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text
```

#### Diff across multiple specs with globs
```
oasdiff diff "data/composed/base/*.yaml" "data/composed/revision/*.yaml" -c
```
