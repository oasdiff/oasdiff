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
	inFlight               map[inFlightPair]struct{}
	direction              direction
}

// inFlightPair identifies a schema pair that is currently being diffed,
// so a cycle that loses its $ref (e.g. after --flatten-allof) can be cut
// even though the ref-based circular guard can't see it.
type inFlightPair struct {
	direction direction
	pair      schemaPair
}

func newState() *state {
	return &state{
		visitedSchemasBase:     map[string]struct{}{},
		visitedSchemasRevision: map[string]struct{}{},
		cache:                  newDirectionalSchemaDiffCache(),
		inFlight:               map[inFlightPair]struct{}{},
		direction:              directionRequest,
	}
}

func (state *state) setDirection(direction direction) {
	state.direction = direction
}
