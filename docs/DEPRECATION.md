# Deprecating APIs, Parameters and Properties
Sometimes [resources](#supported-resources-for-deprecation) need to be removed from the API, for example, when we replace an API endpoint by a new version, or when a parameter is no longer supported.  
As API owners, we want a process that will allow us to phase out the old resource smoothly as possible and with minimal disruptions to business.

OpenAPI specification supports a ```deprecated``` flag which can be used to mark resources as deprecated.  
Deprecating a resource **isn't** considered a breaking change since it doesn't break the client but only serves as an indication of an intent to remove something in the future.  
After deprecating a resource, it can be removed without triggering a breaking change since the client already knows it is going to be removed.

## Deprecation without a sunset date
Oasdiff allows you to gracefully remove a resource without getting a breaking change error, as follows:
1. The resource is marked as ```deprecated```:
   ```
   /api/test:
    get:
     deprecated: true
   ```
2. Subsequently, the resource can be removed without triggering a breaking change error.

## Deprecation with a sunset date
A more mature deprecation process includes a sunset date which tells the API clients how long they can still use the deprecated resource before sunset:
1. The resource is marked as ```deprecated``` and a [special extension](https://swagger.io/specification/#specification-extensions) ```x-sunset``` is added to announce the date at which the resource will be removed:
   ```
   /api/test:
    get:
     deprecated: true
     x-sunset: "2022-12-31"
   ```
   Note sunset date should conform with [RFC3339](https://datatracker.ietf.org/doc/html/rfc3339).    
2. At the sunset date or anytime later, the resource can be removed without triggering a breaking change error. An earlier removal will be considered a breaking change.

Notes:
1. In this mode, oasdiff considers the `x-sunset` extension as **optional** allowing developers to deprecate resources with or without the `x-sunset` extension.  
2. After an `x-sunset` extension is specified, changing it to an earlier date will trigger a breaking change.  

## Enforcing Sunset Grace Period
Oasdiff allows you to enforce a sunset grace period: the minimal number of days required between deprecating a resource and removing it.  
In this mode oasdiff considers the `x-sunset` extension **mandatory**, so that deprecating a resource must be acompanied with an `x-sunset` extension, otherwise it is considered breaking.  

For example, the following command requires deprecation of resources to be accompanied by an ```x-sunset``` extension with a date which is at least 180 days away, otherwise the deprecation itself will be considered a breaking change:
```
oasdiff breaking data/deprecation/base.yaml data/deprecation/deprecated-past.yaml --deprecation-days-stable=180
```

If you are using [API stability levels](STABILITY.md) you can define different sunset grace periods for:
- Stable APIs: with the `--deprecation-days-stable`
- Beta APIs: with the `--deprecation-days-beta`

Notes:
1. Deprecation days can be set to non-negative integers
2. Setting deprecation days to a zero value disables enforcement and reverts to the [Deprecation with a sunset date](#deprecation-with-a-sunset-date) behavior
3. After an `x-sunset` extension is specified, it can only be changed to a future date which respects the sunset grace period relative to date of the change.
4. Changing an `x-sunset` date which doesn't conform with RFC3339 to one that does is not reported, since clients had no date to rely on. Deleting the date, or changing it to another which doesn't conform, is reported even when `x-stability-level` doesn't resolve to a grace period, since only the grace period comparison needs one. This is the same for endpoints, parameters and properties.

## Supported Resources for Deprecation
OpenAPI 3 supports the `deprecated` flag on `Operations`, `Parameters`, `Headers` and `Schemas`. Oasdiff reads it on three of the four:

- `Operations`: as an endpoint. Removing a whole path is reported once per endpoint under it, each according to its own flag.
- `Parameters`: as a parameter.
- `Schemas`: as a request property or a required response property.
- `Headers`: not supported. Removing a response header is reported as `required-response-header-removed` or `optional-response-header-removed` whether or not it was deprecated.

Removing one of them is reported as:

| removed resource | not deprecated | deprecated, no sunset date | sunset date in the future |
|---|---|---|---|
| path, with the endpoints under it | `api-path-removed-without-deprecation` (error) | `api-path-removed-with-deprecation` (info) | `api-path-removed-before-sunset` (error) |
| endpoint | `api-removed-without-deprecation` (error) | `api-removed-with-deprecation` (info) | `api-removed-before-sunset` (error) |
| parameter | `request-parameter-removed` (warning) | `request-parameter-removed-with-deprecation` (info) | `request-parameter-removed-before-sunset` (error) |
| request property | `request-property-removed` (warning) | `request-property-removed-with-deprecation` (info) | `request-property-removed-before-sunset` (error) |
| required response property | `response-required-property-removed` (error) | `response-required-property-removed-with-deprecation` (info) | `response-required-property-removed-before-sunset` (error) |

At the sunset date or later, the removal isn't reported at all.

Notes:
1. An `x-sunset` date which doesn't conform with RFC3339 is reported as `api-path-sunset-parse` for paths and endpoints, and as `request-parameter-sunset-parse` for parameters. For properties it is reported with the plain removal id, since the deprecation checks already report the invalid date.
2. Removing an optional response property is reported as `response-optional-property-removed` whether or not it was deprecated, since that removal is not breaking either way.
