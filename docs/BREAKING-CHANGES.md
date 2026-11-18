# Breaking Changes and Changelog
The `oasdiff breaking` command displays the changes that break existing API clients.  
The `oasdiff changelog` command displays the changes that can affect API consumers, both breaking and non-breaking.  
These commands are typically used in the CI to report or prevent breaking changes.

## Example: display breaking changes
```
oasdiff breaking https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

## Example: display a changelog
```
oasdiff changelog https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

## Example: fail the build on a breaking change
```
oasdiff breaking --fail-on ERR data/openapi-test1.yaml data/openapi-test3.yaml
```
Every change has a level: `ERR` for a definite breaking change, `WARN` for one that cannot be confirmed from the spec, and `INFO` for a non-breaking change.  
`--fail-on ERR` exits with return code 1 when there is an error, and `--fail-on WARN` when there is an error or a warning.  
`breaking` reports `ERR` and `WARN`. `changelog` reports all three, accepts `--fail-on INFO`, and shows only the levels at or above `--level`.

## Example: review the changes side by side, and approve them
```
oasdiff changelog HEAD~1:openapi.yaml HEAD:openapi.yaml --open
```
`--open` prints the changes, then opens a review in your browser with the old and new spec side by side, positioned on each change. The link can be shared with reviewers, who need no install and no account. See [Side-by-side review](SIDE-BY-SIDE-REVIEW.md).  
With [oasdiff Pro](https://www.oasdiff.com/pricing), reviewers approve or reject each breaking change in that review, and a status check on the pull request blocks the merge until all are approved.

## What you can do

### In CI and on pull requests
- **Check every pull request on GitHub** with the [GitHub Action](https://github.com/oasdiff/oasdiff-action), which comments on the pull request with the changes.
- **Mark the changed lines in the pull request** with `--format githubactions`, or produce a `junit` report for GitLab. See [Output formats](OUTPUT-FORMATS.md).
- **Stop a breaking change before it is committed** with the [pre-commit hook](BREAKING-FILES.md#pre-commit-hook).
- **Allow breaking changes only in a new major version** with `--fail-on MISSING-VERSION-BUMP`. See [Version bumps](VERSIONING.md).
- **Run on pull requests from people you do not trust** with `--allow-external-refs=false`, so a spec cannot make the runner fetch a URL or read a file. See [External $refs](EXTERNAL-REFS.md).
- **Keep the flags in one place** in an `.oasdiff.yaml` file, or point at one with `--config`. See [Configuration files](CONFIG-FILES.md).

### Choosing what to compare
- **Compare your working copy with `main`, or any two commits**, without checking out files: `oasdiff breaking origin/main:openapi.yaml openapi.yaml`. Add `--fetch` when the commit is not in your clone. See [Git revisions](GIT-REVISION.md).
- **Check each spec that changed against its own previous version**, when a repository holds several independent APIs: `oasdiff breaking-files --base origin/main` makes one comparison per spec. See [Checking many specs against a git ref](BREAKING-FILES.md).
- **Check one API that is split across several files**, such as one spec per service behind a gateway: `--composed` merges the files into a single comparison. See [Compare APIs split across multiple files](COMPOSED.md).
- **Compare an OpenAPI 3.0 spec with its 3.1 version** without the migration itself showing up as changes: `--auto-upgrade`. See [Comparing across versions](OPENAPI-31.md#comparing-across-versions-with---auto-upgrade).
- **Check only part of the API**, such as the public endpoints: `--match-path`, `--unmatch-path` and `--filter-extension`. See [Filtering endpoints](FILTERING-ENDPOINTS.md).
- **Compare an API that moved to a new base path**, such as `/v1` to `/api/v1`: `--prefix-base`, `--prefix-revision`, `--strip-prefix-base` and `--strip-prefix-revision`. See [Path prefix modification](PATH-PREFIX.md).
- **Rename a path parameter without it counting as a new endpoint.** This is the default, and `--include-path-params` turns it off. See [How oasdiff compares endpoints](MATCHING-ENDPOINTS.md).

### Getting accurate results
- **Get exact results for schemas that use `allOf`**: `--flatten-allof` merges the branches before comparing. Without it a change inside an `allOf` is reported as a warning, because another branch may already describe it. See [Merge allOf schemas](ALLOF.md).
- **Handle parameters declared once for a whole path**: `--flatten-params` merges them into each operation before comparing. See [Common parameters](COMMON-PARAMS.md).
- **Move an inline schema into `components` without a false report**: a `oneOf` or `anyOf` branch and the `$ref` that replaces it are matched as the same branch. `--match-inline-refs=false` turns this off. See [Inline and $ref subschemas](DIFF.md#inline-and-ref-subschemas).
- **Treat `Content-Type` and `content-type` as the same header.** This is the default, and `--case-insensitive-headers=false` turns it off. See [Header comparison](HEADER-DIFF.md).
- **Understand why a change is, or is not, breaking.** oasdiff judges a change against the contract the spec declares, not against what one server accepts. See [How oasdiff decides what is breaking](WHAT-IS-BREAKING.md).
- **Know when making a property nullable breaks clients.** See [Nullability changes](NULLABILITY.md).
- **Read a change to a schema that many properties use.** It is reported once, with the list of properties it affects. See [Schemas used in several places](SHARED-SCHEMAS.md).
- **See everything oasdiff can detect, and at which level**: `oasdiff checks changelog`, or the catalog at [oasdiff.com/docs/breaking-changes](https://www.oasdiff.com/docs/breaking-changes). See [Checks](CHECKS.md).

### Changing an API without breaking clients
- **Remove an endpoint, parameter or property** by deprecating it first and giving clients time. `--deprecation-days-stable` and `--deprecation-days-beta` set how much. See [Deprecating APIs](DEPRECATION.md).
- **Let draft, alpha and beta endpoints change freely** while stable ones are protected: `--stability-level`. See [API stability levels](STABILITY.md).
- **Add a value to an enum that clients receive** without a breaking change, with `x-extensible-enum`. See [Enum values](EXTENSIBLE-ENUM.md).

### Fitting oasdiff to your own policy
- **Change how severe a change is, or turn a check off**: `--severity-levels`. See [Customizing severity levels](SEVERITY-LEVELS.md).
- **Accept a breaking change that you decided to make**, and keep the build green: `--err-ignore` and `--warn-ignore`. See [Ignoring specific breaking changes](IGNORING-CHANGES.md).
- **Add a rule oasdiff does not have.** See [Custom checks](CUSTOMIZING-CHECKS.md).

### Using the output
- **Publish a changelog for the people who use your API**: `--format html` or `--format markdown`, with your own layout through `--template`. See [Changelog templates](CHANGELOG-TEMPLATE.md).
- **Feed the changes to another tool**: `--format json` or `--format yaml`, and `oasdiff schema` for the JSON Schema of that output. See [Output formats](OUTPUT-FORMATS.md).
- **Show which team owns a changed endpoint**: `--attributes` copies an extension such as `x-audience` from the spec into each change. See [Attributes](ATTRIBUTES.md).
- **Go from a change to the line that caused it.** Each change carries the file, line and column in both specs. See [Source location tracking](SOURCE-LOCATOR.md).
- **Recognize the same change in a later run**, for example to remember that it was approved. Each change has a stable id. See [Change fingerprints](FINGERPRINT.md).
- **Read the output in another language, or without color**: `--lang` and `--color`. See [Output formats](OUTPUT-FORMATS.md).

## Known Limitations
- no checks for `context` instead of `schema` for request parameters
- no checks for `callback`s
