package checker

import (
	"slices"

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

// capByGuards lowers each change's level to what the severity law derives
// with the guards observed at the change's location added to its rule's own.
// The guards that lowered the level stay on the change and GetComment
// renders one explanatory sentence per guard; guards that changed nothing
// are removed so no comment claims an effect it did not have.
func capByGuards(config *Config, changes Changes) Changes {
	for i, change := range changes {
		if apiChange, ok := change.(ApiChange); ok && len(apiChange.guards) > 0 {
			changes[i] = capByGuard(config, apiChange)
		}
	}
	return changes
}

// capByGuard applies each observed guard on its own and takes the lowest
// level any of them derives; a guard is a fact about the document that is
// sufficient by itself to nullify the effect, so guards combine as a
// minimum. A level the caller set with --severity-levels is not touched.
func capByGuard(config *Config, apiChange ApiChange) ApiChange {
	rule, ok := ruleById()[apiChange.Id]
	if !ok || config.overriddenLevels[apiChange.Id] {
		apiChange.guards = nil
		return apiChange
	}

	base := rule.DerivedLevel()
	level := base
	var applied []Guard
	for _, g := range apiChange.guards {
		if derived := rules.DeriveLevel(rule.Effect, rule.Direction, append(slices.Clone(rule.Guards), g)...); derived < base {
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
