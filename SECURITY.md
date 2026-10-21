# Security Policy

## Reporting a vulnerability

Report security issues privately through GitHub:
[Report a vulnerability](https://github.com/oasdiff/oasdiff/security/advisories/new).
This opens a private channel visible only to the maintainers.

Please do not open a public issue or pull request for a security problem, since
that discloses it before a fix is available.

Include what you have: the affected version, the commands or inputs that trigger
it, and what an attacker gains. A minimal reproduction is the most useful thing
you can send.

## What to expect

- An acknowledgement within a few days, saying whether the report reproduces.
- A fix developed in a private fork, released as a patch version.
- A published advisory crediting you by the name you prefer, unless you ask not
  to be credited.
- A CVE requested through GitHub when the issue warrants one.

We coordinate disclosure and will tell you before the advisory is published, so
you can publish your own write-up afterwards.

## Scope

This policy covers [oasdiff](https://github.com/oasdiff/oasdiff) and
[oasdiff-action](https://github.com/oasdiff/oasdiff-action).

For how to stop oasdiff following `$ref` values that point outside a spec, see
[External `$ref`s](docs/EXTERNAL-REFS.md).

oasdiff reads OpenAPI specifications, which are often untrusted: a spec may come
from a pull request opened by anyone. Reports about a spec causing oasdiff to
read files, make network requests, or consume unbounded resources are in scope
and are the ones we care most about.
