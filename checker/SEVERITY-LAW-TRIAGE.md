# Severity-Law Triage

Working document for reviewing the 56 rules whose stored level disagrees with
the level derived by the severity law (see `rule_severity_law_test.go`, run
with `go test ./checker -run SeverityLaw`). 500 of 556 rules derive exactly;
every disagreement below has a root cause and a proposed handling.

Fill in the Decision column: **keep** (the stored level stands, and the entry
becomes a permanent ledger row with the stated reason) or **fix** (the stored
level changes to the derived one, and the entry is removed).

The law: Narrows x request = ERR, Narrows x response = INFO, Widens x request
= INFO, Widens x response = ERR, Incomparable = ERR, Unknown = WARN, None
(metadata) = INFO, lifecycle violation = ERR. Guards: readOnly nullifies a
request-side effect, writeOnly nullifies a response-side effect, sanctioned
removal (deprecated with an honoured sunset) = INFO, negotiated (a status,
media type, or header the client selects) applies request polarity.

---

## The two directions are not equally suspect

A deviation is either **below** the law (the rule reports a milder verdict
than the contract implies) or **above** it (a harsher one). The soundness
asymmetry oasdiff already follows makes these very different claims:

- **Below the law is a safety claim.** Reporting a breaking change as safe
  ships it silently, and the failure surfaces in production. A milder verdict
  therefore owes a proof that the change is safe for every consumer that
  conformed to the old contract.
- **Above the law costs a reviewer a glance.** A harsher verdict needs no
  proof, only a reason worth stating.

The distinction disqualifies a whole class of justification. "The restriction
is sometimes legitimately required", "it is only cleanup", "clients should not
rely on it" are reasons an author might *accept* a breaking change. They are
not reasons the change is not breaking, and they are true of every breaking
change: removing an endpoint is sometimes legitimately required too. Such a
judgment belongs to the operator, through `--severity-levels`, and to the
reviewer, who records it when approving a change in the review workflow at
oasdiff.com. It does not belong in the default verdict, where it silently
weakens the report for everyone.

So: the 16 rules above the law need a sentence each. The 40 below it each owe
a contract argument, and a reason that is really an exemption is a finding,
not a convention.

---

## Above the law (16 rules): conservative, ratify with a reason

| Rules | Stored | Derived | Reason | Decision |
|---|---|---|---|---|
| `request-body-removed` | ERR | WARN | **settled: keep.** See the reasoning below | keep |
| `request-parameter-removed`, `request-property-removed`, `response-optional-property-removed` | WARN | INFO | the property rows are above the law only because C2 calls the change safe, which it is not under `additionalProperties: false`; see C2 in more detail | |
| `request-body-wrapped-in-one-of`, `response-body-wrapped-in-one-of` | ERR | WARN | the check cannot verify the original branch survives the wrapping, and #1037 decided to keep the breaking verdict | |
| `request-body-all-of-removed`, `request-property-all-of-removed` | WARN | INFO | dropping a conjunct widens the accepted set; kept as a warning because the removed constraints are invisible in the diff | |
| `request-parameter-default-value-added/changed/removed` (3) | ERR | INFO | a default is server-side, not part of the contract, but the parameter family predates that reading; the body and property defaults are already INFO | |
| `request-body-prefix-items-removed`, `request-property-prefix-items-removed`, `response-body-prefix-items-added`, `response-property-prefix-items-added` | ERR | WARN | see the prefixItems finding below: the containment is undecided, so ERR over-reports rather than under-reports | |
| `response-required-property-became-not-write-only` | WARN | INFO | readOnly and writeOnly are advisory in the specification, so the law reads the change as metadata; the warning flags that the field now appears in responses | |

The parameter-default row is the one to look at twice: it is an escalation
only because the sibling body and property rules are INFO. Whichever way it
goes, the three families should agree.

### request-body-removed, settled

The effect was `widens`, which asserts that a request still carrying a body is
accepted once the declaration is gone. The specification does not say that: it
defines what `requestBody` describes when present, not what an absent one
permits. The containment is undecided, so the effect is `unknown` and the law
derives WARN.

ERR then stands on a contract argument rather than an exemption: the operation
withdraws a declared input, so a client that conformed by sending the body can
no longer convey that data through the operation, whatever becomes of the
request's validity. This is the request-side mirror of
`response-required-property-removed`, which is ERR because an element the
consumer relied on is gone, and it appeals to the declaration rather than to
what a server happens to tolerate.

What the containment question cannot express is the loss of a declared
capability, and that is the gap the ledger entry records.

One asymmetry surfaced while settling it, and it is not a severity question:
the add side splits by requiredness (`request-body-added-required` is ERR,
`request-body-added-optional` INFO, since only the first invalidates a
body-less request), while removal is one rule for both cases. Removal widens
either way, so the split would not change a verdict; it would let a team
downgrade the optional case alone with `--severity-levels`. The check already
holds the base operation, so the requiredness is in reach if that granularity
is wanted.

---

## Below the law (40 rules): each owes a contract argument

### The bound-set family (24 rules)

`request-body-max-length-set`, `request-property-min-items-set`,
`request-parameter-exclusive-max-set` and 21 more: **WARN**, derived **ERR**.

Setting a bound where none existed narrows the accepted set, so a request that
was valid can become invalid. The three families with the identical shape are
already ERR:

| Change | Effect | Stored |
|---|---|---|
| `request-body-became-enum` | narrows | error |
| `request-body-const-added` | narrows | error |
| `request-parameter-pattern-added` | narrows | error |
| `request-body-max-length-set` | narrows | **warning** |

