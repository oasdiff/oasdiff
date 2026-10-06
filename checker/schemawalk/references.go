package schemawalk

import (
	"iter"
	"math"
	"slices"
	"strings"

	"github.com/oasdiff/oasdiff/diff"
)

// References records the schemas below a root that are referenced more than
// once. The walk goes into such a schema only through its first reference, so
// this is where the other references are kept.
type References struct {
	shared []sharedSchema
	// byPath finds a shared schema by the path the walk reached it at first.
	byPath map[string]*sharedSchema
	// counted holds the number of paths through each shared schema, filled in
	// as changes ask.
	counted map[*sharedSchema]counted
}

type sharedSchema struct {
	name string
	// paths reach the schema, one per reference, in walk order. The first is
	// the one a walk continues through.
	paths []string
}

// NewReferences walks root and records the schemas several references reach.
func NewReferences(root *diff.SchemaDiff) References {
	reaches := map[walkVisit][]string{}
	var order []walkVisit
	Walker{arrive: func(path string, visit walkVisit) {
		if _, ok := reaches[visit]; !ok {
			order = append(order, visit)
		}
		reaches[visit] = append(reaches[visit], path)
	}}.Walk(root)

	// A name for each schema that has one, at the path the walk took to it.
	names := map[string]string{}
	for _, visit := range order {
		if name := componentName(visit.schemaDiff); name != "" {
			if _, ok := names[reaches[visit][0]]; !ok {
				names[reaches[visit][0]] = name
			}
		}
	}

	result := References{byPath: map[string]*sharedSchema{}, counted: map[*sharedSchema]counted{}}
	for _, visit := range order {
		paths := reaches[visit]
		if len(paths) < 2 {
			continue
		}
		// A shared schema written inline has no name of its own. It is shared
		// because the schema it belongs to was copied, as the parser does for a
		// $ref with a description beside it, and the copies keep its children,
		// so it takes the name of the innermost named schema the walk passed
		// through on the way to it.
		name := innermost(names, paths[0])
		result.shared = append(result.shared, sharedSchema{name: name, paths: paths})
	}
	for i := range result.shared {
		if _, ok := result.byPath[result.shared[i].paths[0]]; !ok {
			result.byPath[result.shared[i].paths[0]] = &result.shared[i]
		}
	}
	return result
}

// Shared describes the paths a change reaches through shared schemas.
type Shared struct {
	// Name is the name of the innermost shared schema the change is in or
	// below, empty for a schema with none.
	Name string
	// Paths are the first paths the change is at, in walk order, the given
	// path first.
	Paths []string
	// Count is the number of paths the change is at. It is zero when Cyclic.
	Count int
	// Cyclic is set when a cycle makes the number of paths unbounded.
	Cyclic bool
}

// At describes the paths a change reported at path is at, listing at most
// limit of them. It reports false when the change is in no shared schema.
//
// The walk goes into a shared schema once, so every recorded path runs through
// the first reference to each shared schema around it. The other paths are
// found by substituting each recorded reference for the first, at every shared
// schema around the path. Counting adds up the references instead of listing
// the paths, which can multiply into millions.
func (r References) At(path string, limit int) (Shared, bool) {
	found := r.enclosing(path, false)
	if found == nil {
		return Shared{}, false
	}

	shared := Shared{Name: found.name, Paths: []string{path}}
	for other := range r.pathsAt(path, false) {
		if len(shared.Paths) >= limit {
			break
		}
		if !slices.Contains(shared.Paths, other) {
			shared.Paths = append(shared.Paths, other)
		}
	}

	count, ok := r.countAt(path, false, map[*sharedSchema]bool{})
	if ok {
		shared.Count = count
	} else {
		shared.Cyclic = true
	}
	return shared, true
}

// enclosing returns the innermost shared schema that path is in or below. When
// strict is set, a schema reached at path itself does not count.
func (r References) enclosing(path string, strict bool) *sharedSchema {
	if strict {
		if path == "" {
			return nil
		}
		if i := strings.LastIndex(path, "/"); i >= 0 {
			path = path[:i]
		} else {
			path = ""
		}
	}
	return innermost(r.byPath, path)
}

// pathsAt yields every path that path stands for, in walk order. Through a
// cycle there is no end to them, so a caller stops when it has enough.
func (r References) pathsAt(path string, strict bool) iter.Seq[string] {
	return func(yield func(string) bool) {
		s := r.enclosing(path, strict)
		if s == nil {
			yield(path)
			return
		}
		below := strings.TrimPrefix(strings.TrimPrefix(path, s.paths[0]), "/")
		for _, reference := range s.paths {
			for at := range r.pathsAt(reference, true) {
				if below != "" {
					at = PropertyFullName(at, below)
				}
				if !yield(at) {
					return
				}
			}
		}
	}
}

type counted struct {
	count   int
	bounded bool
}

// countAt is the number of paths path stands for, and false when a cycle makes
// it unbounded. visiting holds the shared schemas being counted further up. A
// count too large for an int is capped at math.MaxInt.
func (r References) countAt(path string, strict bool, visiting map[*sharedSchema]bool) (int, bool) {
	s := r.enclosing(path, strict)
	if s == nil {
		return 1, true
	}
	if c, ok := r.counted[s]; ok {
		return c.count, c.bounded
	}
	if visiting[s] {
		return 0, false
	}
	visiting[s] = true
	defer delete(visiting, s)

	count, bounded := 0, true
	for _, reference := range s.paths {
		n, ok := r.countAt(reference, true, visiting)
		if !ok {
			count, bounded = 0, false
			break
		}
		count = min(count, math.MaxInt-n) + n
	}
	r.counted[s] = counted{count: count, bounded: bounded}
	return count, bounded
}

// innermost returns the value at the longest of path and its prefixes that
// has one.
func innermost[V any](byPath map[string]V, path string) V {
	for {
		if v, ok := byPath[path]; ok {
			return v
		}
		if path == "" {
			var zero V
			return zero
		}
		if i := strings.LastIndex(path, "/"); i >= 0 {
			path = path[:i]
		} else {
			path = ""
		}
	}
}

// componentName prefers the revision's name: it is the one a reader will find
// in the spec they are reviewing.
func componentName(schemaDiff *diff.SchemaDiff) string {
	if schemaDiff.RevisionComponent != "" {
		return schemaDiff.RevisionComponent
	}
	return schemaDiff.BaseComponent
}
