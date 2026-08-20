package metaschema

// Action is a syntactic edit applicable at a location: what can literally
// happen to the field in the document, independent of whether it breaks
// clients.
type Action string

const (
	ActionAdd      Action = "add"      // add a map entry, list element, or set member
	ActionRemove   Action = "remove"   // remove a map entry, list element, or set member
	ActionSet      Action = "set"      // the field appears where it was absent
	ActionUnset    Action = "unset"    // the field disappears
	ActionChange   Action = "change"   // the field's value changes to another value
	ActionIncrease Action = "increase" // an ordered value grows
	ActionDecrease Action = "decrease" // an ordered value shrinks
)
