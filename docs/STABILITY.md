## API Stability Levels
When a new API is introduced, you may want to allow developers to change its behavior without triggering a breaking change error. Mark the endpoint with the `x-stability-level` extension:

```
/api/test:
  post:
    x-stability-level: "alpha"
```

There are four levels, in increasing order of stability:

- `draft` and `alpha` — can be changed freely; breaking changes are not reported
- `beta` and `stable` — breaking changes are reported as usual

Endpoints with no `x-stability-level` are treated like `stable`: any breaking change is reported.

### Allowed transitions
Stability may be increased (`draft` → `alpha` → `beta` → `stable`) but never decreased. An endpoint that previously had no `x-stability-level` may be assigned any level.

### See also
Stability levels also control [grace periods for API deprecation](DEPRECATION.md#grace-period).
