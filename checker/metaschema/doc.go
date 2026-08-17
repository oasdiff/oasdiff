// Package metaschema enumerates the edit space of an OpenAPI document:
// every field location in the OpenAPI object model, paired with the
// syntactic edit actions applicable at that location. The enumeration is
// derived by reflection from the kin-openapi types oasdiff parses specs
// into, so a field added upstream shows up as new edits without a manual
// update here.
//
// Actions are syntactic (add, remove, set, unset, change, increase,
// decrease), not verdicts: whether an edit at a edit is breaking is the
// checker's judgment, recorded by mapping rules to edits. Polarity is the
// syntactic position in the document (request, response, shared, document);
// semantic direction inversions (callbacks, webhooks, not-schemas,
// readOnly/writeOnly) are likewise the checker's business.
//
// Schema is a recursive node: the walk stops when a type reappears on the
// current path, so a edit like paths.*.*.requestBody.content.*.schema.maxLength
// stands for that keyword at any nesting depth inside the request body schema.
package metaschema
