// Package nullability answers whether a schema accepts null and whether that
// changed between two versions of it. It reads schemas only, so the diff, the
// checks and the linter share one reading of the nullable keyword, a "null"
// entry in the type array, and a oneOf or anyOf with a branch that accepts
// null.
//
// The rules are the same for OpenAPI 3.0 and OpenAPI 3.1 documents:
//
//   - A schema with a type accepts null when the type names it: nullable: true
//     beside the type (OpenAPI 3.0) or a "null" entry in the type array
//     (OpenAPI 3.1).
//   - A schema with no type accepts every type, null included. OpenAPI 3.0.3
//     says so for 3.0 documents too: an untyped schema allows null, and
//     nullable: false, explicit or by default, does not prohibit it. Some
//     validators reject null there, and this package follows the
//     specification.
//   - An enum lets null through only when null is one of its values, and a
//     const only when it is null.
//   - A oneOf accepts null while exactly one of its branches does, and an
//     anyOf when any branch does. The keywords of the schema that holds the
//     composition must let null through as well.
//   - nullable: true beside a oneOf or anyOf accepts null whatever the
//     branches do. This is the one place the package departs from OpenAPI
//     3.0.3, which gives nullable no effect without a type in the same
//     schema. 3.0 documents declare a nullable composition this way, and it
//     is read as they mean it.
//   - A schema with an allOf, a not or an if is not decided.
package nullability
