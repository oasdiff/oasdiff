// Package nullability answers whether a schema accepts null and whether that
// changed between two versions of it. It reads schemas only, so the diff, the
// checks and the linter share one reading of the nullable keyword, a "null"
// entry in the type array, and a oneOf or anyOf with a branch that accepts
// null.
package nullability
