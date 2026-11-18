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

### How to run
- **Stop a breaking change before it is committed.** See [Pre-commit hook](PRE-COMMIT-HOOK.md).
- **Run without installing anything.** See [Docker](DOCKER.md).
- **Keep the flags in one place** in an `.oasdiff.yaml` file, or point at one with `--config`. See [Configuration file](CONFIG-FILES.md).
- **Pipe the output to another command, or redirect it to a file.** Color is turned off when the output is not a terminal, and `--color` forces it on or off. See [Color, pipes and redirects](COLOR.md).

### Git and GitHub
- **Compare by git refs**, a branch, tag or commit, without checking out files: `oasdiff breaking origin/main:openapi.yaml openapi.yaml`. Add `--fetch` when the commit is not in your clone. See [Git revisions](GIT-REVISION.md).
- **Check each spec that changed against its own previous version**, when a repository holds several independent APIs: `oasdiff breaking-files --base origin/main` makes one comparison per spec. See [Check many independent specs](BREAKING-FILES.md).
- **Check every pull request** with the [GitHub Action](https://github.com/oasdiff/oasdiff-action). It comments on the pull request with the list of changes, each linked to its side-by-side diff.
- **Compare a pull request with its base branch** in a workflow: `base: 'origin/${{ github.base_ref }}:openapi.yaml'`. The checkout step needs `fetch-depth: 0`. See [Git revisions in GitHub Actions](GIT-REVISION.md#github-actions).
- **Mark the changed lines in the "Files changed" tab**: `--format githubactions` writes an annotation on the line of each change. The action does this for you. See [GitHub Actions annotations](SOURCE-LOCATOR.md#github-actions-annotations).
- **Use the same settings locally and in the workflow**: the action reads `.oasdiff.yaml` from the repository. See [Using a configuration file with the GitHub Action](CONFIG-FILES.md#using-with-the-github-action).
- **Accept pull requests from forks**: the action does not load external `$ref`s unless you turn that on, so a spec in a fork cannot make the runner fetch a URL or read a file. See [Security](#security).
- **Block the merge until every breaking change is approved**: with [oasdiff Pro](https://www.oasdiff.com/pricing), reviewers approve or reject each change, and a status check on the pull request follows.

### Security
- **Run on pull requests from people you do not trust**: `--allow-external-refs=false` stops a spec from making the runner fetch a URL or read a local file. See [External `$ref`s](EXTERNAL-REFS.md).

### Inputs
- **Check one API that is split across several files**, such as one spec per service behind a gateway: `--composed` merges the files into a single comparison. See [Compare APIs split across multiple files](COMPOSED.md).
- **Compare an OpenAPI 3.0 spec with its 3.1 version** without the migration itself showing up as changes: `--auto-upgrade`. See [Comparing across versions](OPENAPI-31.md#comparing-across-versions-with---auto-upgrade).

### Comparison
- **Rename a path parameter without it counting as a new endpoint.** This is the default, and `--include-path-params` turns it off. See [Endpoint matching](MATCHING-ENDPOINTS.md).
- **Move an inline schema into `components` without a false report**: a `oneOf` or `anyOf` branch and the `$ref` that replaces it are matched as the same branch. `--match-inline-refs=false` turns this off. See [Inline and `$ref` subschemas](DIFF.md#inline-and-ref-subschemas).

### Normalization
- **Get exact results for schemas that use `allOf`**: `--flatten-allof` merges the branches before comparing. Without it a change inside an `allOf` is reported as a warning, because another branch may already describe it. See [Merge `allOf` schemas](ALLOF.md).
- **Handle parameters declared once for a whole path**: `--flatten-params` merges them into each operation before comparing. See [Merge common (path-level) parameters](COMMON-PARAMS.md).
- **Compare an API that moved to a new base path**, such as `/v1` to `/api/v1`: `--prefix-base`, `--prefix-revision`, `--strip-prefix-base` and `--strip-prefix-revision`. See [Path prefix modification](PATH-PREFIX.md).
- **Treat `Content-Type` and `content-type` as the same header.** This is the default, and `--case-insensitive-headers=false` turns it off. See [Case-insensitive header comparison](HEADER-DIFF.md).

### API lifecycle
- **Remove an endpoint, parameter or property** by deprecating it first and giving clients time. `--deprecation-days-stable` and `--deprecation-days-beta` set how much. See [Deprecate APIs and parameters](DEPRECATION.md).
- **Let draft, alpha and beta endpoints change freely** while stable ones are protected: `--stability-level`. See [API stability levels](STABILITY.md).
- **Allow breaking changes only in a new major version**: `--fail-on MISSING-VERSION-BUMP`. See [Version bumps](VERSIONING.md).
- **Add a value to an enum that clients receive** without a breaking change, with `x-extensible-enum`. See [Enum values](EXTENSIBLE-ENUM.md).

### Filtering changes
- **Check only part of the API**, such as the public endpoints: `--match-path`, `--unmatch-path` and `--filter-extension`. See [Filter endpoints](FILTERING-ENDPOINTS.md).
- **Change how severe a change is, or turn a check off**: `--severity-levels`. See [Customize severity levels](SEVERITY-LEVELS.md).
- **Accept a breaking change that you decided to make**, and keep the build green: `--err-ignore` and `--warn-ignore`. See [Ignore specific breaking changes](IGNORING-CHANGES.md).
- **Add a rule oasdiff does not have.** See [Custom checks](CUSTOMIZING-CHECKS.md).

### Output
- **Produce a JUnit report for GitLab or another CI system**: `--format junit`. See [Output formats](OUTPUT-FORMATS.md).
- **Feed the changes to another tool**: `--format json` or `--format yaml`, and `oasdiff schema` for the JSON Schema of that output. See [Output formats](OUTPUT-FORMATS.md).
- **Publish a changelog for the people who use your API**: `--format html` or `--format markdown`, with your own layout through `--template`. See [Customize HTML and Markdown changelog templates](CHANGELOG-TEMPLATE.md).
- **Go from a change to the line that caused it.** Each change carries the file, line and column in both specs, and `--open` shows the two specs side by side on that line. See [Source location tracking](SOURCE-LOCATOR.md) and [Side-by-side review](SIDE-BY-SIDE-REVIEW.md).
- **Show which team owns a changed endpoint**: `--attributes` copies an extension such as `x-audience` from the spec into each change. See [Add OpenAPI-extension attributes to changelog entries](ATTRIBUTES.md).
- **Recognize the same change in a later run**, for example to remember that it was approved. Each change has a stable id. See [Change fingerprints](FINGERPRINT.md).
- **Show the changes in your team's language**: `--lang`. See [Localization](LOCALIZATION.md).

### Reference
- **How oasdiff decides what is breaking.** It judges a change against the contract the spec declares, not against what one server accepts. See [How oasdiff decides what is breaking](WHAT-IS-BREAKING.md).
- **Every change oasdiff can detect, and its level**: `oasdiff checks changelog`, or the catalog at [oasdiff.com/docs/breaking-changes](https://www.oasdiff.com/docs/breaking-changes). See [Checks](CHECKS.md).
- **When making a property nullable breaks clients.** See [Nullability changes](NULLABILITY.md).
- **A change to a schema that many properties use** is reported once, with the list of properties it affects. See [Schemas used in several places](SHARED-SCHEMAS.md).

## Known Limitations
- no checks for `context` instead of `schema` for request parameters
- no checks for `callback`s
