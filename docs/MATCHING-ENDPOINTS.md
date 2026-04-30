# How oasdiff compares endpoints
Oasdiff compares matching endpoints to each other.  
By default, the matching algorithm **ignores** path parameter names.  
For example, the following endpoints will be compared to each other because they differ only by a path parameter name:
- GET /pet/{id}
- GET /pet/{petId}

This capability allows oasdiff to compare matching endpoints even if their path parameters were renamed.

# Duplicate Endpoints
Because oasdiff compares matching endpoints to each other, it expects a single instance of each endpoint to appear in each of the compared specs (or collections in [Composed Mode](COMPOSED.md))

In some cases, your specs may contain duplicate matching endpoints which will cause oasdiff to return an error, for example:
```
✗ oasdiff diff data/duplicate_endpoints/base.yaml data/duplicate_endpoints/revision.yaml
Error: diff failed with duplicate endpoint (GET /pet/{petId3}) found in data/duplicate_endpoints/base.yaml and data/duplicate_endpoints/base.yaml. You may add the x-since-date extension to specify order
```

There are two ways to overcome this:
1. If the duplication is a result of renaming path pararms, you can instruct oasdiff to include path parameter names in the endpoint matching algorithm with the `--include-path-params` flag
2. Use [`x-since-date`](#duplicate-endpoints-and-x-since-date)

# Duplicate Endpoints and `x-since-date`
If duplicate matching endpoints are found in either of the compared specs (or collections in [Composed Mode](COMPOSED.md)), then oasdiff uses the endpoint with the most recent `x-since-date` value.

- The `x-since-date` extension can be set at the Path or Operation level.
- `x-since-date` extensions set on the Operation level override the value set on Path level.
- If an endpoint doesn't have the `x-since-date` extension, its value is set to the default: "2000-01-01".
- Duplicate endpoints with the same `x-since-date` value will trigger an error.
- The format of the `x-since-date` is the RFC3339 full-date format.

Example of the `x-since-date` usage:
   ```
   /api/test:
    get:
     x-since-date: "2023-01-11"
   ```

