package checker_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
)

// This file audits the rule registry (GetAllRules) for broken symmetries on
// the Direction/Area/Kind/Action taxonomy: a coordinate populated on one side
// of a symmetry axis but empty on the mirror. Each such absence is either a
// real missing check or an intentional asymmetry (the mirror is not a
// wire-contract change, e.g. by request/response contravariance).
//
// TestRuleSymmetry is the guard: every absence must be listed in symmetryWaivers
// with a reason, otherwise the test fails. A new rule that breaks symmetry, or a
// waiver that no longer applies, both fail the build, so the waiver list stays an
// honest, reviewed record of every intentional asymmetry.
//
// TestRuleSymmetryReport is informational (count mismatches that contravariance
// makes legitimate); run with: go test ./checker -run RuleSymmetryReport -v

// symmetryWaivers records every intentional asymmetry. Key is the canonical
// absence string emitted by symmetryAbsences; value is why it is acceptable.
// Removing a real check, or adding one that fills a gap, must update this map.
var symmetryWaivers = map[string]string{
	// --- request <-> response (bidirectional areas only) ---
	"request<->response schema/type/generalize missing-response":        "response type generalization is the breaking direction (a wider returned type can break clients); it is reported by response-property-type-changed (ERR). The safe response direction is specialization. Granularity asymmetry tracked in #1034.",
	"request<->response schema/constraints/generalize missing-response": "response pattern generalization is breaking and reported by response-property-pattern-changed; the safe response direction is specialization. Tracked in #1034.",
	"request<->response schema/constraints/set missing-response":        "setting a constraint on a response narrows the server's output, which is non-breaking for clients, so it is request-only by contravariance (request reports the constraint as newly enforced).",

	// --- add <-> remove (same direction/area/kind) ---
	"add<->remove none/paths/lifecycle missing-add": "a sunset date being added is part of the deprecation flow and is reported by the deprecation rules; only its removal (sunset-deleted) is a standalone change.",

	// --- generalize <-> specialize (same direction/area/kind) ---
	"generalize<->specialize request/schema/type missing-specialize":            "narrowing a request type is breaking and already reported by request-*-type-changed (ERR); the generalize rule exists only to carve out the safe widening as INFO.",
	"generalize<->specialize request/schema/constraints missing-specialize":     "tightening a request pattern is reported by request-*-pattern-changed; the generalize rule carves out the safe loosening.",
	"generalize<->specialize request/parameters/constraints missing-specialize": "same as request schema pattern: tightening is covered by request-parameter-pattern-changed; generalize carves out the safe loosening.",

	// --- KNOWN GAP, tracked (not intentional) ---
	"add<->remove response/headers/existence missing-add": "GAP tracked in #1033: adding a response header has no rule, only removal. Remove this waiver when response-header-added lands.",
}

// symmetryAbsences returns the canonical key for every coordinate that is
// populated on one side of a symmetry axis but completely empty on the mirror.
// Count mismatches (both sides populated) are intentionally excluded: by
// request/response contravariance the counts legitimately differ.
func symmetryAbsences(rules checker.BackwardCompatibilityRules) []string {
	var out []string

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
	type aka struct {
		Area   checker.Area
		Kind   checker.Kind
		Action checker.Action
	}
	req, resp := map[aka]bool{}, map[aka]bool{}
	for _, r := range rules {
		if !reqAreas[r.Area] || !respAreas[r.Area] {
			continue
		}
		k := aka{r.Area, r.Kind, r.Action}
		switch r.Direction {
		case checker.DirectionRequest:
			req[k] = true
		case checker.DirectionResponse:
			resp[k] = true
		}
	}
	for k := range req {
		if !resp[k] {
			out = append(out, fmt.Sprintf("request<->response %s/%s/%s missing-response", k.Area.String(), k.Kind.String(), k.Action.String()))
		}
	}
	for k := range resp {
		if !req[k] {
			out = append(out, fmt.Sprintf("request<->response %s/%s/%s missing-request", k.Area.String(), k.Kind.String(), k.Action.String()))
		}
	}

	// Axis 2: dual action pairs within the same Direction/Area/Kind.
	pairs := [][2]checker.Action{
		{checker.ActionAdd, checker.ActionRemove},
		{checker.ActionIncrease, checker.ActionDecrease},
		{checker.ActionGeneralize, checker.ActionSpecialize},
	}
	type dak struct {
		Direction checker.Direction
		Area      checker.Area
		Kind      checker.Kind
		Action    checker.Action
	}
	present := map[dak]bool{}
	for _, r := range rules {
		present[dak{r.Direction, r.Area, r.Kind, r.Action}] = true
	}
	for _, p := range pairs {
		coords := map[dak]bool{}
		for k := range present {
			if k.Action == p[0] || k.Action == p[1] {
				coords[dak{k.Direction, k.Area, k.Kind, 0}] = true
			}
		}
		for c := range coords {
			has0 := present[dak{c.Direction, c.Area, c.Kind, p[0]}]
			has1 := present[dak{c.Direction, c.Area, c.Kind, p[1]}]
			coord := fmt.Sprintf("%s/%s/%s", c.Direction.String(), c.Area.String(), c.Kind.String())
			if has0 && !has1 {
				out = append(out, fmt.Sprintf("%s<->%s %s missing-%s", p[0].String(), p[1].String(), coord, p[1].String()))
			} else if has1 && !has0 {
				out = append(out, fmt.Sprintf("%s<->%s %s missing-%s", p[0].String(), p[1].String(), coord, p[0].String()))
			}
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

// --- informational report (not asserted) ---

func TestRuleSymmetryReport(t *testing.T) {
	rules := checker.GetAllRules()
	t.Logf("total rules: %d", len(rules))

	reqAreas, respAreas := map[checker.Area]bool{}, map[checker.Area]bool{}
	for _, r := range rules {
		switch r.Direction {
		case checker.DirectionRequest:
			reqAreas[r.Area] = true
		case checker.DirectionResponse:
			respAreas[r.Area] = true
		}
	}
	type key struct {
		Area   checker.Area
		Kind   checker.Kind
		Action checker.Action
	}
	req, resp := map[key][]string{}, map[key][]string{}
	for _, r := range rules {
		if !reqAreas[r.Area] || !respAreas[r.Area] {
			continue
		}
		k := key{r.Area, r.Kind, r.Action}
		switch r.Direction {
		case checker.DirectionRequest:
			req[k] = append(req[k], r.Id)
		case checker.DirectionResponse:
			resp[k] = append(resp[k], r.Id)
		}
	}
	seen := map[key]bool{}
	var keys []key
	for k := range req {
		keys, seen[k] = append(keys, k), true
	}
	for k := range resp {
		if !seen[k] {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if a.Area != b.Area {
			return a.Area < b.Area
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Action < b.Action
	})
	t.Log("=== request vs response count mismatches (contravariance expected) ===")
	for _, k := range keys {
		if len(req[k]) > 0 && len(resp[k]) > 0 && len(req[k]) != len(resp[k]) {
			t.Logf("  %d req / %d resp [%s/%s/%s]", len(req[k]), len(resp[k]), k.Area.String(), k.Kind.String(), k.Action.String())
		}
	}
}
