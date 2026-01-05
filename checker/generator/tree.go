package generator

import (
	"fmt"
	"os"
	"strings"

	"go.yaml.in/yaml/v3"
)

type ChangeTree struct {
	Changes    ChangeMap `yaml:"changes"`
	Components ChangeMap `yaml:"components"`
}

type ChangeMap map[string]Changes

type Changes struct {
	Ref                  string    `yaml:"$ref"`
	ExcludeFromHierarchy bool      `yaml:"excludeFromHierarchy"`
	Actions              Actions   `yaml:"actions"`
	NextLevel            ChangeMap `yaml:"nextLevel"`
}

type Actions map[string]Objects
type Objects []Object

type Object struct {
	Hierarchy            []string `yaml:"hierarchy"`
	Names                []string `yaml:"names"`
	Adverbs              []string `yaml:"adverbs"`
	StartWithName        bool     `yaml:"startWithName"`
	PredicativeAdjective string   `yaml:"predicativeAdjective"`
	AttributiveAdjective string   `yaml:"attributiveAdjective"`
}

func GetTree(file string) func() (MessageGenerator, error) {
	return func() (MessageGenerator, error) {
		yamlFile, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("yamlFile.Get err   #%v ", err)
		}

		var changeMap ChangeTree
		err = yaml.Unmarshal(yamlFile, &changeMap)
		if err != nil {
			return nil, fmt.Errorf("unmarshal: %v", err)
		}

		return changeMap, nil
	}
}

func (changeTree ChangeTree) generate() []string {
	resolveRefs(changeTree.Changes, changeTree.Components)
	fillHierarchy(changeTree.Changes, nil)
	return generateRecursive(changeTree.Changes)
}

func (changeMap ChangeMap) copy() ChangeMap {
	result := ChangeMap{}
	for key, value := range changeMap {
		result[key] = value.copy()
	}
	return result
}

func (changes Changes) copy() Changes {
	return Changes{
		Ref:                  changes.Ref,
		ExcludeFromHierarchy: changes.ExcludeFromHierarchy,
		Actions:              changes.Actions.copy(),
		NextLevel:            changes.NextLevel.copy(),
	}
}

func (actions Actions) copy() Actions {
	result := Actions{}
	for key, value := range actions {
		result[key] = value.copy()
	}
	return result
}

func (objects Objects) copy() Objects {
	result := make(Objects, 0, len(objects))
	return append(result, objects...)
}

func resolveRefs(changes ChangeMap, components ChangeMap) {
	for container, change := range changes {
		if change.Ref != "" {
			changes[container] = components[change.Ref].copy()
		}
		resolveRefs(changes[container].NextLevel, components)
	}
}

func generateRecursive(changes ChangeMap) []string {
	result := []string{}

	for _, change := range changes {
		for action, objects := range change.Actions {
			for _, object := range objects {
				result = append(result, getValueSet(object, action).generate()...)
			}
		}
		result = append(result, generateRecursive(change.NextLevel)...)
	}

	return result
}

func fillHierarchy(changes ChangeMap, hierarchy []string) {
	for container, change := range changes {
		containerHierarchy := getContainerHierarchy(container, change, hierarchy)
		for action, objects := range change.Actions {
			for i := range objects {
				changes[container].Actions[action][i].Hierarchy = containerHierarchy
			}
		}
		fillHierarchy(change.NextLevel, containerHierarchy)
	}
}

func getContainerHierarchy(container string, change Changes, hierarchy []string) []string {
	if change.ExcludeFromHierarchy {
		return hierarchy
	}
	return append([]string{container}, hierarchy...)
}

func getValueSet(object Object, action string) IValueSet {
	valueSet := ValueSet{
		AttributiveAdjective: object.AttributiveAdjective,
		PredicativeAdjective: object.PredicativeAdjective,
		Hierarchy:            object.Hierarchy,
		Names:                object.Names,
		Actions:              parseAction(action),
		Adverbs:              object.Adverbs,
	}

	if object.StartWithName {
		return ValueSetA(valueSet)
	}
	return ValueSetB(valueSet)
}

func parseAction(action string) []string {
	return strings.Split(action, "/")
}
