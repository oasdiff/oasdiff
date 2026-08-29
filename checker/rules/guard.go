package rules

// Guard is a named predicate over the document state that a rule requires
// before it fires. Rules covering the same edits may carry different
// Effects and severities because their guards select different document
// states.
type Guard string

const (
	// GuardReadOnly: the changed property is readOnly, so it does not
	// appear in requests; request-side effects are nullified.
	GuardReadOnly Guard = "read-only"
	// GuardWriteOnly: the changed property is writeOnly, so it does not
	// appear in responses; response-side effects are nullified.
	GuardWriteOnly Guard = "write-only"
	// GuardSanctioned: the removed element was deprecated and its sunset
	// period was honored, so the removal follows the deprecation contract.
	GuardSanctioned Guard = "sanctioned"
	// GuardNonSuccess: the affected response status is a non-success
	// status. The responses map does not promise that the server returns
	// only the statuses it lists, so neither documenting one more nor
	// dropping one changes what a conforming client can receive; the
	// effect is nullified.
	GuardNonSuccess Guard = "non-success"
	// GuardHasDefault: the changed element declares a default value.
	GuardHasDefault Guard = "has-default"
	// GuardNegotiated: the rule judges the availability of a client-selected
	// variant (a response status, media type, or header). The client chooses
	// or relies on the variant, so contravariance applies with request
	// polarity even though the variant lives on the response side.
	GuardNegotiated Guard = "negotiated"
)
