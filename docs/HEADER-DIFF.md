## Case-Insensitive Header Comparison

By default, oasdiff compares header names case-sensitively. This matches OpenAPI's general rule that "all field names in the specification are case sensitive".

On the wire, however, HTTP header names are case-insensitive (RFC 7230) — `Content-Type` and `content-type` refer to the same header. Set the `--case-insensitive-headers` flag to lowercase header names in each spec before diffing, so headers that differ only in case are treated as the same. The flag affects header parameters (`in: header`) and response headers.

```
oasdiff diff data/header-case/base.yaml data/header-case/revision.yaml --case-insensitive-headers
```
