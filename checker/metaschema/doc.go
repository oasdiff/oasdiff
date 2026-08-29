// Package metaschema enumerates every possible edit of an OpenAPI document:
// each field location in the OpenAPI object model, paired with the syntactic
// edit actions applicable at that location.
//
// The enumeration is derived independently of the checker's rules, which is
// the point of the package: it is the yardstick the rules are measured
// against (see checker/coverage). It comes by reflection from the
// kin-openapi types oasdiff parses specs into, so a field added upstream
// shows up as new edits without a manual update here.
//
// Actions are syntactic (add, remove, set, unset, change, increase,
// decrease), not verdicts: whether an edit is breaking is the checker's
// judgment. Polarity is the syntactic position in the document (request,
// response, shared, document); semantic direction inversions (callbacks,
// webhooks, not-schemas, readOnly/writeOnly) are likewise the checker's
// business. The exception is NonContracts, which records the fields whose
// edits cannot change which payloads are valid: that is a fact about the
// object model, so it belongs with the model.
//
// Schema is a recursive node: the walk stops when a type reappears on the
// current path, so an edit like paths.*.*.requestBody.content.*.schema.maxLength
// stands for that keyword at any nesting depth inside the request body schema.
package metaschema
