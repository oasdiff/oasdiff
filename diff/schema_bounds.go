package diff

// SchemaBound is one ordered validation keyword of a schema: where its diff
// lives and how that diff encodes absence. The pointer-backed fields report
// an absent keyword as nil; the plain-uint64 fields (minLength, minItems,
// minProperties) report it as 0, since a zero lower bound requires nothing;
// the exclusive bounds report false as absent, since exclusiveMinimum: false
// (the OpenAPI 3.0 boolean form) declares the bound not exclusive, the same
// contract as leaving it out.
type SchemaBound struct {
	Keyword  string // the schema field name, e.g. "maxLength"
	Diff     func(*SchemaDiff) *ValueDiff
	encoding boundEncoding
}

// boundEncoding is how a bound's diff represents the absent keyword.
type boundEncoding int

const (
	nilIsAbsent   boundEncoding = iota
	zeroIsAbsent                // a plain uint64 whose zero requires nothing
	falseIsAbsent               // an ExclusiveBound whose false form requires nothing
)

// absent reports whether the diffed value represents the absent keyword.
func (b SchemaBound) absent(v any) bool {
	switch b.encoding {
	case zeroIsAbsent:
		return v == nil || v == uint64(0)
	case falseIsAbsent:
		return v == nil || v == false
	}
	return v == nil
}

// WasSet returns the value the keyword was set to, when it went from absent to
// present.
func (b SchemaBound) WasSet(d *SchemaDiff) (any, bool) {
	if b.Diff == nil {
		return nil, false
	}
	vd := b.Diff(d)
	if vd == nil || !b.absent(vd.From) || b.absent(vd.To) {
		return nil, false
	}
	return vd.To, true
}

// WasUnset returns the value the keyword was unset from, when it went from
// present to absent.
func (b SchemaBound) WasUnset(d *SchemaDiff) (any, bool) {
	if b.Diff == nil {
		return nil, false
	}
	vd := b.Diff(d)
	if vd == nil || b.absent(vd.From) || !b.absent(vd.To) {
		return nil, false
	}
	return vd.From, true
}

// WasIncreased returns the from and to values when the keyword was present on
// both sides and its value increased.
func (b SchemaBound) WasIncreased(d *SchemaDiff) (any, any, bool) {
	return b.ordered(d, lessValue)
}

// WasDecreased returns the from and to values when the keyword was present on
// both sides and its value decreased.
func (b SchemaBound) WasDecreased(d *SchemaDiff) (any, any, bool) {
	return b.ordered(d, func(a, b any) bool { return lessValue(b, a) })
}

func (b SchemaBound) ordered(d *SchemaDiff, less func(a, b any) bool) (any, any, bool) {
	if b.Diff == nil {
		return nil, nil, false
	}
	vd := b.Diff(d)
	if vd == nil || b.absent(vd.From) || b.absent(vd.To) || !less(vd.From, vd.To) {
		return nil, nil, false
	}
	return vd.From, vd.To, true
}

// lessValue reports whether a and b are ordered values of the same type with
// a < b. The exclusive bounds mix types (a bool form against a numeric form);
// such a pair is not ordered.
func lessValue(a, b any) bool {
	if au, ok := a.(uint64); ok {
		bu, ok := b.(uint64)
		return ok && au < bu
	}
	if af, ok := a.(float64); ok {
		bf, ok := b.(float64)
		return ok && af < bf
	}
	return false
}

// SchemaBounds lists every ordered validation keyword the schema diff
// compares. TestSchemaBounds pins each row to its getter's absence encoding.
var SchemaBounds = []SchemaBound{
	{"maximum", func(d *SchemaDiff) *ValueDiff { return d.MaxDiff }, nilIsAbsent},
	{"minimum", func(d *SchemaDiff) *ValueDiff { return d.MinDiff }, nilIsAbsent},
	{"multipleOf", func(d *SchemaDiff) *ValueDiff { return d.MultipleOfDiff }, nilIsAbsent},
	{"maxLength", func(d *SchemaDiff) *ValueDiff { return d.MaxLengthDiff }, nilIsAbsent},
	{"minLength", func(d *SchemaDiff) *ValueDiff { return d.MinLengthDiff }, zeroIsAbsent},
	{"maxItems", func(d *SchemaDiff) *ValueDiff { return d.MaxItemsDiff }, nilIsAbsent},
	{"minItems", func(d *SchemaDiff) *ValueDiff { return d.MinItemsDiff }, zeroIsAbsent},
	{"maxProperties", func(d *SchemaDiff) *ValueDiff { return d.MaxPropsDiff }, nilIsAbsent},
	{"minProperties", func(d *SchemaDiff) *ValueDiff { return d.MinPropsDiff }, zeroIsAbsent},
	{"minContains", func(d *SchemaDiff) *ValueDiff { return d.MinContainsDiff }, nilIsAbsent},
	{"maxContains", func(d *SchemaDiff) *ValueDiff { return d.MaxContainsDiff }, nilIsAbsent},
	{"exclusiveMinimum", func(d *SchemaDiff) *ValueDiff { return d.ExclusiveMinDiff }, falseIsAbsent},
	{"exclusiveMaximum", func(d *SchemaDiff) *ValueDiff { return d.ExclusiveMaxDiff }, falseIsAbsent},
}
