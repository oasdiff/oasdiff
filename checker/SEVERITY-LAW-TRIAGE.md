# Severity-Law Triage

Working document for reviewing the rules whose stored level disagrees with the
level derived by the severity law (see `rule_severity_law_test.go`, run with
`go test ./checker -run SeverityLaw`). 9 of 558 rules disagree; every
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

So: the rules above the law need a sentence each. The rules below it each owe
a contract argument, and a reason that is really an exemption is a finding,
not a convention.

---

## Above the law (4 rules): conservative, ratify with a reason

Eleven entries left this section while it was being reviewed.
`request-parameter-removed` and `request-property-removed` were above the law
only because their effects claimed a safety nobody had proved (see *Removals
that cannot be proved safe* below); they now derive their stored WARN.
`response-optional-property-removed` was **fixed to INFO**: a conforming client
already tolerates the property's absence, so the contract does not break, and a
client that treated an optional property as guaranteed is out of specification
rather than broken by the change. Reporting it as potentially breaking would
fail a gate over a client-side defect. `request-body-all-of-removed`,
`request-property-all-of-removed` and the three parameter-default rules were
**fixed to INFO** (see below).

| Rules | Stored | Derived | Reason | Decision |
|---|---|---|---|---|
| `request-body-prefix-items-removed`, `request-property-prefix-items-removed`, `response-body-prefix-items-added`, `response-property-prefix-items-added` | ERR | WARN | the containment is undecided, so ERR over-reports rather than under-reports; the fix is to decide it (see the prefixItems finding) | |

### became-not-write-only, settled: fixed

WARN, derived INFO, now **INFO**. It was the lone outlier in a family of
fourteen mutability rules (request and response, optional and required, across
readOnly and writeOnly), every other one of which is `none` with the matching
guard and INFO.

Removing `writeOnly` means the response may now carry a field it did not, which
is what adding a property does, and both `response-required-property-added` and
`response-optional-property-added` are INFO. The comment it carried, that the
change is valid only if the property was always returned before, is a note
about whether the specification described past behaviour accurately, not a
question about what a conforming client sees; it is dropped with the WARN.

C4 does not have to be ruled on to settle this, but a related question is left
open. `GuardWriteOnly` describes the base state, and the four `became-not-X`
rules fire precisely when that state ends, so nullifying response-side effects
on them is backwards. It changes no verdict today because the family is
uniformly `none`, and reshaping eight registrations wants its own pass.

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

### request-body-removed, settled: the effect was wrong

The effect was `widens`, which asserts that a request still carrying a body is
accepted once the declaration is gone. The specification does not say that. It
was then read as `unknown`, on the argument that removing the whole declaration
leaves nothing to compare, and ERR was kept above the derived WARN on the
grounds that the operation withdraws a declared input.

Both readings were wrong, and containment shows why. Removing the request body
removes every media type it declared, and `request-body-media-type-removed` is
`narrows` and ERR. So the change strictly contains an ERR-level change, and
`TestSeverityMonotonicity` already pairs the two. "Nothing left to compare" was
an artifact of comparing the declaration to its absence rather than the media
types to theirs.

The effect is `narrows`, which derives ERR on a request, so nothing is recorded
in the ledger. The narrowing is real at the media-type level: a content type
absent from the map has a specified rejection, 415, so the request genuinely
may become invalid. That is what separates this from `request-parameter-removed`,
which stays `unknown` because nothing in the specification says a stray query
parameter is rejected.

Only `request-body-removed` fires on a whole-body removal; the media-type rule
does not also report. So its level is the only verdict the change gets.

One asymmetry surfaced while settling it, and it is not a severity question:
the add side splits by requiredness (`request-body-added-required` is ERR,
`request-body-added-optional` INFO, since only the first invalidates a
body-less request), while removal is one rule for both cases. Both cases lose
the same declared input, so the split would not change a verdict; it would let
a team downgrade the optional case alone with `--severity-levels`. The check
already holds the base operation, so the requiredness is in reach if that
granularity is wanted. `TestRequestBodyRemovedOptional` now pins the current
behaviour, so a split has to be a decision rather than a drift.

