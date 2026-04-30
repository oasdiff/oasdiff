# Adding custom attributes to changelog entries basing on OpenAPI extension tags
Some people annotate their endpoints with OpenAPI Extension tags, for example:
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

Oasdiff can add these attributes to the changelog in JSON or YAML formats as follows:

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