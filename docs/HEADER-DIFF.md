# Case-Insensitive Header Comparison

HTTP header names are case-insensitive (RFC 7230, RFC 9110): `Content-Type` and `content-type` are the same header. `--case-insensitive-headers` compares them as the same. It applies to header parameters (`in: header`), response headers, and the header an `apiKey` security scheme is sent in (`in: header`). Query parameter and cookie names stay case-sensitive.

The default depends on what the command reports:

| Command | Default | Why |
|---|---|---|
| `breaking`, `changelog`, `breaking-files`, the git diff driver | `true` | They report what changed for clients, and a client sees no difference between `X-Trace` and `x-trace`. |
| `diff`, `summary` | `false` | They report what changed in the document, and OpenAPI field names are case-sensitive. A header renamed only in case is a deleted and an added header. |

To report a rename in case from the checks:

```
oasdiff changelog data/header-case/base.yaml data/header-case/revision.yaml --case-insensitive-headers=false
```

To ignore it in `diff`:

```
oasdiff diff data/header-case/base.yaml data/header-case/revision.yaml --case-insensitive-headers
```
