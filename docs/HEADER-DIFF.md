## Case-Insensitive Header Comparison

HTTP header names are case-insensitive — `Content-Type` and `content-type` refer to the same header. By default, oasdiff compares header names as written, so two specs declaring the same header in different cases will appear to differ.

The `--case-insensitive-headers` flag normalizes all header names to lowercase in both specs before diffing. This covers both request header parameters (`in: header`) and response headers.

```
oasdiff diff data/header-case/base.yaml data/header-case/revision.yaml --case-insensitive-headers
```
