package metaschema_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker/metaschema"
)

func editSet(t *testing.T) map[metaschema.Edit]bool {
	t.Helper()
	set := map[metaschema.Edit]bool{}
	for _, c := range metaschema.Edits() {
		if set[c] {
			t.Errorf("duplicate edit: %+v", c)
		}
		set[c] = true
	}
	return set
}

func find(set map[metaschema.Edit]bool, location string, action metaschema.Action) (metaschema.Edit, bool) {
	for c := range set {
		if c.Location == location && c.Action == action {
			return c, true
		}
	}
	return metaschema.Edit{}, false
}

func TestCube_Deterministic(t *testing.T) {
	a, b := metaschema.Edits(), metaschema.Edits()
	if len(a) != len(b) {
		t.Fatalf("edit enumeration size changed between runs: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("edit enumeration order changed between runs at %d: %+v vs %+v", i, a[i], b[i])
		}
	}
}

func TestCube_KnownCells(t *testing.T) {
	set := editSet(t)

	for _, tc := range []struct {
		location string
		action   metaschema.Action
		polarity metaschema.Polarity
	}{
		// operations collapse to paths.*.*
		{"paths.*", metaschema.ActionRemove, metaschema.PolarityDocument},
		{"paths.*.*", metaschema.ActionAdd, metaschema.PolarityDocument},
		{"paths.*.*.deprecated", metaschema.ActionSet, metaschema.PolarityDocument},
		// request side
		{"paths.*.*.parameters.*", metaschema.ActionAdd, metaschema.PolarityRequest},
		{"paths.*.*.parameters.*.required", metaschema.ActionSet, metaschema.PolarityRequest},
		{"paths.*.*.parameters.*.schema.enum", metaschema.ActionRemove, metaschema.PolarityRequest},
		{"paths.*.*.requestBody", metaschema.ActionSet, metaschema.PolarityRequest},
		{"paths.*.*.requestBody.content.*.schema.maxLength", metaschema.ActionDecrease, metaschema.PolarityRequest},
		{"paths.*.*.requestBody.content.*.schema.required", metaschema.ActionAdd, metaschema.PolarityRequest},
		{"paths.*.*.requestBody.content.*.schema.type", metaschema.ActionRemove, metaschema.PolarityRequest},
		{"paths.*.*.requestBody.content.*.schema.additionalProperties", metaschema.ActionChange, metaschema.PolarityRequest},
		{"paths.*.*.requestBody.content.*.schema.exclusiveMinimum", metaschema.ActionIncrease, metaschema.PolarityRequest},
		// response side
		{"paths.*.*.responses.*", metaschema.ActionRemove, metaschema.PolarityResponse},
		{"paths.*.*.responses.*.content.*.schema.properties.*", metaschema.ActionRemove, metaschema.PolarityResponse},
		{"paths.*.*.responses.*.headers.*", metaschema.ActionRemove, metaschema.PolarityResponse},
		// shared components and document-level objects
		{"components.schemas.*", metaschema.ActionRemove, metaschema.PolarityShared},
		{"security.*", metaschema.ActionAdd, metaschema.PolarityDocument},
		{"webhooks.*.*.requestBody.content.*.schema.pattern", metaschema.ActionChange, metaschema.PolarityRequest},
	} {
		c, ok := find(set, tc.location, tc.action)
		if !ok {
			t.Errorf("missing edit: %s %s", tc.location, tc.action)
			continue
		}
		if c.Polarity != tc.polarity {
			t.Errorf("%s %s: polarity %s, want %s", tc.location, tc.action, c.Polarity, tc.polarity)
		}
	}
}

func TestCube_RecursionFolds(t *testing.T) {
	set := editSet(t)

	// schema keywords appear at the top schema node only; deeper nesting is
	// folded into it
	if _, ok := find(set, "paths.*.*.requestBody.content.*.schema.properties.*.maxLength", metaschema.ActionDecrease); ok {
		t.Error("schema recursion not folded: found a property-level maxLength edit")
	}
	// callbacks fold at the nested path item
	if _, ok := find(set, "paths.*.*.callbacks.*.*", metaschema.ActionAdd); !ok {
		t.Error("missing callback path-item edit")
	}
	if _, ok := find(set, "paths.*.*.callbacks.*.*.*.requestBody", metaschema.ActionSet); ok {
		t.Error("path-item recursion not folded under callbacks")
	}
}

func TestCube_AnnotationAndExtension(t *testing.T) {
	set := editSet(t)

	c, ok := find(set, "paths.*.*.description", metaschema.ActionChange)
	if !ok || !c.Annotation {
		t.Errorf("operation description should be an annotation edit (found=%v edit=%+v)", ok, c)
	}
	c, ok = find(set, "paths.*.*.x-*", metaschema.ActionChange)
	if !ok || !c.Extension {
		t.Errorf("operation x-* should be an extension edit (found=%v edit=%+v)", ok, c)
	}
	c, ok = find(set, "paths.*.*.requestBody.content.*.schema.maxLength", metaschema.ActionDecrease)
	if !ok || c.Annotation {
		t.Errorf("maxLength should not be an annotation edit (found=%v edit=%+v)", ok, c)
	}
}

func TestMatchLocation(t *testing.T) {
	for _, tc := range []struct {
		pattern  string
		location string
		want     bool
	}{
		{"paths.*.*", "paths.*.*", true},
		{"paths.*.*.parameters.*.schema.maxLength", "paths.*.*.parameters.*.schema.maxLength", true},
		{"**.schema.maxLength", "paths.*.*.parameters.*.schema.maxLength", true},
		{"**.schema.maxLength", "webhooks.*.*.requestBody.content.*.schema.maxLength", true},
		{"paths.**.schema.maxLength", "paths.*.*.parameters.*.schema.maxLength", true},
		{"paths.**", "paths.*", true},
		{"**", "anything.at.all", true},
		{"paths.*.*", "paths.*", false},
		{"paths.*", "paths.*.*", false},
		{"**.maxLength", "paths.*.*.parameters.*.schema.minLength", false},
		{"components.**", "paths.*.*", false},
	} {
		if got := metaschema.MatchLocation(tc.pattern, tc.location); got != tc.want {
			t.Errorf("MatchLocation(%q, %q) = %v, want %v", tc.pattern, tc.location, got, tc.want)
		}
	}
}

// Every NonContract entry must match at least one enumerated edit that is
// not already excluded as an annotation or extension; an entry that matches
// nothing is stale and must be removed.
func TestNonContracts_NotStale(t *testing.T) {
	edits := metaschema.Edits()
	for _, nc := range metaschema.NonContracts {
		found := false
		for _, edit := range edits {
			if edit.Annotation || edit.Extension {
				continue
			}
			matches, err := metaschema.MatchPattern(nc.Pattern, edit)
			if err != nil {
				t.Fatalf("invalid NonContract pattern %q: %v", nc.Pattern, err)
			}
			if matches {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("stale NonContract entry: %q matches no edit", nc.Pattern)
		}
	}
}
