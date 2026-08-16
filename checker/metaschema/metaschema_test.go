package metaschema_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker/metaschema"
)

func cellSet(t *testing.T) map[metaschema.Cell]bool {
	t.Helper()
	set := map[metaschema.Cell]bool{}
	for _, c := range metaschema.Cube() {
		if set[c] {
			t.Errorf("duplicate cell: %+v", c)
		}
		set[c] = true
	}
	return set
}

func find(set map[metaschema.Cell]bool, location string, action metaschema.Action) (metaschema.Cell, bool) {
	for c := range set {
		if c.Location == location && c.Action == action {
			return c, true
		}
	}
	return metaschema.Cell{}, false
}

func TestCube_Deterministic(t *testing.T) {
	a, b := metaschema.Cube(), metaschema.Cube()
	if len(a) != len(b) {
		t.Fatalf("cube size changed between runs: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("cube order changed between runs at %d: %+v vs %+v", i, a[i], b[i])
		}
	}
}

func TestCube_KnownCells(t *testing.T) {
	set := cellSet(t)

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
			t.Errorf("missing cell: %s %s", tc.location, tc.action)
			continue
		}
		if c.Polarity != tc.polarity {
			t.Errorf("%s %s: polarity %s, want %s", tc.location, tc.action, c.Polarity, tc.polarity)
		}
	}
}

func TestCube_RecursionFolds(t *testing.T) {
	set := cellSet(t)

	// schema keywords appear at the top schema node only; deeper nesting is
	// folded into it
	if _, ok := find(set, "paths.*.*.requestBody.content.*.schema.properties.*.maxLength", metaschema.ActionDecrease); ok {
		t.Error("schema recursion not folded: found a property-level maxLength cell")
	}
	// callbacks fold at the nested path item
	if _, ok := find(set, "paths.*.*.callbacks.*.*", metaschema.ActionAdd); !ok {
		t.Error("missing callback path-item cell")
	}
	if _, ok := find(set, "paths.*.*.callbacks.*.*.*.requestBody", metaschema.ActionSet); ok {
		t.Error("path-item recursion not folded under callbacks")
	}
}

func TestCube_AnnotationAndExtension(t *testing.T) {
	set := cellSet(t)

	c, ok := find(set, "paths.*.*.description", metaschema.ActionChange)
	if !ok || !c.Annotation {
		t.Errorf("operation description should be an annotation cell (found=%v cell=%+v)", ok, c)
	}
	c, ok = find(set, "paths.*.*.x-*", metaschema.ActionChange)
	if !ok || !c.Extension {
		t.Errorf("operation x-* should be an extension cell (found=%v cell=%+v)", ok, c)
	}
	c, ok = find(set, "paths.*.*.requestBody.content.*.schema.maxLength", metaschema.ActionDecrease)
	if !ok || c.Annotation {
		t.Errorf("maxLength should not be an annotation cell (found=%v cell=%+v)", ok, c)
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
