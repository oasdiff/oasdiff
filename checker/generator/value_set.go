package generator

import (
	"fmt"
	"io"
	"strings"
)

type ValueSets []IValueSet

// NewValueSets creates a new ValueSets object
func NewValueSets(hierarchy []string, valueSets ValueSets) ValueSets {

	result := make(ValueSets, len(valueSets))

	for i, vs := range valueSets {
		result[i] = vs.setHierarchy(hierarchy)
	}

	return result
}

func (vs ValueSets) generate(out io.Writer) {
	for _, v := range vs {
		v.generate(out)
	}
}

type IValueSet interface {
	generate(out io.Writer)
	setHierarchy(hierarchy []string) IValueSet
}

type ValueSet struct {
	attributiveAdjective string // attributive adjectives are added before the noun
	predicativeAdjective string // predicative adjectives are added after the noun
	hierarchy            []string
	objects              []string
	actions              []string
	adverb               []string
}

func (v ValueSet) setHierarchy(hierarchy []string) ValueSet {
	if len(hierarchy) == 0 {
		return v
	}

	v.hierarchy = append(v.hierarchy, hierarchy...)

	return v
}

// ValueSetA messages start with the noun
type ValueSetA ValueSet

func (v ValueSetA) setHierarchy(hierarchy []string) IValueSet {
	return ValueSetA(ValueSet(v).setHierarchy(hierarchy))
}

func (v ValueSetA) generate(out io.Writer) {
	generateMessage := func(hierarchy []string, object, attributiveAdjective, predicativeAdjective, action, adverb string) string {
		prefix := addAttribute(object, attributiveAdjective, predicativeAdjective)
		if hierarchyMessage := getHierarchyMessage(hierarchy); hierarchyMessage != "" {
			prefix += " of " + hierarchyMessage
		}

		return standardizeSpaces(fmt.Sprintf("%s was %s %s %s", prefix, conjugate(action), getActionMessage(action), adverb))
	}

	for _, object := range v.objects {
		for _, action := range v.actions {
			id := generateId(v.hierarchy, object, action)

			adverbs := v.adverb
			if v.adverb == nil {
				adverbs = []string{""}
			}
			for _, adverb := range adverbs {
				message := generateMessage(v.hierarchy, object, v.attributiveAdjective, v.predicativeAdjective, action, adverb)
				fmt.Fprintf(out, "%s: %s\n", id, message)
			}
		}
	}
}

// ValueSetB messages start with the action
type ValueSetB ValueSet

func (v ValueSetB) setHierarchy(hierarchy []string) IValueSet {
	return ValueSetB(ValueSet(v).setHierarchy(hierarchy))
}

func (v ValueSetB) generate(out io.Writer) {
	generateMessage := func(hierarchy []string, noun, attributiveAdjective, predicativeAdjective, action string) string {
		return standardizeSpaces(strings.Join([]string{conjugate(action), addAttribute(noun, attributiveAdjective, predicativeAdjective), getHierarchyPostfix(action, hierarchy)}, " "))
	}

	for _, noun := range v.objects {
		for _, action := range v.actions {
			fmt.Fprintf(out, "%s: %s\n", generateId(v.hierarchy, noun, action), generateMessage(v.hierarchy, noun, v.attributiveAdjective, v.predicativeAdjective, action))
		}
	}
}
