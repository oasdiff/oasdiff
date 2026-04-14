## Merging AllOf Schemas [Beta]
OpenAPI 3.0 provides several keywords which can be used to combine schemas.  
You can use these keywords to create a complex schema or validate a value against multiple criteria.  
- oneOf – validates the value against exactly one of the subschemas
- allOf – validates the value against all the subschemas
- anyOf – validates the value against any (one or more) of the subschemas

Using these keywords can be useful to describe complex data models but it complicates breaking changes detection.
Consider, for example, the following comparison of two OpenAPI specs:
```
oasdiff breaking data/allof/simple.yaml data/allof/revision.yaml 
```

The result shows one breaking change which is due to a new subschema that was added under allOf. But the new subschema, doesn't actually add any new constraints, because it is identical to a previously existing subschema, and, as such, this isn't a breaking chanage.
You can verify this with a regular diff comparison:
```
diff --side-by-side data/allof/simple.yaml data/allof/revision.yaml
```

In order to reduce such false-positives, oasdiff supports the ability to replace allOf by a merged equivalent before comparing the specs, like this:

```
oasdiff breaking data/allof/simple.yaml data/allof/revision.yaml --flatten-allof
```
In this case no breaking changes are reported, correctly.  
The `--flatten-allof` flag is also supported with `diff`, `changelog` and `summary`.

In order to see how oasdiff merges allOf, you can use the dedicated `flatten` command:
```
oasdiff flatten data/allof/simple.yaml
```

### Source location limitations
When using `--flatten-allof`, source locations (file, line, column) are not available for changes detected in flattened schemas. This is because the merged schema is a new construct that doesn't correspond to a single location in the original file.

The following schema fields are not merged:
- Extensions
- Example
- ExternalDocs
- AllowEmptyValue
- Deprecated
- XML
- Discriminator

Please help us improve this feature by providing feedback and reporting issues.
