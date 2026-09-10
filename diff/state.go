package diff

type state struct {
	cache    schemaDiffCache
	inFlight map[schemaPair]struct{}

	// when a cycle is detected, the cut count is incremented; a diff whose
	// computation included a cut is not cached.
	cuts int
}

func newState() *state {
	return &state{
		cache:    schemaDiffCache{},
		inFlight: map[schemaPair]struct{}{},
	}
}
