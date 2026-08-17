# Severity-Law Triage

Working document for reviewing the 56 rules whose stored level disagrees with
the level derived by the severity law (see `rule_severity_law_test.go`, run
with `go test ./checker -run SeverityLawReport -v`). 500 of 556 rules derive
exactly; every mismatch below has a root cause and a proposed handling.

Fill in the Decision column: **keep** (stored level stands, entry moves to the
deviation ledger with the stated reason), **fix** (stored level changes to the
derived one, files as an issue), or an alternative.

The law: Narrows x request = ERR, Narrows x response = INFO, Widens x request
= INFO, Widens x response = ERR, Incomparable = ERR, Unknown = WARN, None
(metadata) = INFO, lifecycle violation = ERR. Guards: readOnly nullifies
request-side effect, writeOnly nullifies response-side effect, sanctioned
removal (deprecated with honored sunset) = INFO.

---

## Bucket A: apparent conventions (35 rules)

Stored levels that deviate from the law deliberately and consistently. If
kept, each group becomes one row in the deviation ledger.

### A1. Bound-set reported as WARN (24 rules)

Setting a bound on a request narrows the contract (derived ERR), but every
`*-set` rule is WARN with the standard comment ("the restriction is sometimes
legitimately required..."). This is the documented convention for
narrowing-from-unbounded on numeric bounds.

| Rule | Stored | Derived | Decision |
|---|---|---|---|
| request-body-exclusive-max-set | WARN | ERR | |
| request-body-exclusive-min-set | WARN | ERR | |
| request-body-max-items-set | WARN | ERR | |
| request-body-max-length-set | WARN | ERR | |
| request-body-max-properties-set | WARN | ERR | |
| request-body-max-set | WARN | ERR | |
| request-body-min-items-set | WARN | ERR | |
| request-body-min-set | WARN | ERR | |
| request-body-multiple-of-set | WARN | ERR | |
| request-parameter-exclusive-max-set | WARN | ERR | |
| request-parameter-exclusive-min-set | WARN | ERR | |
| request-parameter-max-length-set | WARN | ERR | |
| request-parameter-max-set | WARN | ERR | |
| request-parameter-min-items-set | WARN | ERR | |
| request-parameter-min-set | WARN | ERR | |
| request-property-exclusive-max-set | WARN | ERR | |
| request-property-exclusive-min-set | WARN | ERR | |
| request-property-max-items-set | WARN | ERR | |
| request-property-max-length-set | WARN | ERR | |
| request-property-max-properties-set | WARN | ERR | |
| request-property-max-set | WARN | ERR | |
| request-property-min-items-set | WARN | ERR | |
| request-property-min-set | WARN | ERR | |
| request-property-multiple-of-set | WARN | ERR | |

Note the tension: `request-body-became-enum` and `request-body-const-added`
are the same narrowing-from-unbounded shape and are ERR. If A1 is kept, the
ledger reason should say why bounds differ from enums/const; if fixed, 24
levels change at once.

### A2. Removal that widens the contract but signals behavior change (4 rules)

Removing a parameter, property, or body definition widens or leaves unchanged
the accepted set (unknown fields are allowed by default), so the law derives
INFO. Stored levels treat removal as a signal that the server stopped
processing the element.

| Rule | Stored | Derived | Decision |
|---|---|---|---|
| request-body-removed | ERR | INFO | |
| request-parameter-removed | WARN | INFO | |
| request-property-removed | WARN | INFO | |
| response-optional-property-removed | WARN | INFO | |

Note the internal inconsistency: request-body-removed is ERR while the
parameter/property analogs are WARN.

### A3. Single-to-oneOf wrapping kept breaking (2 rules)

Wrapping in oneOf that includes the original branch widens; #1037 deliberately
kept the breaking verdict. Law derives WARN (unknown, since the check does not
verify the original branch survives intact).

| Rule | Stored | Derived | Decision |
|---|---|---|---|
| request-body-wrapped-in-one-of | ERR | WARN | |
| response-body-wrapped-in-one-of | ERR | WARN | |

### A4. Miscellaneous single-rule conventions (5 rules)

| Rule | Stored | Derived | Root cause | Decision |
|---|---|---|---|---|
| response-non-success-status-removed | INFO | ERR | removing an error status is treated as sanctioned cleanup, though clients handling that status break | |
| optional-response-header-removed | WARN | ERR | optional softens the variant-removal verdict; clients "should not" rely on it | |
| response-required-property-became-not-write-only | WARN | INFO | readOnly/writeOnly are advisory (SHOULD) in the spec, so the law says INFO; stored WARN flags that the field now appears in responses | |
| request-body-all-of-removed | WARN | INFO | removing an allOf conjunct widens (safe), stored WARN as caution | |
| request-property-all-of-removed | WARN | INFO | same | |

---

## Bucket B: candidate severity bugs (21 rules)

Stored levels that contradict the contract rule with no visible convention
behind them.

### B1. Security family: unproven "safe" verdicts (4 rules)

Security requirements are OR-alternatives and scopes are AND-ed within one
requirement. Removing an alternative breaks clients authenticating with it;
adding a scope breaks clients lacking it. All stored INFO.

| Rule | Stored | Derived | Decision |
|---|---|---|---|
| api-security-removed | INFO | ERR | |
| api-global-security-removed | INFO | ERR | |
| api-security-scope-added | INFO | ERR | |
| api-global-security-scope-added | INFO | ERR | |

### B2. anyOf-added vs oneOf-added contradiction (2 rules)

`response-body-one-of-added` is ERR; the identical widening via anyOf is INFO.

| Rule | Stored | Derived | Decision |
|---|---|---|---|
| response-body-any-of-added | INFO | ERR | |
| response-property-any-of-added | INFO | ERR | |

### B3. Response pattern rules (2 rules, confirms #1034)

The law independently re-derived the tracked gap: loosening a response
pattern weakens the output guarantee.

| Rule | Stored | Derived | Decision |
|---|---|---|---|
| response-property-pattern-removed | INFO | ERR | |
| response-property-pattern-changed | INFO | WARN | |

### B4. prefixItems verdicts assume a direction that does not exist (8 rules)

prefixItems reshapes positional constraints; in general neither the old nor
the new accepted set contains the other, so the law derives WARN (unknown).
Stored levels treat added-as-widening / removed-as-narrowing, making
`request-*-prefix-items-added` an unproven "safe" (INFO) and the ERR entries
unproven "breaking".

| Rule | Stored | Derived | Decision |
|---|---|---|---|
| request-body-prefix-items-added | INFO | WARN | |
| request-body-prefix-items-removed | ERR | WARN | |
| request-property-prefix-items-added | INFO | WARN | |
| request-property-prefix-items-removed | ERR | WARN | |
| response-body-prefix-items-added | ERR | WARN | |
| response-body-prefix-items-removed | INFO | WARN | |
| response-property-prefix-items-added | ERR | WARN | |
| response-property-prefix-items-removed | INFO | WARN | |

A finer fix than WARN across the board: compare the prefix schemas and the
items schema for containment, as the multipleOf checks do with divisibility.

### B5. Remaining singles (5 rules)

| Rule | Stored | Derived | Root cause | Decision |
|---|---|---|---|---|
| response-media-type-name-changed | INFO | ERR | clients negotiating the old media type name break | |
| response-property-enum-value-added | WARN | ERR | server may emit a value clients never handled; WARN understates under the law | |
| request-parameter-default-value-added | ERR | INFO | defaults are server-side fallback, not contract (the #1109 reasoning); note the body/property default rules are INFO, so the parameter family is internally inconsistent either way | |
| request-parameter-default-value-changed | ERR | INFO | same | |
| request-parameter-default-value-removed | ERR | INFO | same | |

---

## Bucket C: model refinements the data forced

Not mismatches. These are law/effect-table decisions I made to explain the
registry; each needs sign-off because the derived levels above depend on them.

| # | Decision in the model | Justification | Decision |
|---|---|---|---|
| C1 | Negotiated variants (response statuses, media types, response headers) obey request polarity | the client chooses or relies on the variant, so removal breaks clients even on the response side; matches stored ERR for success-status-removed, required-header-removed, response-media-type-removed | |
| C2 | Adding/removing an optional property is effect None | with default additionalProperties the field was already allowed, so the accepted set is unchanged | |
| C3 | Security requirements are OR-alternatives, scopes AND within one | per the OpenAPI security model; makes security-added INFO correct and makes B1 findings | |
| C4 | readOnly/writeOnly are advisory (SHOULD-level) metadata | the spec does not make them validation constraints; all became-read-only/write-only transitions derive INFO | |
| C5 | Adding if/then/else narrows; removing widens | a dormant then/else is activated by if; conditionals only add constraints | |
| C6 | media-type-schema-added narrows / removed widens | no schema means any payload is accepted | |
| C7 | Lifecycle violations (removal before sunset, invalid or missing sunset, stability decreased) are ERR by policy | oasdiff's deprecation contract, distinct from the wire contract | |
| C8 | type-compatible is direction-relative | request-*-type-compatible means widened, response-*-type-compatible means narrowed; both are the safe side for their direction | |
| C9 | api-schema-removed (unused component) is effect None | an unused component is not part of the wire contract | |

---

## Bookkeeping

- Findings that already have issues: B3 is #1034. The min-items-set dead rules
  (#1171) intersect A1: three of its 24 rules can never fire at all.
- If a bucket-B group is confirmed as a bug, it gets its own issue and the fix
  changes the stored level to the derived one; the law test then passes for it
  without an entry anywhere.
- If a bucket-A group is kept, it becomes a deviation-ledger row in the PR A
  design, and the ledger test enforces that the deviation stays documented.
