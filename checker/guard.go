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
// with the guards observed at the change's location added to its rule's own:
// narrowing a read-only property cannot invalidate a request, so the change
// is reported at info rather than error. A level the caller set is not
// lowered. Guards that do not change the rule's derived level are dropped,
// so GetComment explains exactly the ones that did.
func capByGuards(config *Config, changes Changes) Changes {
	for i, change := range changes {
		apiChange, ok := change.(ApiChange)
		if !ok || len(apiChange.guards) == 0 {
			continue
		}

		applied := appliedGuards(apiChange, config)
		if !slices.Equal(applied.guards, apiChange.guards) || applied.level < apiChange.Level {
			apiChange.guards = applied.guards
			if applied.level < apiChange.Level {
				apiChange.Level = applied.level
				// the check's own comment explains the verdict the guard
				// replaced, so the guard comment speaks alone
				apiChange.Comment = ""
			}
			changes[i] = apiChange
		}
	}
	return changes
}

type guardVerdict struct {
	guards []Guard
	level  Level
}

// appliedGuards keeps the guards of the change that lower its rule's derived
// level, and the level the law derives with them.
func appliedGuards(apiChange ApiChange, config *Config) guardVerdict {
	rule, ok := ruleById()[apiChange.Id]
	if !ok || config.overriddenLevels[apiChange.Id] {
		return guardVerdict{guards: nil, level: apiChange.Level}
	}

	base := rule.DerivedLevel()
	var applied []Guard
	for _, g := range apiChange.guards {
		if rules.DeriveLevel(rule.Effect, rule.Direction, append(slices.Clone(rule.Guards), g)...) < base {
			applied = append(applied, g)
		}
	}
	if len(applied) == 0 {
		return guardVerdict{guards: nil, level: apiChange.Level}
	}
	return guardVerdict{
		guards: applied,
		level:  rules.DeriveLevel(rule.Effect, rule.Direction, append(slices.Clone(rule.Guards), applied...)...),
	}
}
