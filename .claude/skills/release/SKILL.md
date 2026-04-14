---
name: release
description: Tag and create a GitHub release, only after verifying CI passes
allowed-tools: Bash, Read, Glob, Grep
---

## Release workflow

Before tagging and releasing, verify that CI has passed:

1. Find the latest CI run for the current branch:
   ```
   gh run list --branch $(git branch --show-current) --limit 5
   ```
2. Check that all required workflow runs have completed successfully (go, lint, govulncheck, etc.)
3. If any run is still in progress, wait for it to complete before proceeding
4. If any run has failed, stop and report the failure — do NOT tag or release

Only after all CI checks pass:

5. Determine the version tag (from user input or by incrementing the latest tag)
6. Tag the commit: `git tag <version>`
7. Push the tag: `git push origin <version>`
8. Create the GitHub release: `gh release create <version> ...`
   - Use `--prerelease` for beta/rc tags
   - Include a summary of changes since the previous release
