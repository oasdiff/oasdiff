package schemawalk

import (
	"slices"
	"strings"

	"github.com/oasdiff/oasdiff/diff"
)

// References are the schemas below a root that more than one reference
// reaches. A walk continues through the first reference to a schema only, so
// the others are known from here.
type References struct {
	shared []sharedSchema
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
	type named struct{ path, name string }
	var names []named
	for _, visit := range order {
		if name := componentName(visit.schemaDiff); name != "" {
			names = append(names, named{reaches[visit][0], name})
		}
	}

	var result References
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
		name, depth := "", -1
		for _, n := range names {
			if under(paths[0], n.path) && len(n.path) > depth {
				name, depth = n.name, len(n.path)
			}
		}
		result.shared = append(result.shared, sharedSchema{name: name, paths: paths})
	}
	return result
}

// At finds the innermost schema several references reach that a change
// reported at path is in or below. It returns the schema's name, empty for a
// schema with none, and the path to the change through each reference, the
// given path first.
func (r References) At(path string) (name string, paths []string, ok bool) {
	var found *sharedSchema
	for i := range r.shared {
		s := &r.shared[i]
		if under(path, s.paths[0]) && (found == nil || len(s.paths[0]) > len(found.paths[0])) {
			found = s
		}
	}
	if found == nil {
		return "", nil, false
	}

	below := path[len(found.paths[0]):]
	paths = []string{path}
	for _, p := range found.paths {
		if other := p + below; !slices.Contains(paths, other) {
			paths = append(paths, other)
		}
	}
	return found.name, paths, true
}

// under reports whether path is at or below the schema the walk reached at
// prefix. The root is reached at the empty path, and everything is below it.
func under(path, prefix string) bool {
	return prefix == "" || path == prefix || strings.HasPrefix(path, prefix+"/")
}

// componentName prefers the revision's name: it is the one a reader will find
// in the spec they are reviewing.
func componentName(schemaDiff *diff.SchemaDiff) string {
	if schemaDiff.RevisionComponent != "" {
		return schemaDiff.RevisionComponent
	}
	return schemaDiff.BaseComponent
}
