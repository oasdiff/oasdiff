# Add OpenAPI Extensions to the Changelog
OpenAPI specs can carry custom `x-*` extension fields that attach metadata to operations. For example:
```
/restapi/oauth/token:
  post:
    operationId: getToken
    x-audience: Public
    summary: ...
    requestBody:
        ...
    responses:
        ...
```

Use the `--attributes` flag to include these values in JSON or YAML changelog entries:

```
❯ oasdiff changelog base.yaml revision.yaml -f yaml --attributes x-audience
- id: new-optional-request-property
  text: added the new optional request property ivr_pin
  level: 1
  operation: POST
  operationId: getToken
  path: /restapi/oauth/token
  source: new-revision.yaml
  section: paths
  attributes:
    x-audience: Public
```