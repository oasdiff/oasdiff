# Ignoring Specific Breaking Changes
Sometimes, you want to allow certain breaking changes, for example, when your spec and service are out-of-sync and you need to correct the spec.  
Oasdiff allows you define breaking changes that you want to ignore in a configuration file.  
You can specify the configuration file name in the oasdiff command-line with the `--warn-ignore` flag for WARNINGS or the `--err-ignore` flag for ERRORS.  
Each line in the configuration file should contain two parts:
1. Method and path (the first field in the line beginning with slash) to ignore a change to an endpoint, or the keyword 'components' to ignore a change in components
2. Description of the breaking change

For example:
```
GET /api/{domain}/{project}/badges/security-score removed the success response with the status '200'
```
Or, for a component change:
```
components removed the schema 'rules'
```

The required parts may appear in any order, in lower or upper case, and the configuration line may contain additional text, like this:
```
 - 12.01.2023 In GET /api/{domain}/{project}/badges/security-score we removed the success response with the status '200'
 - 31.10.2023 Removed the schema 'network-policies' from components
```

The configuration files can be of any text type, e.g., Markdown, so you can use them to document breaking changes and other important changes.
