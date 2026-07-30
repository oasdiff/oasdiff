# Version Bumps and Breaking Changes
If you version your API with [semantic versioning](https://semver.org), a breaking change is supposed to come with a major version bump.  
Oasdiff can tell you when it didn't, because it already knows both halves: the severity of the change, and `info.version` on each side.

## The checks
The policy is one rule: **a breaking change requires a major version increase**. There are three ways to break it, each with its own check id:

| Check id | Reported when a breaking change was detected and |
| ------------- | ------------- |
| `api-version-not-bumped` | `info.version` is unchanged |
| `api-version-decreased` | `info.version` moved backwards, for example `2.0.0` to `1.0.0` |
| `api-major-version-not-bumped` | `info.version` moved, but the major version did not increase, for example `1.0.0` to `1.1.0` |

For example:
```
info    [api-major-version-not-bumped]
        in info
                a breaking change was detected but the major version did not increase, from '1.0.0' to '1.1.0'
```

Their default level is `INFO`, so they never fail a build on their own.

## Enforcing the policy
To fail the build, raise the checks to `ERR` with [custom severity levels](BREAKING-CHANGES.md#customizing-severity-levels):
```
oasdiff breaking base.yaml revision.yaml --severity-levels oasdiff-levels.txt --fail-on ERR
```
Where `oasdiff-levels.txt` contains:
```
api-version-not-bumped          err
api-version-decreased           err
api-major-version-not-bumped    err
```

## Turning the policy off
If you don't want these checks at all, set them to `none` in the same file:
```
api-version-not-bumped          none
api-version-decreased           none
api-major-version-not-bumped    none
```

## When the checks stay quiet
Nothing is reported unless both of these hold:

1. **A breaking change was detected**, at the severity levels you configured. If you have downgraded a check to `INFO`, it is not breaking here either. A release with no breaking changes is never asked to bump its major version.
2. **Both versions are semantic versions.** `1.0`, `v1`, `2026-06-01` and other schemes have no "major version" to compare, so they are skipped rather than guessed at. A leading `v` is accepted (`v1.2.3`). Prerelease and build metadata are accepted and ignored, so `1.1.0-rc.1` is a minor bump.

Below `1.0.0`, semver gives the minor version the major's role, so `0.1.0` to `0.2.0` carries a breaking change and `0.1.0` to `0.1.1` does not.

## What is not checked
- Which version bump a non-breaking change deserves. Only breaking changes are checked.
- Any versioning scheme other than semver, including version numbers in the URL path.
