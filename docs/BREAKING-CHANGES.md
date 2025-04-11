## Breaking Changes and Changelog
As your API evolves, it undergoes changes. Some of these changes may be "breaking" while others are not.  
The `oasdiff breaking` command displays the breaking changes between OpenAPI specifications.  
The `oasdiff changelog` command displays all significant changes between OpenAPI specifications, including breaking and non-breaking changes.  
These commands are typically used in the CI to report or prevent breaking changes.

### Example: display breaking changes
```
oasdiff breaking https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

### Example: display a changelog
```
oasdiff changelog https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

### Checks
Oasdiff supports over 250 checks, categorized into three levels:  
- `ERR` - Errors are definite breaking changes which should be avoided
- `WARN` - Warnings are potential breaking changes which developers should be aware of, but cannot be confirmed programmatically as breaking
- `INFO` - Non-breaking changes

`oasdiff breaking` detects changes with level `ERR` and `WARN` only.  
`oasdiff changelog` detects changes with levels that are greater or equal to the `--level` argument. The default level, `INFO`, includes all checks.

To see the full list of checks and their descriptions, run:
```
oasdiff checks
```
See also [Customizing Severity Levels](#customizing-severity-levels)

### Preventing Breaking Changes
A common way to use oasdiff is by running it as a step the CI/CD pipeline to detect changes.  
In order to prevent changes, oasdiff can be configured to return an error if changes above a certain level are found.
- To exit with return code 1 if ERR-level changes are found, add the `--fail-on ERR` flag.  
- To exit with return code 1 if ERR-level or WARN-level changes are found, add the `--fail-on WARN` flag.
- To exit with return code 1 if any changes are found, add the `--fail-on INFO` flag.

For example:
```
oasdiff breaking --fail-on ERR data/openapi-test1.yaml data/openapi-test3.yaml
```

### Output Formats
By default, oasdiff displays changes in a human-readable [colorized](#color) text format.  
Additional formats can be generated using the `--format` flag:
- json
- yaml
- githubactions: suitable for integration with github
- junit: suitable for integration with gitlab
- html: [see example](https://html-preview.github.io/?url=https://github.com/oasdiff/oasdiff/blob/main/docs/changelog.html)
- markdown: [see example](changelog.md)
- text: the default, human-readable, format
- singleline: displays each change on a single line, this can be useful to prepare [ignore files](#ignoring-specific-breaking-changes)

For example:
```
oasdiff breaking -f yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

### Color
When outputting changes to a Unix terminal, oasdiff automatically adds colors with ANSI color escape sequences.  
If output is piped into another process or redirected to a file, oasdiff disables color.  
To control color manually, use the `--color` flag with `always` or `never`.

### API Stability Levels
Assigning [stability levels](STABILITY.md) to APIs allows fine-grained control over how APIs are allowed to change based on their maturity.  

### Deprecating APIs
Before deleting an endpoint, it is recommended to give consumers a heads-up in the form of "deprecation". 
Oasdiff allows you to [deprecate APIs gracefully](DEPRECATION.md) without triggering a breaking-change error.

### Ignoring Specific Breaking Changes
Sometimes, you want to allow certain breaking changes, for example, when your spec and service are out-of-sync and you need to correct the spec.  
Oasdiff allows you define breaking changes that you want to ignore in a configuration file.  
You can specify the configuration file name in the oasdiff command-line with the `--warn-ignore` flag for WARNINGS or the `--err-ignore` flag for ERRORS.  
Each line in the configuration file should contain two parts:
1. Method and path (the first field in the line beginning with slash) to ignore a change to an endpoint, or the keyword 'components' to ignore a change in components
2. Description of the breaking change

For example:
```
GET /api/{domain}/{project}/badges/security-score removed the success response with the status '200'
```
Or, for a component change:
```
components removed the schema 'rules'
```

The required parts may appear in any order, in lower or upper case, and the configuration line may contain additional text, like this:
```
 - 12.01.2023 In GET /api/{domain}/{project}/badges/security-score we removed the success response with the status '200'
 - 31.10.2023 Removed the schema 'network-policies' from components
```

The configuration files can be of any text type, e.g., Markdown, so you can use them to document breaking changes and other important changes.

### Breaking Changes to Enum Values
Oasdiff supports special rules for enum changes using the `x-extensible-enum` extension.  
This method allows adding new entries to enums used in responses which is very usable in many cases but requires clients to support a fallback to default logic when they receive an unknown value.
`x-extensible-enum` was introduced by [Zalando](https://opensource.zalando.com/restful-api-guidelines/#112) and picked up by the OpenAPI community. Technically, it could be replaced with anyOf+classical enum but the `x-extensible-enum` is a more explicit way to do it.  
In most cases the `x-extensible-enum` is similar to enum values, except it allows adding new entries in messages sent to the client (responses or callbacks).
If you don't use the `x-extensible-enum` in your OpenAPI specifications, nothing changes for you, but if you do, oasdiff will identify breaking changes related to `x-extensible-enum` parameters and properties.

### Localization
To display changes in other languages, use the `--lang` flag.  
Currently English, Russian and Brazilian Portuguese are supported.  
[Please improve oasdiff by adding your own language](https://github.com/oasdiff/oasdiff/issues/383).

### Customizing Severity Levels
Oasdiff allows you to change the default severity levels according to your needs.  
For example, the default severity level of the `api-security-removed` check is `INFO`. You can verify this by running `oasdiff checks`.  
To change the `api-security-removed` check's severity level to `ERR` use the following command:
```
oasdiff changelog data/checker/api_security_added_revision.yaml data/checker/api_security_added_base.yaml --severity-levels oasdiff-levels.txt
```
Where the file `oasdiff-levels.txt` contains a single line:
```
api-security-removed    err
```

Checks can be customized with the following levels:
| Custom Level  | Check Status |
| ------------- | ------------- |
| err  | Enabled with level ERR |
| warn  | Enabled with level WARN |
| info  | Enabled with level INFO |
| none  | Disabled  |

### Customizing Breaking Changes Checks
If you encounter a change that isn't reported, you may:
1. Run `oasdiff checks` to see if the check is available, and [customize the level as needed](#customizing-severity-levels).  
2. Add a [custom check](CUSTOMIZING-CHECKS.md)

### Additional Options
- [Merging AllOf Schemas](ALLOF.md)
- [Merging common parameters from the path level into the operation level](COMMON-PARAMS.md)
- [Filtering endpoints](FILTERING-ENDPOINTS.md)
- [Path parameter renaming](PATH-PARAM-RENAME.md)
- [Case-insensitive header comparison](HEADER-DIFF.md)
- [Comparing multiple specs](COMPOSED.md)
- [Adding OpenAPI Extensions to the changelog output](ATTRIBUTES.md)
- [Customize with configuration files](CONFIG-FILES.md)
- [Running from docker](DOCKER.md)
- [Embedding in your go program](GO.md)

### Known Limitations
- no checks for `context` instead of `schema` for request parameters
- no checks for `callback`s