Two rules at `paths.*.*.requestBody` also disagreed on their area:
`request-body-removed` said `schema` while every other rule on that subtree,
its own add-side mirror included, says `request-body`. Corrected, which is what
lets `TestRuleSymmetry` compare the pair at all.

### wrapped-in-one-of, settled: fixed

The effect was `unknown` because the check never compared the alternatives
against the original schema. That was an implementation gap, not spec
ambiguity: the information is in the document, and
`diff.SchemaRefsValidationEquivalent` answers exactly the question, so the
verdict now follows from the answer rather than from caution.

`getOneOfWrappingDiff` asks whether any alternative has the base schema's
validation contract, and the two checks split on it:

- No alternative does, so a payload that was valid may match nothing. That is
  `incomparable`, which derives the ERR the rules already reported.
  `request-body-wrapped-in-one-of` and `response-body-wrapped-in-one-of` keep
  their ids and their level, now earned rather than assumed.
- One does, so every payload the base accepted still matches it, and the only
  remaining hazard is `oneOf` rejecting a payload that matches a second
  alternative as well. The spec does not say whether the alternatives overlap,
  which is `unknown`, deriving WARN. Reported as the new
  `request-body-wrapped-in-one-of-original-preserved` and
  `response-body-wrapped-in-one-of-original-preserved`, with a comment naming
  the missing information.

The residual is deliberate. Proving the second case safe needs branch
disjointness, which is not equivalence and which oasdiff cannot decide in
general. `isNullableWrap` shows the shape of an answer for the one case where
it is provable, a bare `null` branch against a base that rejects null, and
`discriminator` is not one: the spec makes it a deserialization aid, not a
validation constraint, so a payload matching two branches still fails. WARN
rather than INFO is what the soundness asymmetry requires.

This is a user-visible downgrade for the preserved case: a wrapping that keeps
the original schema now warns where it previously failed `oasdiff breaking`.

### all-of-removed, settled: fixed

`request-body-all-of-removed` and `request-property-all-of-removed` were WARN
and are now INFO. The subschemas of an `allOf` are conjunctive, so dropping one
removes constraints and every payload that was valid still is: the widening is
provable from the document, which is exactly what INFO asks for.

The warning existed because a relaxed constraint is worth a reader's attention,
not because the change can invalidate a request. That is a reporting concern,
and the message already names the subschema that was dropped. The response-side
siblings were INFO all along, so this also removes an asymmetry the law had no
reason to allow.

### The parameter defaults, settled: fixed

`request-parameter-default-value-added/changed/removed` were ERR while the body
and property equivalents were INFO. They are now INFO, so the three families
agree.

The derivation calls a default metadata: it does not change which payloads are
valid, so every request that conformed to the old contract still conforms. The
argument for ERR was not about validity but about meaning, that a client
omitting the parameter had a documented behaviour and the change alters what
its unchanged request does. That is a real effect, and it is the same effect a
body default has, which #1109 already settled as a server-side fallback rather
than a contract term.

The changes are still reported in the changelog, where a reader who cares about
the fallback will see them.

---

## Below the law (5 rules): each owes a contract argument

### The bound-set family (24 rules) — settled: fixed

`request-body-max-length-set`, `request-property-min-items-set`,
`request-parameter-exclusive-max-set` and 21 more were **WARN**, derived
**ERR**, and are now ERR.

