package checker_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/metaschema"
)

// This file audits the rule registry (GetAllRules) for broken symmetries: a
// coordinate populated on one side of a symmetry axis but empty on the
// mirror. Coordinates are built from each rule's declared Direction, Area and
// Kind plus the syntactic actions derived from its location claims; the
// semantic axis uses the rule's declared Effect. Each absence is either a
// real missing check or an intentional asymmetry (usually request/response
// contravariance).
//
// TestRuleSymmetry is the guard: every absence must be listed in
// symmetryWaivers with a reason, otherwise the test fails. A new rule that
// breaks symmetry, or a waiver that no longer applies, both fail the build,
// so the waiver list stays an honest, reviewed record of every intentional
// asymmetry.

// symmetryWaivers records every intentional asymmetry. Key is the canonical
// absence string emitted by symmetryAbsences; value is why it is acceptable.
// Removing a real check, or adding one that fills a gap, must update this map.
var symmetryWaivers = map[string]string{
	"add<->remove request/schema/lifecycle missing-remove":  "deprecation rules claim x-* add/change (sunset annotations appearing or changing); deleting a property-level sunset has no check yet, unlike the operation-level sunset-deleted.",
	"add<->remove response/schema/lifecycle missing-remove": "same as the request side: property-level sunset deletion is unchecked.",
}

// symmetryAbsences returns the canonical key for every coordinate that is
// populated on one side of a symmetry axis but completely empty on the mirror.
func symmetryAbsences(rules checker.BackwardCompatibilityRules) []string {
	var out []string

	// a coordinate of the taxonomy; group is a coordinate without its
	// action, the scope within which action duals and effects are compared;
	// mirror is one without its direction, compared across the two
	type coord struct {
		Direction checker.Direction
		Area      checker.Area
		Kind      checker.Kind
		Action    metaschema.Action
	}
	type group struct {
		Direction checker.Direction
		Area      checker.Area
		Kind      checker.Kind
	}
	type mirror struct {
		Area   checker.Area
		Kind   checker.Kind
		Action metaschema.Action
	}

	// Axis 1: request <-> response, restricted to Areas that appear in both
	// directions (parameters/request-body are request-only, responses/headers
	// response-only, so their missing mirror is structural, not a gap).
	reqAreas, respAreas := map[checker.Area]bool{}, map[checker.Area]bool{}
	for _, r := range rules {
		switch r.Direction {
		case checker.DirectionRequest:
			reqAreas[r.Area] = true
		case checker.DirectionResponse:
			respAreas[r.Area] = true
		}
	}
	req, resp := map[mirror]bool{}, map[mirror]bool{}
	present := map[coord]bool{}
	effects := map[group]map[checker.Effect]bool{}
	for _, r := range rules {
		for _, action := range r.Actions() {
			if reqAreas[r.Area] && respAreas[r.Area] {
				k := mirror{r.Area, r.Kind, action}
				switch r.Direction {
				case checker.DirectionRequest:
					req[k] = true
				case checker.DirectionResponse:
					resp[k] = true
				}
			}
			present[coord{r.Direction, r.Area, r.Kind, action}] = true
		}
		e := group{r.Direction, r.Area, r.Kind}
		if effects[e] == nil {
			effects[e] = map[checker.Effect]bool{}
		}
		effects[e][r.Effect] = true
	}
	for k := range req {
		if !resp[k] {
			out = append(out, fmt.Sprintf("request<->response %s/%s/%s missing-response", k.Area.String(), k.Kind.String(), k.Action))
		}
	}
	for k := range resp {
		if !req[k] {
			out = append(out, fmt.Sprintf("request<->response %s/%s/%s missing-request", k.Area.String(), k.Kind.String(), k.Action))
		}
	}

	// Axis 2: dual action pairs within the same Direction/Area/Kind.
	pairs := [][2]metaschema.Action{
		{metaschema.ActionAdd, metaschema.ActionRemove},
		{metaschema.ActionIncrease, metaschema.ActionDecrease},
		{metaschema.ActionSet, metaschema.ActionUnset},
	}
	for _, p := range pairs {
		coords := map[group]bool{}
		for k := range present {
			if k.Action == p[0] || k.Action == p[1] {
				coords[group{k.Direction, k.Area, k.Kind}] = true
			}
		}
		for c := range coords {
			has0 := present[coord{c.Direction, c.Area, c.Kind, p[0]}]
			has1 := present[coord{c.Direction, c.Area, c.Kind, p[1]}]
			coord := fmt.Sprintf("%s/%s/%s", c.Direction.String(), c.Area.String(), c.Kind.String())
			if has0 && !has1 {
				out = append(out, fmt.Sprintf("%s<->%s %s missing-%s", p[0], p[1], coord, p[1]))
			} else if has1 && !has0 {
				out = append(out, fmt.Sprintf("%s<->%s %s missing-%s", p[0], p[1], coord, p[0]))
			}
		}
	}

	// Axis 3: effect duality within the same Direction/Area/Kind. Where a
	// narrowing verdict exists, its widening counterpart should exist too
	// (usually as the safe-direction changelog entry), and vice versa.
	for e, effs := range effects {
		coord := fmt.Sprintf("%s/%s/%s", e.Direction.String(), e.Area.String(), e.Kind.String())
		if effs[checker.EffectNarrows] && !effs[checker.EffectWidens] {
			out = append(out, fmt.Sprintf("widens<->narrows %s missing-widens", coord))
		} else if effs[checker.EffectWidens] && !effs[checker.EffectNarrows] {
			out = append(out, fmt.Sprintf("widens<->narrows %s missing-narrows", coord))
		}
	}

	sort.Strings(out)
	return out
}

func TestRuleSymmetry(t *testing.T) {
	absences := symmetryAbsences(checker.GetAllRules())

	absent := map[string]bool{}
	for _, a := range absences {
		absent[a] = true
		if _, ok := symmetryWaivers[a]; !ok {
			t.Errorf("unwaived rule asymmetry: %q\n  fix it by adding the mirror rule, or document it in symmetryWaivers with a reason", a)
		}
	}
	for w := range symmetryWaivers {
		if !absent[w] {
			t.Errorf("stale symmetry waiver: %q\n  this asymmetry no longer exists; remove the waiver", w)
		}
	}
}
