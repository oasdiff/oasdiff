# Version Bumps and Breaking Changes
If you version your API with [semantic versioning](https://semver.org), a breaking change is supposed to come with a major version bump.
Oasdiff can fail the build when it didn't, because it already knows both halves: the severity of each change, and `info.version` on each side.

## Three ways to fail the run
`--fail-on` decides when a comparison fails the run, and the versioning policy is one of its values:

| `--fail-on` | Fails the run when |
| ------------- | ------------- |
| unset (the default) | never |
| `ERR`, `WARN`, `INFO` | a change reaches that severity |
| `MISSING-VERSION-BUMP` | a breaking change was found and `info.version` did not get a major version increase |

The values are alternatives: either the policy decides the exit code or a severity does, never both.

```
oasdiff breaking base.yaml revision.yaml --fail-on MISSING-VERSION-BUMP
```

`MISSING-VERSION-BUMP` is available on `breaking`, `changelog` and `breaking-files`. The rest of this page is about it.

The report lists every change at its own severity, whatever the verdict. The verdict is one line on standard error, with the exit code it carries:

| Verdict | Exit code | Line on standard error |
| ------------- | ------------- | ------------- |
| Satisfied | 0 | `versioning policy: satisfied. info.version moved from 1.0.0 to 2.0.0, which allows the breaking changes.` |
| Violated | 1 | `versioning policy: violated. info.version is still 1.0.0.` |
| Nothing to judge | 0 | `versioning policy: not applicable. No breaking changes, so no version increase is required.` |
| Cannot be judged | 1 | `versioning policy: cannot be evaluated. info.version must be a semantic version in both specs (base "1.0.0", revision "2026-06-01").` |

The line goes to standard error, never to standard output, so it does not corrupt a piped `--format json` or `--format yaml` report.

There are three ways to violate the policy, and the verdict says which one applies:

| Violated because | The verdict says |
| ------------- | ------------- |
| `info.version` is unchanged | `info.version is still 1.0.0` |
| `info.version` moved backwards | `info.version decreased from 2.0.0 to 1.0.0` |
| `info.version` moved, but the major version did not increase | `info.version moved from 1.0.0 to 1.1.0 without a major version increase` |

## The version change itself
Any change to `info.version` is reported in the changelog as `api-version-changed`, at `INFO`:

```
info    [api-version-changed]
        in info
                the API version changed from `1.0.0` to `2.0.0`
```

It is reported whatever the versioning scheme is, and whether or not anything breaking came with it, so a release's changelog says which version it is. Whether the change was the right one is the policy's question, and the policy's answer is the verdict above.

## What counts as a breaking change
Changes reported at `WARN` or `ERR` level are considered breaking and require a major version bump.

## Which versions can be compared
Both versions have to be semantic versions. `1.0`, `v1`, `2026-06-01` and other schemes have no major version to compare, so oasdiff reaches no verdict rather than guessing. A leading `v` is accepted (`v1.2.3`). Prerelease and build metadata are accepted and ignored, so `1.1.0-rc.1` is a minor bump.

With `--fail-on MISSING-VERSION-BUMP`, a comparison oasdiff cannot judge exits 1: you asked for the policy to be enforced, so passing the release silently would be worse than failing it.

Below `1.0.0`, semver gives the minor version the major's role, so `0.1.0` to `0.2.0` carries a breaking change and `0.1.0` to `0.1.1` does not.

## What is not checked
- Which version bump a non-breaking change deserves. Only breaking changes are judged.
- Any versioning scheme other than semver, including version numbers in the URL path.

## Upgrading from the version checks
`api-version-not-bumped`, `api-version-decreased` and `api-major-version-not-bumped` used to be checks: reported at `INFO` by default, and raised to `err` with `--severity-levels` to enforce the policy. They are no longer checks, because a verdict about the release should not be a change in the API. The verdict line says which of the three applies in words, and `api-version-changed` reports the version change itself.

A `--severity-levels` file that still sets a level for one of them fails to load, like any other unknown id, so remove those lines when you upgrade:

```
Error: failed to load custom severity levels from oasdiff-levels.txt: invalid rule id "api-version-not-bumped" on line 1
```
