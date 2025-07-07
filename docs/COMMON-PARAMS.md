## Common Parameters

### Common Parameters Definition
Parameters shared by all operations of a path can be defined on the path level instead of the operation level.  
Path level parameters are inherited by all operations of that path.  
A typical use case are the GET/PUT/PATCH/DELETE operations that manipulate a resource accessed via a path parameter.

### Diff and Common Parameters
The oasdiff `diff` sub-command reports changes to parameters at both levels: the path level and the operation level.  
Note, however, that each level is checked seperately.  
For example, if a parameter is moved from the path level to the operation level, it will be reported as a deletion and an addition:

params_in_path.yaml:
```
paths:
  '/admin/v0/abc/{id}':
    parameters:
      - $ref: '#/components/parameters/tenant_id'
      - $ref: '#/components/parameters/id'
    get:
	  ...
```

params_in_op.yaml:
```
paths:
  '/admin/v0/abc/{id}':
    get:
      parameters:
        - $ref: '#/components/parameters/tenant_id'
        - $ref: '#/components/parameters/id'
    ...
```

Command-line:
```
oasdiff diff data/common-params/params_in_path.yaml data/common-params/params_in_op.yaml --exclude-elements endpoints
```
Output: 
```
paths:
    modified:
        /admin/v0/abc/{id}:
            operations:
                modified:
                    GET:
                        parameters:
                            added:
                                header:
                                    - tenant-id
                                path:
                                    - id
            parameters:
                deleted:
                    header:
                        - tenant-id
                    path:
                        - id
```						

To overcome this limitation use the `--flatten-params` flag which merges common parameters from the path level into the operation level before running the diff:
```
oasdiff diff data/common-params/params_in_path.yaml data/common-params/params_in_op.yaml --flatten-params
 ```
The output will be empty meaning that no change was found.

### Changelog (incl. Breaking Changes) and Common Parameters
The `changelog` and `breaking` sub-commands are focused on operation level parameters and ignore most parameter changes at the path level.  
For example, changing `maxItems` of a path level parameter, won't normally be reported:

```
paths:
  /api/v1.0/groups:
    parameters:
      - in: query
        name: category
        schema:
          type: array
          items:
            type: string
            maxItems: 20
```

```
paths:
  /api/v1.0/groups:
    parameters:
      - in: query
        name: category
        schema:
          type: array
          items:
            type: string
            maxItems: 10
```

But if you add the `--flatten-params` flag which merges common parameters from the path level into the operation level before running the diff, it will be reported as a breaking change:

```
oasdiff changelog data/checker/common_request_parameter_max_items_updated_revision.yaml data/checker/common_request_parameter_max_items_updated_base.yaml --flatten-params
```

Outout:
```
1 changes: 1 error, 0 warning, 0 info
error	[request-parameter-max-items-decreased] at data/common-params/request_parameter_max_items_updated_base.yaml
	in API POST /api/v1.0/groups
		for the 'query' request parameter 'category', the maxItems was decreased from '20' to '10'
```

### Summary
It is recommended to use the `--flatten-params` flag to increase accuracy for common parameters.
