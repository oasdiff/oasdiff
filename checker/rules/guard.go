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
	// GuardNegotiated: the rule judges the availability of something the
	// client selects or relies on, such as a response status, media type,
	// or header, or the payload that inhabits one. Removing an option the
	// client chooses rejects that choice the way narrowing a request does,
	// and adding one harms nobody, so the level is derived as if the
	// element were on the request side: narrowing is breaking, widening is
	// not, the reverse of plain response polarity.
	GuardNegotiated Guard = "negotiated"
)
