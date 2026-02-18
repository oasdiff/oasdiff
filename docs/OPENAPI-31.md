# OpenAPI 3.1 Support (Beta)

oasdiff now supports OpenAPI 3.1 specs across all commands: `diff`, `breaking`, and `changelog`.

This feature is currently in **beta** and available for community testing before general availability.

## What's supported

### Spec loading
- OpenAPI 3.1.0 document parsing
- JSON Schema 2020-12 keywords
- Webhooks

### Diff
Changes are detected for all 3.1-specific fields:
- `exclusiveMinimum`/`exclusiveMaximum` as numeric values
- `const`, `prefixItems`, `contains`, `minContains`, `maxContains`
- `if`/`then`/`else` conditional schemas
- `dependentRequired`, `dependentSchemas`
- `patternProperties`, `propertyNames`
- `unevaluatedItems`, `unevaluatedProperties`
- `contentSchema`, `contentMediaType`, `contentEncoding`
- Webhooks (added/deleted/modified)
- Info `summary`, License `identifier`

### Breaking changes and changelog
148 new rule IDs covering:
- **Nullable type arrays**: `type: ["string", "null"]` correctly detected as nullable changes (not false-positive type changes)
- **Webhooks**: added/removed detection + all existing operation-level checks automatically applied to modified webhook operations
- **const**: added/removed/changed for request and response body/properties
- **exclusiveMinimum/exclusiveMaximum**: increased/decreased/set for request body/properties/parameters and response body/properties
- **prefixItems**: added/removed for request and response
- **if/then/else**: added/removed for request and response
- **contains/minContains/maxContains**: added/removed/increased/decreased
- **dependentRequired**: added/removed/changed
- **dependentSchemas**: added/removed
- **patternProperties**: added/removed
- **propertyNames**: added/removed
- **unevaluatedItems/unevaluatedProperties**: added/removed
- **contentSchema/contentMediaType/contentEncoding**: added/removed/changed

## How to try it

### Docker
```bash
docker pull tufin/oasdiff:feat-openapi-3.1-support

# Diff
docker run --rm -v $(pwd):/specs tufin/oasdiff:feat-openapi-3.1-support diff /specs/base.yaml /specs/revision.yaml

# Breaking changes
docker run --rm -v $(pwd):/specs tufin/oasdiff:feat-openapi-3.1-support breaking /specs/base.yaml /specs/revision.yaml

# Changelog
docker run --rm -v $(pwd):/specs tufin/oasdiff:feat-openapi-3.1-support changelog /specs/base.yaml /specs/revision.yaml
```

### Build from source
```bash
git clone -b feat/openapi-3.1-support https://github.com/oasdiff/oasdiff.git
cd oasdiff
go build
./oasdiff diff base.yaml revision.yaml
```

## Provide feedback

We need your feedback to move this to general availability.

### Found an issue?
[Open an issue](https://github.com/oasdiff/oasdiff/issues/new?template=bug_report.md&title=[3.1]%20) with `[3.1]` in the title.

### It works for you?
We need to hear from you too! Please add a thumbs-up reaction to [this tracking issue](https://github.com/oasdiff/oasdiff/issues/52) and optionally leave a comment describing your use case and which 3.1 features you tested. Knowing how many users are running 3.1 successfully helps us decide when to make it generally available.

## Status

| Milestone | Status |
|-----------|--------|
| Spec loading | Done |
| Diff | Done |
| Breaking changes | Done |
| Changelog | Done |
| Community testing | **In progress** |
| General availability | Pending testing feedback |
