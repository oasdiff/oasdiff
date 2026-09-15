package diff

import "math"

// unroll returns the tree of changes below a node of the schema diff graph,
// nil when there are none. A node reached again on its own path is a cycle
// and contributes nothing there: every change below it is already reported
// where it was first reached. A node whose diff is still in progress
// contributes nothing either, since its fields are not filled yet.
func (state *state) unroll(node *SchemaDiff) *SchemaDiff {
	u := unroller{state: state, path: map[*SchemaDiff]int{}, minCut: math.MaxInt}
	return u.unroll(node)
}

type unroller struct {
	state *state

	// path maps each node on the current unrolling path to its depth.
	path map[*SchemaDiff]int

	// minCut is the shallowest depth a cut has targeted since the current
	// node began; -1 for a node in progress, math.MaxInt while none has.
	minCut int
}

func (u *unroller) unroll(node *SchemaDiff) *SchemaDiff {
	if node == nil {
		return nil
	}
	if _, ok := u.state.inProgress[node]; ok {
		u.minCut = -1
		return nil
	}
	if depth, ok := u.path[node]; ok {
		u.minCut = min(u.minCut, depth)
		return nil
	}
	if tree, ok := u.state.unrolled[node]; ok {
		return tree
	}

	depth := len(u.path)
	u.path[node] = depth
	defer delete(u.path, node)

	outerMinCut := u.minCut
	u.minCut = math.MaxInt
	tree := u.copy(node)
	// A cut into this node itself falls at the same place on every path, so
	// the tree is a function of the node alone; a cut into a node above it
	// makes the tree depend on the path.
	if u.minCut >= depth {
		u.state.unrolled[node] = tree
	}
	u.minCut = min(outerMinCut, u.minCut)
	return tree
}

func (u *unroller) copy(node *SchemaDiff) *SchemaDiff {
	tree := *node
	tree.OneOfDiff = u.subschemas(node.OneOfDiff)
	tree.AnyOfDiff = u.subschemas(node.AnyOfDiff)
	tree.AllOfDiff = u.subschemas(node.AllOfDiff)
	tree.NotDiff = u.unroll(node.NotDiff)
	tree.ItemsDiff = u.unroll(node.ItemsDiff)
	tree.PropertiesDiff = u.schemas(node.PropertiesDiff)
	tree.AdditionalPropertiesDiff = u.unroll(node.AdditionalPropertiesDiff)
	tree.PrefixItemsDiff = u.subschemas(node.PrefixItemsDiff)
	tree.ContainsDiff = u.unroll(node.ContainsDiff)
	tree.PatternPropertiesDiff = u.schemas(node.PatternPropertiesDiff)
	tree.DependentSchemasDiff = u.schemas(node.DependentSchemasDiff)
	tree.PropertyNamesDiff = u.unroll(node.PropertyNamesDiff)
	tree.UnevaluatedItemsDiff = u.unroll(node.UnevaluatedItemsDiff)
	tree.UnevaluatedPropertiesDiff = u.unroll(node.UnevaluatedPropertiesDiff)
	tree.IfDiff = u.unroll(node.IfDiff)
	tree.ThenDiff = u.unroll(node.ThenDiff)
	tree.ElseDiff = u.unroll(node.ElseDiff)
	tree.ContentSchemaDiff = u.unroll(node.ContentSchemaDiff)
	tree.DefsDiff = u.schemas(node.DefsDiff)
	if tree.Empty() {
		return nil
	}
	return &tree
}

func (u *unroller) schemas(diff *SchemasDiff) *SchemasDiff {
	if diff == nil {
		return nil
	}
	tree := *diff
	tree.Modified = ModifiedSchemasMap{}
	for name, node := range diff.Modified {
		if child := u.unroll(node); child != nil {
			tree.Modified[name] = child
		}
	}
	if tree.Empty() {
		return nil
	}
	return &tree
}

func (u *unroller) subschemas(diff *SubschemasDiff) *SubschemasDiff {
	if diff == nil {
		return nil
	}
	tree := *diff
	tree.Modified = ModifiedSubschemas{}
	for _, modified := range diff.Modified {
		if child := u.unroll(modified.Diff); child != nil {
			tree.Modified = append(tree.Modified, &ModifiedSubschema{Base: modified.Base, Revision: modified.Revision, Diff: child})
		}
	}
	if tree.Empty() {
		return nil
	}
	return &tree
}
