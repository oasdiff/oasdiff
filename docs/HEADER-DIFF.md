# Case-Insensitive Header Comparison

oasdiff compares header names case-insensitively by default, matching HTTP's wire rule that header names are case-insensitive (RFC 7230, RFC 9110). `Content-Type` and `content-type` refer to the same header and are treated as the same by the diff. The behaviour applies to header parameters (`in: header`) and response headers.

To restore case-sensitive comparison (matching OpenAPI's general rule that "all field names in the specification are case sensitive"), set `--case-insensitive-headers=false`:

```
oasdiff diff data/header-case/base.yaml data/header-case/revision.yaml --case-insensitive-headers=false
```
