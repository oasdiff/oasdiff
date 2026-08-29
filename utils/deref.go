package utils

// Deref returns the pointed-to value, or nil for a nil pointer. The any
// return keeps "absent" (nil) distinguishable from a zero value, which a
// plain *T dereference cannot express.
func Deref[T any](ref *T) any {
	if ref == nil {
		return nil
	}
	return *ref
}
