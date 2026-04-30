# Filtering Endpoints

You can filter endpoints in two ways:

## By path name
Use the `--match-path` option to exclude paths that don't match the given regular expression.  
For example, this diff includes only endpoints containing "/api" in the path:
```
oasdiff diff --match-path "/api" https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text
```

Use the `--unmatch-path` option to exclude paths that match the given regular expression.  
For example, this diff excludes endpoints containing "beta" in the path:
```
oasdiff diff --unmatch-path "beta" https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text
```

Note:  
If a path contains a callback, the filter will be applied both to the path itself and to the callback path.  
To include both the path and the callback, use a regular expression with a filter for each level, for example: "path|callback-path"
   
## By extension
Use the `--filter-extension` option to exclude paths and operations with an OpenAPI Extension matching the given regular expression.  
For example, this diff excludes paths and operations with extension "x-beta":
```
oasdiff diff --filter-extension "x-beta" https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml -f text
```
Notes:
1. OpenAPI Extensions can be defined both at the path level and at the operation level. Both are matched and excluded with this flag.
2. If a path or operation has a matching extension only in one of the specs, but not in the other, it will appear as Added or Deleted.