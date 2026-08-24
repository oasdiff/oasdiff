# Severity-Law Triage

Working document for reviewing the rules whose stored level disagrees with the
level derived by the severity law (see `rule_severity_law_test.go`, run with
`go test ./checker -run SeverityLaw`). 54 of 556 rules disagree; every
disagreement below has a root cause and a proposed handling.

Settling one often corrects the model rather than the rule: an effect that
claims a safety nobody proved is the finding, and the stored level is usually
sound. Three questions settle an entry: is the effect actually proved? If not,
what is the honest one? And does any remaining gap rest on the contract or on
an exemption?

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

## Above the law (13 rules): conservative, ratify with a reason

Three entries left this section while it was being reviewed.
`request-parameter-removed` and `request-property-removed` were above the law
only because their effects claimed a safety nobody had proved (see *Removals
that cannot be proved safe* below); they now derive their stored WARN.
`response-optional-property-removed` was **fixed to INFO**: a conforming client
already tolerates the property's absence, so the contract does not break, and a
client that treated an optional property as guaranteed is out of specification
rather than broken by the change. Reporting it as potentially breaking would
fail a gate over a client-side defect.

| Rules | Stored | Derived | Reason | Decision |
|---|---|---|---|---|
| `request-body-removed` | ERR | WARN | the operation withdraws a declared input; see below | **keep** |
| `request-body-wrapped-in-one-of`, `response-body-wrapped-in-one-of` | ERR | WARN | the check does not verify that the original branch survives the wrapping, so #1037 kept the breaking verdict; see below | |
| `request-body-all-of-removed`, `request-property-all-of-removed` | WARN | INFO | dropping a conjunct strictly widens the accepted set, which is provable, so INFO is the sound verdict; the warning exists because the dropped constraints are invisible in the diff, which is a reporting concern rather than a contract one | |
| `request-parameter-default-value-added/changed/removed` (3) | ERR | INFO | see below: this one needs a product decision, not a derivation | |
| `request-body-prefix-items-removed`, `request-property-prefix-items-removed`, `response-body-prefix-items-added`, `response-property-prefix-items-added` | ERR | WARN | the containment is undecided, so ERR over-reports rather than under-reports; the fix is to decide it (see the prefixItems finding) | |
| `response-required-property-became-not-write-only` | WARN | INFO | the effect says metadata, on the reading that writeOnly is advisory; but the change does mean the response may now carry a property it did not, which is a response widening and would derive ERR. C4 is as unproven as C2 was | |

### Removals that cannot be proved safe

`request-parameter-removed` and `request-property-removed` were marked `widens`
and `none`. Neither is provable:

- Nothing in the specification says an undeclared query parameter is rejected,
  so a request still carrying the removed parameter cannot be shown to remain
  valid.
- With `additionalProperties: false`, a removed property is rejected where it
  used to be accepted, which narrows the request outright.

Both effects are now `unknown`, which derives WARN and matches what the rules
already reported. The stored levels were sound and the model was not, which is
the same correction C2 needs.

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

### wrapped-in-one-of, and what would settle it

The effect is `unknown` because the check does not compare the wrapped branch
against the original schema. That is an implementation gap, not spec
ambiguity: the information is in the document. `SchemaRefsValidationEquivalent`
would answer it, as it would for prefixItems, and the verdict would follow from
the answer rather than from caution: a wrapping that preserves the original
branch widens, and one that does not is incomparable and breaking either way.
Until then ERR is the conservative reading, which is the correct side to err on.

### The parameter defaults need a product decision

`request-parameter-default-value-added/changed/removed` are ERR while the body
and property equivalents are INFO. The derivation calls a default metadata,
since it does not change which payloads are valid.

The argument for ERR is not about validity: a client that omits the parameter
had a documented meaning for that omission, and changing the default changes
what its unchanged request does. The argument for INFO is that a default is a
server-side fallback, which is where #1109 landed for required-with-default.

Both readings are defensible, and they apply equally to body properties, so
whichever is chosen the three families should agree. This is the one entry in
this section that a derivation cannot settle.

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
