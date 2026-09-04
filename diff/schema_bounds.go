package diff

// SchemaBound is one ordered validation keyword of a schema: where its diff
// lives and how that diff encodes absence. The pointer-backed fields report
// an absent keyword as nil; the plain-uint64 fields (minLength, minItems,
// minProperties) report it as 0, since a zero lower bound requires nothing.
type SchemaBound struct {
	Keyword      string // the schema field name, e.g. "maxLength"
	Diff         func(*SchemaDiff) *ValueDiff
	zeroIsAbsent bool
}

// WasSet returns the value the keyword was set to, when it went from absent to
// present.
func (b SchemaBound) WasSet(d *SchemaDiff) (any, bool) {
	vd := b.Diff(d)
	if vd == nil {
		return nil, false
	}
	if b.zeroIsAbsent {
		if vd.From == uint64(0) && vd.To != uint64(0) {
			return vd.To, true
		}
		return nil, false
	}
	if vd.From == nil && vd.To != nil {
		return vd.To, true
	}
	return nil, false
}

// WasUnset returns the value the keyword was unset from, when it went from
// present to absent.
func (b SchemaBound) WasUnset(d *SchemaDiff) (any, bool) {
	vd := b.Diff(d)
	if vd == nil {
		return nil, false
	}
	if b.zeroIsAbsent {
		if vd.From != uint64(0) && vd.To == uint64(0) {
			return vd.From, true
		}
		return nil, false
	}
	if vd.From != nil && vd.To == nil {
		return vd.From, true
	}
	return nil, false
}

// SchemaBounds lists every ordered validation keyword the schema diff
// compares. TestSchemaBounds pins each row to its getter's absence encoding.
var SchemaBounds = []SchemaBound{
	{"maximum", func(d *SchemaDiff) *ValueDiff { return d.MaxDiff }, false},
	{"minimum", func(d *SchemaDiff) *ValueDiff { return d.MinDiff }, false},
	{"multipleOf", func(d *SchemaDiff) *ValueDiff { return d.MultipleOfDiff }, false},
	{"maxLength", func(d *SchemaDiff) *ValueDiff { return d.MaxLengthDiff }, false},
	{"minLength", func(d *SchemaDiff) *ValueDiff { return d.MinLengthDiff }, true},
	{"maxItems", func(d *SchemaDiff) *ValueDiff { return d.MaxItemsDiff }, false},
	{"minItems", func(d *SchemaDiff) *ValueDiff { return d.MinItemsDiff }, true},
	{"maxProperties", func(d *SchemaDiff) *ValueDiff { return d.MaxPropsDiff }, false},
	{"minProperties", func(d *SchemaDiff) *ValueDiff { return d.MinPropsDiff }, true},
	{"minContains", func(d *SchemaDiff) *ValueDiff { return d.MinContainsDiff }, false},
	{"maxContains", func(d *SchemaDiff) *ValueDiff { return d.MaxContainsDiff }, false},
}
