package schemawalk

import "github.com/oasdiff/oasdiff/diff"

// Negations records which schemas inside a payload or parameter schema are
// under an odd number of `not` keywords.
type Negations struct {
	// byPath holds, for each path the walk reached a schema at, whether the
	// schema there is negated.
	byPath map[string]bool
}

// NewNegations walks a payload or parameter schema and records the paths
// that are under an odd number of `not` keywords.
func NewNegations(root *diff.SchemaDiff) Negations {
	result := Negations{byPath: map[string]bool{}}
	// A property named "not" shares its path with the `not` keyword beside
	// it. Such a path is recorded as not negated.
	ambiguous := map[string]bool{}
	Walker{arrive: func(path string, visit walkVisit) {
		if negated, ok := result.byPath[path]; ok && negated != visit.negated {
			ambiguous[path] = true
		}
		result.byPath[path] = visit.negated
	}}.Walk(root)
	for path := range ambiguous {
		result.byPath[path] = false
	}
	return result
}

// At reports whether a change reported at path is under an odd number of
// `not` keywords. A path the walk did not reach, such as an added property,
// takes the answer of the closest schema above it.
func (n Negations) At(path string) bool {
	return innermost(n.byPath, path)
}
