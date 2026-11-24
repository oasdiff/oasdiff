# Case-Insensitive Header Comparison

HTTP header names are case-insensitive (RFC 7230, RFC 9110): `Content-Type` and `content-type` are the same header. This applies to header parameters (`in: header`), response headers, and the header an `apiKey` security scheme is sent in (`in: header`). Query parameter and cookie names are case-sensitive.

## breaking and changelog
`oasdiff breaking` and `oasdiff changelog` report what changed for clients, so they compare header names ignoring case. A header renamed only in case is not a change:

```
oasdiff changelog data/header-case/base.yaml data/header-case/revision.yaml
```

A header that was renamed in case and also changed is reported as one changed header, under its name in the base spec. Header names are printed as they are written in the spec.

`--case-insensitive-headers` is deprecated in these commands: it is accepted and ignored.

## diff and summary
`oasdiff diff` and `oasdiff summary` report what changed in the document, and OpenAPI field names are case-sensitive. A header renamed only in case is a deleted and an added header:

```
oasdiff diff data/header-case/base.yaml data/header-case/revision.yaml
```

To ignore case, pass `--case-insensitive-headers`. It lowercases the header names in both specs before comparing them:

```
oasdiff diff data/header-case/base.yaml data/header-case/revision.yaml --case-insensitive-headers
```
