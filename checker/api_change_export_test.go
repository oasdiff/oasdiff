package checker

// WithoutSchemaDiffs clears the fields that point into the diff. A test in
// checker_test cannot set them on its expected value, so it clears them on the
// actual one before comparing.
func (c ApiChange) WithoutSchemaDiffs() ApiChange {
	c.schema, c.root, c.propertyPath = nil, nil, ""
	return c
}
