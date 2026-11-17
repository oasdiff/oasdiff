# Breaking Changes and Changelog
As your API evolves, it undergoes changes. Some of these changes may be "breaking" while others are not.  
The `oasdiff breaking` command displays the changes that break existing API clients.  
The `oasdiff changelog` command displays the changes that can affect API consumers, both breaking and non-breaking. Documentation-only edits, such as descriptions and examples, are not included; use `oasdiff diff` to see every difference in the API definition.  
These commands are typically used in the CI to report or prevent breaking changes.

> **Note:** `breaking` and `changelog` run on the diff engine described in [DIFF.md](DIFF.md). The flags and concepts there — extension tracking, path matching, `allOf` flattening, the path-prefix family, header case, `--allow-external-refs`, `--fail-on-diff` — apply here too, in addition to the breaking-specific options below.

> **For teams:** [oasdiff.com](https://www.oasdiff.com) wraps these commands with a per-change PR comment, one-click approve/reject buttons, and commit-status checks — so reviewers can act on each breaking change directly from the pull request.

## Example: display breaking changes
```
oasdiff breaking https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

## Example: display a changelog
```
oasdiff changelog https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

## Checks
Oasdiff supports hundreds of checks (run `oasdiff checks changelog` for the current list, or browse the full catalog at [oasdiff.com/docs/breaking-changes](https://www.oasdiff.com/docs/breaking-changes)), categorized into three levels:  
- `ERR` - Errors are definite breaking changes which should be avoided
- `WARN` - Warnings are potential breaking changes which developers should be aware of, but cannot be confirmed programmatically as breaking
- `INFO` - Non-breaking changes

`oasdiff breaking` detects changes with level `ERR` and `WARN` only.  
`oasdiff changelog` detects changes with levels that are greater or equal to the `--level` argument. The default level, `INFO`, includes all checks.

To see the full list of checks and their descriptions, run:
```
oasdiff checks changelog
```
See also [Customizing Severity Levels](SEVERITY-LEVELS.md)

## Preventing Breaking Changes
A common way to use oasdiff is by running it as a step the CI/CD pipeline to detect changes.  
In order to prevent changes, oasdiff can be configured to return an error if changes above a certain level are found.
- To exit with return code 1 if ERR-level changes are found, add the `--fail-on ERR` flag.  
- To exit with return code 1 if ERR-level or WARN-level changes are found, add the `--fail-on WARN` flag.
- To exit with return code 1 if any changes are found, add the `--fail-on INFO` flag to `changelog`. `breaking` doesn't accept `INFO`, because it reports `WARN` and above.
- To exit with return code 1 only when breaking changes aren't accompanied by a major version increase in `info.version`, add the `--fail-on MISSING-VERSION-BUMP` flag, described in [Version Bumps and Breaking Changes](VERSIONING.md).

For example:
```
oasdiff breaking --fail-on ERR data/openapi-test1.yaml data/openapi-test3.yaml
```

## Topics

### Running the commands
- [Side-by-side review in your browser](SIDE-BY-SIDE-REVIEW.md): `--open` uploads the comparison, encrypted, and opens it as a side-by-side review you can share.
- [Output formats](OUTPUT-FORMATS.md): text, json, yaml, html, markdown, junit and more, the JSON Schema of the output, color, and other languages.
- [Checking many specs at once](BREAKING-FILES.md): one comparison per changed spec against a git ref, as the pre-commit hook does.

### What is reported, and at which level
- [How oasdiff decides what is breaking](WHAT-IS-BREAKING.md): a change is judged against the contract the definition declares, not against what a server happens to accept.
- [The list of checks](CHECKS.md): every check with its level, and how the checks are categorized.
- [When a change is reported below its check's level](SEVERITY-LEVELS.md#when-oasdiff-reports-a-change-below-its-checks-level): an inexact comparison inside an `allOf`, or a `readOnly` or `writeOnly` property that the change cannot affect.
- [Schemas used in several places](SHARED-SCHEMAS.md): a change to a shared schema is reported once and lists the properties it is at.

### Rules for specific kinds of change
- [Deprecating APIs](DEPRECATION.md): remove an endpoint without a breaking-change error by deprecating it first.
- [API stability levels](STABILITY.md): let draft, alpha and beta APIs change more freely than stable ones.
- [Version bumps](VERSIONING.md): `--fail-on MISSING-VERSION-BUMP` fails the run when a breaking change has no major version increase.
- [Nullability changes](NULLABILITY.md): the ways a schema can allow `null`, and when changing them is breaking.
- [Enum values](EXTENSIBLE-ENUM.md): `x-extensible-enum` lets a response enum gain values without a breaking change.

### Adjusting the result
- [Customizing severity levels](SEVERITY-LEVELS.md): raise, lower or disable a check.
- [Ignoring specific breaking changes](IGNORING-CHANGES.md): list the changes you accept in a file.
- [Custom checks](CUSTOMIZING-CHECKS.md): add a check of your own when a change is not reported.

## Known Limitations
- no checks for `context` instead of `schema` for request parameters
- no checks for `callback`s
