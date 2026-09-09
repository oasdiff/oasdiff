package diff

type state struct {
	visitedSchemasBase     map[string]struct{}
	visitedSchemasRevision map[string]struct{}
	cache                  schemaDiffCache
	inFlight               map[schemaPair]struct{}
}

func newState() *state {
	return &state{
		visitedSchemasBase:     map[string]struct{}{},
		visitedSchemasRevision: map[string]struct{}{},
		cache:                  schemaDiffCache{},
		inFlight:               map[schemaPair]struct{}{},
	}
}