Setting a bound where none existed narrows the accepted set, so a request that
was valid can become invalid. Three families with the identical shape were
already ERR (`became-enum`, `const-added`, `pattern-added`), and the stated
reason for the exception ("the restriction is sometimes legitimately required,
for security reasons or to correct an error in the specification") is a reason
an author might accept the break, not a reason it is not a break. It is true
of every breaking change.

The 24 localized comments were reworded to match: they no longer explain a
warning, and instead name the override, since a team that knows its clients
stay within the new limit is the party entitled to decide that.

This is a user-visible escalation: a spec that adds a bound now fails
`oasdiff breaking` where it previously passed with a warning. Teams that
accept the change can lower the check with `--severity-levels` or approve it
in review.

### Security (4 rules) — settled: fixed

`api-security-removed`, `api-global-security-removed`,
`api-security-scope-added`, `api-global-security-scope-added` were **INFO** and
are now **ERR**.

Security requirements are alternatives, so removing one leaves a client
authenticating with it unable to call the operation; scopes within a
requirement are conjunctive, so adding one rejects a client whose token lacks
it. Both are failures for a client that conformed to the old contract. No
reason was ever recorded for the INFO. The scheme-field gap in the same family
stays open in #1175.

### Response anyOf-added (2 rules) — settled: fixed

`response-body-any-of-added` and `response-property-any-of-added` were **INFO**
and are now **ERR**, matching `response-body-one-of-added`, which was already
ERR for the identical change. Adding a branch to a response union lets the
server emit a payload the old contract did not allow, so a client that
validated responses against it rejects them. Nothing distinguishes `anyOf` from
`oneOf` here.

### Response pattern (2 rules) — settled: fixed

`response-property-pattern-removed` was **INFO** and is now **ERR**;
`response-property-pattern-changed` was **INFO** and is now **WARN**. Removing
a pattern lets the response carry values the old contract excluded, and a
changed pattern is undecidable between the two regular languages, hence the
warning. This is the gap already tracked in #1034, which the law re-derived
independently.

### optional-response-header-removed, settled: fixed

WARN, derived ERR, now **INFO with effect `none`**, the exact mirror of
`response-optional-property-removed`. An optional header's absence is already
permitted by the old contract, so its removal breaks no conforming client, and
a client that treated it as guaranteed is out of specification rather than
broken by the change. The two rules previously modelled the same situation
differently, the header one carrying `narrows` with the negotiated guard; the
guard is gone with the effect, since nothing about an optional header is
client-selected.

### response-media-type-name-changed, settled: fixed

INFO, derived ERR from `incomparable`, now **WARN with effect `unknown`** and a
comment.

The id is narrower than its name suggests. `ResponseMediaTypeNameUpdatedCheck`
reports it only when the media type *parameters* differ; any other name change
goes to `-generalized` or `-specialized`. And it reports it for a parameter
appearing, disappearing or changing value alike, which are a narrowing, a
widening and an incomparable change respectively. So `incomparable` claimed a
proof the check never had: all it observed is that the parameter maps differ.

`unknown` is the honest reading and derives WARN. Splitting the id by which of
the three happened would let each case earn its own verdict, tracked as #1186.

### response-property-enum-value-added, settled: fixed

WARN, derived ERR, now **ERR**. A response `enum` is a promise about what the
server emits, so adding a value lets it return something the old contract
excluded and a client written against that contract may not handle it. The
family already agreed: `response-property-any-of-added` and
`response-body-list-of-types-widened` are both ERR for the same widening.

The comment is reworded to match. It used to say the change "can be unexpected
for clients", which reads as a warning; it now states what the server may do
and names `x-extensible-enum`, which is the declaration to use when the value
set is meant to grow.

This is a user-visible escalation on a common change.

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

### Remaining singles (1 rule)

| Rule | Stored | Derived | Question it must answer | Decision |
|---|---|---|---|---|
| `response-non-success-status-removed` | INFO | ERR | a client handling that status loses the behaviour; "it is only cleanup" is an exemption | |

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

- Findings that already have issues: the response pattern rules were #1034,
  now fixed; the scheme-field gap in the security family stays in #1175. The
  bound-set family overlaps #1171, where three of its 24 rules cannot fire at
  all.
- A **fix** changes the stored level to the derived one and removes the ledger
  entry, so the law then holds for that rule with nothing recorded.
- A **keep** turns the entry's reason into a permanent one, and the test keeps
  it honest: the entry fails the build if the rule's level ever matches the
  law again.
