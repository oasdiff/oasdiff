# How oasdiff Decides What Is Breaking
oasdiff judges a change against the API contract your OpenAPI definition declares, not against what a particular server happens to accept. A change is breaking if a consumer that followed the old contract can stop working under the new one.

This matters because an OpenAPI definition declares which requests and responses are valid, but most servers do not enforce it at runtime, and a server may quietly accept a request that the contract says is invalid. oasdiff still reports the change as breaking, because other consumers of the same contract do enforce it: API gateways and validators reject non-conforming requests, and generated client SDKs turn the contract into typed code that no longer compiles. Whether your own server is lenient is your choice to make; it does not mean the contract is unchanged.

Two consequences worth knowing:

- A change that makes a previously valid request invalid is breaking even if your server would still accept it. For example, adding a required request property is breaking: a request that omits it is invalid under the new contract, whether or not the property has a default. A default is a server-side fallback value; it does not make an omitted required property valid.
- `WARN` is reserved for changes where the definition genuinely does not contain enough information to decide, such as a change to an object parameter whose serialization the definition leaves unspecified. It is not used for changes that break some setups and not others; those are errors.

If a check's default severity does not match your API's compatibility policy, change it with [Customizing Severity Levels](SEVERITY-LEVELS.md).
