package checker

import (
	"slices"

	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/checker/rules"
	"github.com/oasdiff/oasdiff/diff"
)

// propertyGuards reports the document states of a property that qualify a
// verdict under the severity law: a readOnly property does not appear in
// requests, a writeOnly property does not appear in responses.
func propertyGuards(d *diff.SchemaDiff) []Guard {
	var guards []Guard
	if d.Revision.ReadOnly {
		guards = append(guards, GuardReadOnly)
	}
	if d.Revision.WriteOnly {
		guards = append(guards, GuardWriteOnly)
	}
	return guards
}

// levelByGuards sets each change's level to what the severity law derives
// with the guards observed at the change's location added to its rule's own.
// The guards that changed the level stay on the change and GetComment
// renders one explanatory sentence per guard; guards that changed nothing
// are removed so no comment claims an effect it did not have.
func levelByGuards(config *Config, changes Changes) Changes {
	for i, change := range changes {
		if apiChange, ok := change.(ApiChange); ok && len(apiChange.guards) > 0 {
			changes[i] = levelByGuard(config, apiChange)
		}
	}
	return changes
}

// levelByGuard applies GuardNegated first: it reverses the effect, so it can
// raise the level as well as lower it, and the other guards are judged on
// the reversed effect. Each of the others is then applied on its own and the
// lowest level any of them derives is taken; such a guard is a fact about the
// document that is sufficient by itself to nullify the effect, so they
// combine as a minimum. A level the caller set with --severity-levels is not
// touched.
func levelByGuard(config *Config, apiChange ApiChange) ApiChange {
	rule, ok := ruleById()[apiChange.Id]
	if !ok || config.overriddenLevels[apiChange.Id] {
		apiChange.guards = nil
		return apiChange
	}

	ruleGuards := rule.Guards
	base := rule.DerivedLevel()
	var applied []Guard
	if slices.Contains(apiChange.guards, GuardNegated) {
		ruleGuards = append(slices.Clone(ruleGuards), GuardNegated)
		if negated := rules.DeriveLevel(rule.Effect, rule.Direction, ruleGuards...); negated != base {
			base = negated
			applied = append(applied, GuardNegated)
			apiChange.Level = negated
			// the check's own comment explains the verdict the guard replaced
			apiChange.Comment = ""
		}
	}

	level := base
	for _, g := range apiChange.guards {
		if g == GuardNegated {
			continue
		}
		if derived := rules.DeriveLevel(rule.Effect, rule.Direction, append(slices.Clone(ruleGuards), g)...); derived < base {
			applied = append(applied, g)
			level = min(level, derived)
		}
	}

	apiChange.guards = applied
	if level < apiChange.Level {
		apiChange.Level = level
		// the check's own comment explains the verdict the guard replaced,
		// so the guard comment speaks alone
		apiChange.Comment = ""
	}
	return apiChange
}

// attachNegations adds GuardNegated to each change in a schema under an odd
// number of `not` keywords.
func attachNegations(changes Changes) Changes {
	negations := map[*diff.SchemaDiff]schemawalk.Negations{}
	for i, change := range changes {
		apiChange, ok := change.(ApiChange)
		if !ok || apiChange.root == nil || apiChange.propertyPath == "" {
			continue
		}
		negationsOfRoot, ok := negations[apiChange.root]
		if !ok {
			negationsOfRoot = schemawalk.NewNegations(apiChange.root)
			negations[apiChange.root] = negationsOfRoot
		}
		if negationsOfRoot.At(apiChange.propertyPath) {
			changes[i] = apiChange.WithGuards([]Guard{GuardNegated})
		}
	}
	return changes
}
