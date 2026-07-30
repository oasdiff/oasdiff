# Version Bumps and Breaking Changes
If you version your API with [semantic versioning](https://semver.org), a breaking change is supposed to come with a major version bump.  
Oasdiff can tell you when it didn't, because it already knows both halves: the severity of the change, and `info.version` on each side.

## The check
The check id is `api-major-version-not-bumped`.  
It reports a comparison where a breaking change was detected but `info.version` moved by less than a major version:
```
info    [api-major-version-not-bumped]
        in info
                a breaking change was detected but the major version did not increase, from '1.0.0' to '1.1.0'
```

Its default level is `INFO`, so by itself it never fails a build. To enforce it, raise it to `ERR` with [custom severity levels](BREAKING-CHANGES.md#customizing-severity-levels):
```
oasdiff breaking base.yaml revision.yaml --severity-levels oasdiff-levels.txt --fail-on ERR
```
Where `oasdiff-levels.txt` contains:
```
api-major-version-not-bumped    err
```

## When it stays quiet
The check reports nothing unless all of these hold. This is deliberate: most specs don't use `info.version` as a release signal, and a check that fires on all of them would be noise.

1. **`info.version` changed.** An unchanged version is not evidence of a missed bump. A spec whose version has been `1.0.0` since the day it was written is not tracking releases in `info.version` at all, and oasdiff can't tell that case apart from a genuine oversight.
2. **Both versions are semantic versions.** `1.0`, `v1`, `2026-06-01` and other schemes have no "major version" to compare, so they are skipped. A leading `v` is accepted (`v1.2.3`). Prerelease and build metadata are accepted and ignored: `1.1.0-rc.1` is a minor bump.
3. **A breaking change was detected**, at the severity levels you configured. If you have downgraded a check to `INFO`, it is not breaking here either.

Below `1.0.0`, semver gives the minor version the major's role, so `0.1.0` to `0.2.0` carries a breaking change and `0.1.0` to `0.1.1` does not.

A version that moves backwards, for example `2.0.0` to `1.0.0`, is not reported. Nothing was bumped, so "the bump was too small" would be the wrong thing to say about it.

## What it does not check
- That every change bumps the version. Only breaking changes are checked, for the reason in point 1 above.
- Which version bump a non-breaking change deserves.
- Any versioning scheme other than semver, including version numbers in the URL path.
