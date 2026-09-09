package diff

type state struct {
	visitedSchemasBase     map[string]struct{}
	visitedSchemasRevision map[string]struct{}
	cache                  schemaDiffCache
	inFlight               map[schemaPair]struct{}

	// when a cycle is detected, the cut count is incremented so that diffs
	// computed above the cut are not cached.
	cuts int
}

func newState() *state {
	return &state{
		visitedSchemasBase:     map[string]struct{}{},
		visitedSchemasRevision: map[string]struct{}{},
		cache:                  schemaDiffCache{},
		inFlight:               map[schemaPair]struct{}{},
	}
}
