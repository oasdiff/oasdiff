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
Grouped by what you're trying to do.

### How to run
Where and how the commands run.

- [Fail the build on a breaking change](#example-fail-the-build-on-a-breaking-change): stop a breaking change in the CI, with `--fail-on`
- [Pre-commit hook](PRE-COMMIT-HOOK.md): stop a breaking change before it is committed
- [Docker](DOCKER.md): run without installing anything
- [Configuration file](CONFIG-FILES.md): keep the flags in one `.oasdiff.yaml` file, or name one with `--config`
- [Color, pipes and redirects](COLOR.md): pipe or redirect the output, and control color with `--color`

### Git
Compare what is in git.

- [Git revisions](GIT-REVISION.md): compare a branch, tag or commit without checking out files, with `--fetch` for a commit your clone lacks
- [Check many independent specs](BREAKING-FILES.md): check each changed spec against its own previous version, with `oasdiff breaking-files`

### GitHub
Report on pull requests.

- [GitHub Action](https://github.com/oasdiff/oasdiff-action): check every pull request and comment on it with the changes
- [Git revisions in GitHub Actions](GIT-REVISION.md#github-actions): compare a pull request with its base branch
- [GitHub Actions annotations](SOURCE-LOCATOR.md#github-actions-annotations): mark the changed lines in the "Files changed" tab, with `--format githubactions`
- [Pull requests from forks](#security): the action does not load external `$ref`s unless you turn that on
- [oasdiff Pro](https://www.oasdiff.com/pricing): block the merge until reviewers approve every breaking change

### GitLab
Report on merge requests.

- [Output formats](OUTPUT-FORMATS.md): show the changes in the merge request's test report, with `--format junit` and `artifacts:reports:junit`

### Security
Run on specs you do not control.

- [External `$ref`s](EXTERNAL-REFS.md): run on pull requests from people you do not trust, with `--allow-external-refs=false`

### Inputs
Where specs come from.

Both commands take two arguments, `base` and `revision`. Each can be:
- a path to a file, in YAML or JSON: `openapi.yaml`
- an http or https URL: `https://example.com/openapi.yaml`
- a git ref and a path: `origin/main:openapi.yaml` (see [Git revisions](GIT-REVISION.md))
- `-` to read the spec from standard input
- a glob, with `--composed`: `"specs/*.yaml"` (see [Compare APIs split across multiple files](COMPOSED.md))

- [Compare APIs split across multiple files](COMPOSED.md): check one API that is split across several files, with `--composed`
- [Comparing across OpenAPI versions](OPENAPI-31.md#comparing-across-versions-with---auto-upgrade): compare a 3.0 spec with its 3.1 version, with `--auto-upgrade`

### Comparison
How oasdiff pairs up base and revision and what counts as a difference.

- [Endpoint matching](MATCHING-ENDPOINTS.md): rename a path parameter without it counting as a new endpoint (`--include-path-params` turns this off)
- [Subschema matching](MATCHING-SUBSCHEMAS.md): move an inline schema into `components` without a false report (`--match-inline-refs=false` turns this off)
- [Enum values](EXTENSIBLE-ENUM.md): add a value to an enum that clients receive, with `x-extensible-enum`

### Normalization
Align each spec before comparing so equivalent things line up.

- [Merge `allOf` schemas](ALLOF.md): get exact results, not warnings, for schemas that use `allOf`, with `--flatten-allof`
- [Merge common (path-level) parameters](COMMON-PARAMS.md): handle parameters declared once for a whole path, with `--flatten-params`
- [Path prefix modification](PATH-PREFIX.md): compare an API that moved to a new base path, with `--prefix-base`, `--prefix-revision`, `--strip-prefix-base` and `--strip-prefix-revision`
- [Case-insensitive header comparison](HEADER-DIFF.md): treat `Content-Type` and `content-type` as the same header (`--case-insensitive-headers=false` turns this off)

### API lifecycle
Communicate intent across versions.

- [Deprecate APIs and parameters](DEPRECATION.md): remove an endpoint, parameter or property after giving clients time, set with `--deprecation-days-stable` and `--deprecation-days-beta`
- [API stability levels](STABILITY.md): let draft, alpha and beta endpoints change freely, with `--stability-level`
- [Version bumps](VERSIONING.md): allow breaking changes only in a new major version, with `--fail-on MISSING-VERSION-BUMP`

### Filtering changes
Choose which changes are reported, and at which level.

- [Filter endpoints](FILTERING-ENDPOINTS.md): check only part of the API, with `--match-path`, `--unmatch-path` and `--filter-extension`
- [Customize severity levels](SEVERITY-LEVELS.md): change how severe a change is, or turn a check off, with `--severity-levels`
- [Ignore specific breaking changes](IGNORING-CHANGES.md): accept a breaking change you decided to make, with `--err-ignore` and `--warn-ignore`
- [Custom checks](CUSTOMIZING-CHECKS.md): add a rule oasdiff does not have

### Output
Shape and enrich the report.

- [Output formats](OUTPUT-FORMATS.md): feed the changes to another tool, with `--format json` or `yaml`
- [Customize HTML and Markdown changelog templates](CHANGELOG-TEMPLATE.md): publish a changelog for the people who use your API, with `--template`
- [Source location tracking](SOURCE-LOCATOR.md): go from a change to the line that caused it
- [Side-by-side review in your browser](SIDE-BY-SIDE-REVIEW.md): see the two specs side by side on the line of each change, with `--open`
- [Add OpenAPI-extension attributes to changelog entries](ATTRIBUTES.md): show which team owns a changed endpoint, with `--attributes`
- [Change fingerprints](FINGERPRINT.md): recognize the same change in a later run
- [Localization](LOCALIZATION.md): show the changes in your team's language, with `--lang`

### Reference
- [How oasdiff decides what is breaking](WHAT-IS-BREAKING.md): a change is judged against the contract, not against what one server accepts
- [Checks](CHECKS.md): every change oasdiff can detect, and its level
- [Nullability changes](NULLABILITY.md): when making a property nullable breaks clients
- [Schemas used in several places](SHARED-SCHEMAS.md): a change to a shared schema is reported once

## Known Limitations
- no checks for `context` instead of `schema` for request parameters
- no checks for `callback`s
