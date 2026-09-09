package diff

type state struct {
	visitedSchemasBase     map[string]struct{}
	visitedSchemasRevision map[string]struct{}
	cache                  schemaDiffCache
	inFlight               map[schemaPair]struct{}

	// cuts counts the cycle-guard cuts taken so far. A cut's verdict
	// depends on which ancestors are on the path, so a diff computed while
	// a cut fires beneath it holds for its own path only: getSchemaDiff
	// caches a pair's diff only when no cut fired during its computation.
	// Caching such a diff replays one path's answer at every other use of
	// the pair, and which path computes first follows the map iteration
	// order over properties, so the output differs between runs (#1230).
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
