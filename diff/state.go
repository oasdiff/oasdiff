package diff

type state struct {
	graph schemaGraph
}

func newState() *state {
	return &state{graph: newSchemaGraph()}
}
