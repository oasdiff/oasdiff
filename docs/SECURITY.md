# Security

## SSRF via External $refs

OpenAPI specs can contain `$ref` values that point to external URLs or files.
By default, oasdiff follows these external refs to fully resolve the spec.

This can be a security concern when oasdiff processes **untrusted specs** — for example, when running in CI on a pull request authored by an external contributor.
A malicious spec could include a `$ref` like:

```yaml
$ref: 'http://169.254.169.254/latest/meta-data/iam/security-credentials/'
```

When oasdiff runs on this spec in a cloud-hosted CI environment, it would fetch that URL, potentially leaking cloud instance metadata and credentials.

## Mitigation

Use `--allow-external-refs=false` to prevent oasdiff from fetching external URLs or files referenced in specs:

```bash
oasdiff breaking base.yaml revision.yaml --allow-external-refs=false
```

This flag is supported by the `diff`, `breaking`, `changelog`, `summary`, and `flatten` commands.

When external refs are disabled, oasdiff will return an error if a spec contains an external `$ref`, rather than fetching it.

## Recommendation for CI

If your CI pipeline runs oasdiff on specs from untrusted sources (e.g., external pull requests), consider:

1. Using `--allow-external-refs=false` if your specs are self-contained
2. Running oasdiff in a sandboxed environment without access to cloud metadata endpoints or internal services
