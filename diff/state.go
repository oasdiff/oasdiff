package diff

type direction int

const (
	directionRequest direction = iota
	directionResponse
)

type state struct {
	visitedSchemasBase     map[string]struct{}
	visitedSchemasRevision map[string]struct{}
	cache                  directionalSchemaDiffCache
	direction              direction
}

func newState() *state {
	return &state{
		visitedSchemasBase:     map[string]struct{}{},
		visitedSchemasRevision: map[string]struct{}{},
		cache:                  newDirectionalSchemaDiffCache(),
		direction:              directionRequest,
	}
}

func (state *state) setDirection(direction direction) {
	state.direction = direction
}
