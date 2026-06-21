## API Stability Levels
When a new API is introduced, you may want to allow developers to change its behavior without triggering a breaking change error.  
You can define an endpoint's stability level with the `x-stability-level` extension.  
There are four stability levels: `draft`->`alpha`->`beta`->`stable`.  

### Stability threshold (`--stability-level`)
By default, oasdiff uses a **beta** threshold: endpoints marked `draft` or `alpha` are excluded from breaking-change detection, while `beta` and `stable` endpoints are checked.

You can override the threshold with the `--stability-level` flag:

| Flag value | Endpoints checked |
|---|---|
| `draft` | all (`draft`, `alpha`, `beta`, `stable`) |
| `alpha` | `alpha`, `beta`, `stable` |
| `beta` *(default)* | `beta`, `stable` |
| `stable` | `stable` only |

Example:
```bash
# Include draft and alpha endpoints in breaking-change detection
oasdiff changelog base.yaml revision.yaml --stability-level draft
```

Endpoints with **no** `x-stability-level` are treated as `stable` and are always included regardless of the threshold.

Invalid values (e.g. `--stability-level=banana`) are rejected at flag-parse time.

### Bidirectional stability-level change detection
oasdiff detects changes to an endpoint's `x-stability-level` in **both** directions:

- **Decreased** (`stable`→`beta`, `beta`→`alpha`, etc.) — reported as `api-stability-decreased`
- **Increased** (`draft`→`alpha`, `alpha`→`beta`, etc.) — reported as `api-stability-increased`

The same detection applies to request and response properties:
- `request-property-stability-decreased` / `request-property-stability-increased`
- `response-property-stability-decreased` / `response-property-stability-increased`

These changes are only reported when the base stability (the level being left) meets the configured threshold. A destabilization from an in-scope level is reported, while a change to an element that was already below the threshold is not.

### Programmatic usage
Pass `WithStabilityLevel` as an option to `NewConfig`:
```go
config := checker.NewConfig(
    checker.GetAllChecks(),
    checker.WithOptionalChecks(optionalChecks),
    checker.WithStabilityLevel("alpha"),
)
```

### Example
   ```yaml
   /api/test:
    post:
     x-stability-level: "alpha"
   ```