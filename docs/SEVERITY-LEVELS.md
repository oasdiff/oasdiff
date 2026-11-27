# Customizing Severity Levels
Oasdiff allows you to change the default severity levels according to your needs.  
For example, the default severity level of the `api-security-scope-removed` check is `INFO`. You can verify this by running `oasdiff checks changelog`.  
To change the `api-security-scope-removed` check's severity level to `ERR` use the following command:
```
oasdiff changelog data/checker/api_security_updated_base.yaml data/checker/api_security_updated_revision.yaml --severity-levels oasdiff-levels.txt
```
Where the file `oasdiff-levels.txt` contains a single line:
```
api-security-scope-removed    err
```

Checks can be customized with the following levels:
| Custom Level  | Check Status |
| ------------- | ------------- |
| err  | Enabled with level ERR |
| warn  | Enabled with level WARN |
| info  | Enabled with level INFO |
| none  | Disabled  |

## When oasdiff reports a change at another level than its check's
The level a check declares, which is what `oasdiff checks changelog` lists, is what the change means in general. For a specific change, oasdiff may report another level, with a comment explaining why. This happens for three reasons:

- **The comparison was not exact.** A change inside an `allOf` is the common case: without [`--flatten-allof`](ALLOF.md) the branches are compared one at a time, so another branch may already describe what changed in one branch. Such a change is reported at `WARN` even where its check declares `ERR`, and the comment says why and what to run to get the exact answer.
- **The spec shows the change cannot break anyone.** A property marked `readOnly` never appears in requests, so restricting it (adding a `pattern`, setting a `maximum`) cannot invalidate a request; the same holds for `writeOnly` properties in responses. Such a change is reported at `INFO` even where its check declares `ERR`, and the comment states the reason. The conditions a check recognizes are listed in its GUARDS column in `oasdiff checks changelog` (see [CHECKS.md](CHECKS.md#categorization)).
- **The change is under `not`.** A value is valid against `not` when it is invalid against the schema inside it, so a change to that schema has the opposite effect: raising a `minLength` under `not` accepts more values, and lowering it accepts fewer. Such a change is reported at the level of the opposite effect, which can be higher than the check's level as well as lower: `ERR` where its check declares `INFO`, and `INFO` where it declares `ERR`. Under two `not`s the effect is reversed twice, and the change keeps its check's level.

**Setting a level yourself turns this off for that check.** The level you set is the level you get, whether or not one of the reasons above applies. This matters in both directions: pinning a check to `err` means it will report `ERR` even where oasdiff would otherwise have lowered it, and pinning one to the level it already has is enough to opt out.
