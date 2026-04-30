# Errors

## Invalid Specs
Oasdiff expects valid OpenAPI 3 specs as input.  
The specs can be written in JSON or YAML.  
Oasdiff may return an error when given invalid specs, for example:
```
Error: failed to load base spec from "spec.yaml": error converting YAML to JSON: yaml: line 2: mapping values are not allowed in this context
```
The reason for this error is that the underlying library, [kin-openapi](https://github.com/getkin/kin-openapi), converts YAML specs to JSON before parsing them.