The stated reason ("the restriction is sometimes legitimately required, for
security reasons or to correct an error in the specification") is an
exemption, not a contract argument, and it applies equally to enum, const and
pattern.

**Recommendation: fix.** 24 levels move WARN to ERR, and the 24 localized
`-comment` entries are reworded from justifying the warning to naming the
override: a team that knows its clients are within the new limit downgrades
the check with `--severity-levels`, or approves the change with a
justification in the review workflow. This is a user-visible escalation and
wants a release note.

### Security (4 rules)

`api-security-removed`, `api-global-security-removed`,
`api-security-scope-added`, `api-global-security-scope-added`: **INFO**,
derived **ERR**.

Security requirements are alternatives, so removing one breaks clients
authenticating with it; scopes within a requirement are conjunctive, so adding
one breaks clients whose tokens lack it. No reason is recorded for the INFO.
**Recommendation: fix**, tracked with the scheme-field gap in #1175.

### Response anyOf-added (2 rules)

`response-body-any-of-added`, `response-property-any-of-added`: **INFO**,
derived **ERR**, while `response-body-one-of-added` is already ERR for the
identical widening. **Recommendation: fix**, or explain what distinguishes the
two keywords.

### Response pattern (2 rules)

`response-property-pattern-removed` (INFO, derived ERR) and
`response-property-pattern-changed` (INFO, derived WARN): the law
independently re-derived the gap already tracked in #1034.
**Recommendation: fix** under that issue.

### prefixItems (4 of the 8 rules)

`request-*-prefix-items-added` and `response-*-prefix-items-removed`: **INFO**,
derived **WARN**. prefixItems reshapes positional constraints, so in general
neither accepted set contains the other; treating the change as a widening
makes these an unproven "safe". The sibling four are the same mistake in the
harsher direction (above the law).

**Recommendation: decide the containment rather than pick a level.** The
comparison already exists: `diff.SchemaRefsValidationEquivalent` answers the
wire-contract question about two schemas, and three checks already call it.
Ask it whether each prefix schema is equivalent to the items schema it
replaces; where it is, the change is proved safe, and where it is not, the
verdict falls to `unknown` (WARN) rather than to an assumed direction. Fixing
all eight to WARN is the blunt version of the same correction, and is the
fallback if the comparison turns out not to fit.

### Remaining singles (4 rules)

| Rule | Stored | Derived | Question it must answer | Decision |
|---|---|---|---|---|
| `response-media-type-name-changed` | INFO | ERR | a client negotiating the old media type gets no response; what makes that safe? | |
| `response-property-enum-value-added` | WARN | ERR | the server may now emit a value no client was written to handle | |
| `response-non-success-status-removed` | INFO | ERR | a client handling that status loses the behaviour; "it is only cleanup" is an exemption | |
| `optional-response-header-removed` | WARN | ERR | "clients should not rely on it" describes what clients ought to do, not what the contract said | |

---

## Model refinements the derivations depend on

Not deviations. These are decisions in the law itself; each one changes which
rules appear above, so a verdict here can move entries between the sections.

| # | Decision | Justification | Decision |
|---|---|---|---|
| C1 | A status, media type, or header the client selects (`negotiated`) applies request polarity | the client chooses or relies on the variant, so removing it breaks clients even though it lives on the response side | |
| C2 | Adding or removing an optional property is effect None | **unsound as written.** It holds while `additionalProperties` is unrestricted, but with `additionalProperties: false` a removed optional property is rejected where it used to be accepted, which narrows a request. The claim was reached by failing to find a reason it was unsafe rather than by proving it safe. See the note below | |
| C3 | Security requirements are alternatives; scopes within one are conjunctive | per the OpenAPI security model; this is what makes the security findings above | |
| C4 | readOnly and writeOnly are advisory metadata | the specification does not make them validation constraints | |
| C5 | Adding if/then/else narrows; removing widens | a dormant then or else is activated by if, so conditionals only add constraints | |
| C6 | A media type schema appearing narrows, disappearing widens | with no schema, any payload is accepted | |
| C7 | Lifecycle violations (removal before sunset, invalid or missing sunset, stability decreased) are ERR by policy | oasdiff's deprecation contract, distinct from the wire contract | |
| C8 | `type-compatible` is direction-relative | on a request it means widened, on a response narrowed: the safe side for each | |
| C9 | Removing an unused component schema is effect None | an unused component is not part of the wire contract | |

### C2 in more detail

`request-property-removed` and `response-optional-property-removed` appear
above the law only because C2 calls the change safe. Under
`additionalProperties: false` it is not: the removed property is no longer
accepted, so a request that carried it becomes invalid. The stored WARN is
the sound verdict and the model is what is wrong.

Two ways to correct it:

- **Cheap:** make the removal case `unknown`. Both rules then derive WARN,
  match their stored level, and leave the ledger honestly.
- **Precise:** make `additionalProperties` a guard, so a closed object gives
  `narrows` (ERR on a request) and an open one gives `none`. This is what
  guards are for, and it needs the checks to read the sibling keyword.

The addition case is different and deliberate: adding an optional property to
a response is reported INFO, which assumes tolerant readers. That is a stance
worth stating rather than a proof, and it belongs in the ledger if it stays.

---

## Bookkeeping

- Findings that already have issues: the response pattern rules are #1034, the
  security family joins #1175. The bound-set family overlaps #1171, where
  three of its 24 rules cannot fire at all.
- A **fix** changes the stored level to the derived one and removes the ledger
  entry, so the law then holds for that rule with nothing recorded.
- A **keep** turns the entry's reason into a permanent one, and the test keeps
  it honest: the entry fails the build if the rule's level ever matches the
  law again.
