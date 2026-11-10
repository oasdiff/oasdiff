# Version Bumps and Breaking Changes
If you version your API with [semantic versioning](https://semver.org), a breaking change is supposed to come with a major version bump.
Oasdiff can fail the build when it didn't, because it already knows both halves: the severity of each change, and `info.version` on each side.

## The flag
`--enforce-version-bump` changes what the exit code means. Instead of "are there changes at this level", it answers "were the breaking changes carried by a major version increase":

```
oasdiff breaking base.yaml revision.yaml --enforce-version-bump
```

It is available on `breaking`, `changelog` and `breaking-files`, and is off by default.

The report itself is unchanged: every change is still listed, at its own severity. What the flag adds is the exit code and one line on standard error saying how the version was judged:

| Verdict | Exit code | Line on standard error |
| ------------- | ------------- | ------------- |
| Satisfied | 0 | `versioning policy: satisfied. info.version moved from 1.0.0 to 2.0.0, which allows the 3 breaking changes.` |
| Violated | 1 | `versioning policy: violated (api-version-not-bumped). 3 breaking changes, and info.version is still 1.0.0.` |
| Nothing to judge | follows `--fail-on` | `versioning policy: not applicable. No breaking changes, so no version increase is required.` |
| Cannot be judged | 1 | `versioning policy: cannot be evaluated. info.version must be a semantic version in both specs (base "1.0.0", revision "2026-06-01").` |

The line goes to standard error, never to standard output, so it does not corrupt a piped `--format json` or `--format yaml` report.

There are three ways to violate the policy, and the verdict names the one that applies:

| Reason | Violated because |
| ------------- | ------------- |
| `api-version-not-bumped` | `info.version` is unchanged |
| `api-version-decreased` | `info.version` moved backwards, for example `2.0.0` to `1.0.0` |
| `api-major-version-not-bumped` | `info.version` moved, but the major version did not increase, for example `1.0.0` to `1.1.0` |

## What it does to `--fail-on`
A major version increase licenses the breaking changes, so with the policy satisfied, errors no longer fail the run on their own. Anything below error still answers to `--fail-on`: with `--enforce-version-bump --fail-on WARN`, a warning fails the run whatever the version did, because oasdiff is not sure a warning is breaking and a major bump is not a declaration about it.

When there are no breaking changes at all, the policy has nothing to judge and `--fail-on` behaves as it does without the flag.

## When the policy has nothing to judge
A release is asked for a major version increase only when a breaking change was found, at the severity levels you configured. If you have downgraded a check below error with [custom severity levels](BREAKING-CHANGES.md#customizing-severity-levels), or ignored it with an ignore file, it is not breaking here either, and the verdict line says so.

## Which versions can be compared
Both versions have to be semantic versions. `1.0`, `v1`, `2026-06-01` and other schemes have no major version to compare, so oasdiff reaches no verdict rather than guessing. A leading `v` is accepted (`v1.2.3`). Prerelease and build metadata are accepted and ignored, so `1.1.0-rc.1` is a minor bump.

With `--enforce-version-bump`, a comparison oasdiff cannot judge exits 1: you asked for the policy to be enforced, so passing the release silently would be worse than failing it.

Below `1.0.0`, semver gives the minor version the major's role, so `0.1.0` to `0.2.0` carries a breaking change and `0.1.0` to `0.1.1` does not.

## What is not checked
- Which version bump a non-breaking change deserves. Only breaking changes are judged.
- Any versioning scheme other than semver, including version numbers in the URL path.

## Upgrading from the version checks
`api-version-not-bumped`, `api-version-decreased` and `api-major-version-not-bumped` used to be checks: reported at `INFO` by default, and raised to `err` with `--severity-levels` to enforce the policy. They are no longer checks, because the verdict is about the release rather than about the API. It named no endpoint, it was counted in the change totals, and raising it to an error turned it into a finding a reviewer had to approve like any other. The three names survive as the reasons printed with a violation.

A `--severity-levels` file that still sets a level for one of them keeps working, so an upgrade does not fail on it. The level has no effect, and oasdiff says so:

```
warning: api-version-not-bumped is no longer a check, so the level set in oasdiff-levels.txt has no effect. Use --enforce-version-bump to require a major version increase for breaking changes.
```
