# Notes for Go Developers

## Embedding oasdiff into your program
Load each spec with the kin-openapi loader, then call `diff.Get`:

```go
import (
    "github.com/getkin/kin-openapi/openapi3"
    "github.com/oasdiff/oasdiff/diff"
)

loader := openapi3.NewLoader()

s1, err := loader.LoadFromFile("base.yaml")
// handle err

s2, err := loader.LoadFromFile("revision.yaml")
// handle err

diffReport, err := diff.Get(diff.NewConfig(), s1, s2)
// handle err
```

Use `diff.NewConfig()` rather than `&diff.Config{}` — the constructor initializes internal maps that the diff engine expects to be non-nil.

## Runnable Examples
- [diff](https://pkg.go.dev/github.com/oasdiff/oasdiff/diff#example-Get)
- [breaking changes](https://pkg.go.dev/github.com/oasdiff/oasdiff/diff#example-GetPathsDiff)

## OpenAPI References
oasdiff expects [OpenAPI references](https://swagger.io/docs/specification/using-ref/) to be resolved. The kin-openapi loader resolves them automatically when you load the spec; if you build a spec another way, resolve them with [Loader.ResolveRefsIn](https://pkg.go.dev/github.com/getkin/kin-openapi/openapi3#Loader.ResolveRefsIn).
