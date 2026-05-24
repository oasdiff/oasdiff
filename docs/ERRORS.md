# Errors

## Invalid Specs
Oasdiff expects valid OpenAPI 3 specs as input.  
The specs can be written in JSON or YAML.  
Oasdiff may return an error when given invalid specs, for example:
```
Error: failed to load base spec from "spec.yaml": error converting YAML to JSON: yaml: line 2: mapping values are not allowed in this context
```
The reason for this error is that the underlying library, [kin-openapi](https://github.com/getkin/kin-openapi), converts YAML specs to JSON before parsing them.

When a spec fails to load, oasdiff exits with code `102` (`101` for invalid flags, `103` when a glob matches no specs).

## Invalid specs vs. spec violations

`diff`, `breaking`, `changelog`, and `summary` fail only on specs they can't load or parse (the error above). They're otherwise lenient: a spec that loads but violates the OpenAPI or JSON Schema rules is still diffed.

To check a single spec for compliance (invalid types, missing required fields, bad regex, unresolved `$ref`s, and similar), use [`oasdiff validate`](VALIDATE.md). It reports each violation with a stable rule ID, a severity, and a source location, and exits 102 when the spec can't be loaded at all.