package validate

// Source locations for findings. kin-openapi's Origin model is the substrate:
// Origin.Key points at the start of the enclosing collection and
// Origin.Fields.Lookup(X) at a specific scalar field inside it, so pinning a finding
// to the offending line means preferring the field entry over the key. These
// helpers own that resolution for both the kin-error path (locationForKinError)
// and the native lints (schemaFieldLocation).

import (
	"errors"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// lineColumnForKinError returns the line and column of a typed kin error's
// Origin, resolving the location once for both. Zero when origin metadata
// isn't available (untyped error, doc-root field, or loader.IncludeOrigin
// = false).
func lineColumnForKinError(err error) (int, int) {
	if loc := locationForKinError(err); loc != nil {
		return loc.Line, loc.Column
	}
	return 0, 0
}

// locationForKinError returns the most-specific *Location available
// for a typed kin error. kin's Origin model:
//
//   - Origin.Key       points at the start of the enclosing collection
//     (e.g. for a LicenseIdentifierFieldFor31Plus, Key is the line of
//     the parent "license:" key, not "identifier:").
//   - Origin.Fields.Lookup(X) points at the specific scalar field X inside
//     that collection.
//
// The cases below are refinements: each names the field its cluster should
// pin to, so the finding lands on the offending line rather than on the
// enclosing object. A cluster with no case here still gets a location from
// originKey, just a coarser one, so a kin error type nobody has looked at
// yet degrades in precision rather than reporting no location at all.
//
// Key is the right answer, not merely the default, whenever the offending
// thing has no scalar field of its own to point at: a constraint violated
// across several fields (none of which may even be present), a duplicated
// object, or a map key, since Origin.Fields tracks struct fields rather
// than map keys. Those clusters are left to originKey deliberately.
//
// Returns nil only for a cluster carrying no Origin at all (WebhookNilError,
// whose offending key is on the document root, which the loader does not
// track per-key), or when the document was loaded without origin tracking.
func locationForKinError(err error) *openapi3.Location {
	if rfe, ok := errors.AsType[*openapi3.RequiredFieldError](err); ok && rfe.Origin != nil {
		return fieldLoc(rfe.Origin, rfe.Field)
	}
	if fvm, ok := errors.AsType[*openapi3.FieldVersionMismatchError](err); ok && fvm.Origin != nil {
		return fieldLoc(fvm.Origin, fvm.Field)
	}
	if sve, ok := errors.AsType[*openapi3.SchemaValueError](err); ok && sve.Origin != nil {
		// SchemaValueError carries ValueKind (e.g. "example", "default")
		// — the per-field key under the schema where the value lives.
		return fieldLoc(sve.Origin, sve.ValueKind)
	}
	if mef, ok := errors.AsType[*openapi3.MutuallyExclusiveFieldsError](err); ok && mef.Origin != nil {
		// Prefer Field1's location; either offender pins the finding to
		// the right object. We don't carry both since a single Source
		// is the contract.
		return fieldLoc(mef.Origin, mef.Field1)
	}
	if ffe, ok := errors.AsType[*openapi3.ForbiddenFieldError](err); ok && ffe.Origin != nil {
		return fieldLoc(ffe.Origin, ffe.Field)
	}
	if sbf, ok := errors.AsType[*openapi3.SchemaBothFormsExclusive](err); ok && sbf.Origin != nil {
		return fieldLoc(sbf.Origin, sbf.Field)
	}
	if sec, ok := errors.AsType[*openapi3.SingleEntryContentError](err); ok && sec.Origin != nil {
		return fieldLoc(sec.Origin, sec.Subject)
	}
	if ste, ok := errors.AsType[*openapi3.SchemaTypeError](err); ok && ste.Origin != nil {
		// SchemaTypeError fires on the offending `type:` field of a
		// schema. Pin to the type field if the Origin tracks it,
		// otherwise to the schema's Key.
		return fieldLoc(ste.Origin, "type")
	}
	if doid, ok := errors.AsType[*openapi3.DuplicateOperationIDError](err); ok && doid.Origin != nil {
		// Pin to the offending operationId scalar inside the second
		// operation (not the operation's start), since the duplicate
		// is the field value, not the surrounding block. Falls back
		// to the operation Key if the loader didn't track the field.
		return fieldLoc(doid.Origin, "operationId")
	}
	if ipe, ok := errors.AsType[*openapi3.InvalidParameterInError](err); ok && ipe.Origin != nil {
		// Pin to the parameter's `in` field if the loader tracks it,
		// otherwise the parameter object's Key.
		return fieldLoc(ipe.Origin, "in")
	}
	if spre, ok := errors.AsType[*openapi3.SchemaPatternRegexError](err); ok && spre.Origin != nil {
		// Pin to the schema's `pattern` field where the bad regex
		// lives, otherwise the schema's Key.
		return fieldLoc(spre.Origin, "pattern")
	}
	if isste, ok := errors.AsType[*openapi3.InvalidSecuritySchemeTypeError](err); ok && isste.Origin != nil {
		return fieldLoc(isste.Origin, "type")
	}
	if ihse, ok := errors.AsType[*openapi3.InvalidHTTPSchemeError](err); ok && ihse.Origin != nil {
		return fieldLoc(ihse.Origin, "scheme")
	}
	if ure, ok := errors.AsType[*openapi3.UnresolvedRefError](err); ok && ure.Origin != nil {
		// Pin to the $ref field if the loader tracks it, otherwise
		// the ref-bearing object's Key.
		return fieldLoc(ure.Origin, "$ref")
	}
	if akie, ok := errors.AsType[*openapi3.APIKeyInInvalidError](err); ok && akie.Origin != nil {
		return fieldLoc(akie.Origin, "in")
	}
	if isme, ok := errors.AsType[*openapi3.InvalidSerializationMethodError](err); ok && isme.Origin != nil {
		// Pin to the `style` field if the loader tracks it on the
		// encoding/parameter/header object.
		return fieldLoc(isme.Origin, "style")
	}
	if aod, ok := errors.AsType[*openapi3.AdditionalOperationsDuplicateMethodError](err); ok && aod.Origin != nil {
		return additionalOperationsLoc(aod.Origin)
	}
	if aoi, ok := errors.AsType[*openapi3.AdditionalOperationsInvalidMethodError](err); ok && aoi.Origin != nil {
		return additionalOperationsLoc(aoi.Origin)
	}
	if drf, ok := errors.AsType[*openapi3.DuplicateRequiredFieldError](err); ok && drf.Origin != nil {
		// Origin is the schema; the duplicate is inside its `required`
		// array, so pin there rather than at the schema's own line.
		return fieldLoc(drf.Origin, "required")
	}
	return originKey(err)
}

// additionalOperationsLoc pins an additionalOperations key error to the
// `additionalOperations:` field of the path item.
//
// The offending thing is a key inside that map, and the loader does record it
// on the keyed Operation's own Origin, but the error carries the path item's
// Origin, so the map field is as close as the error itself reaches. Without
// this the fallback would pin to the path item instead.
func additionalOperationsLoc(origin *openapi3.Origin) *openapi3.Location {
	return fieldLoc(origin, "additionalOperations")
}

// originKey returns the Key of whatever Origin the error carries, or nil if it
// carries none.
//
// The lookup is by reflection rather than a type switch on purpose. Every kin
// validation error that tracks a source location exposes it the same way, as a
// public `Origin *openapi3.Origin` field, so reading that field covers the
// clusters locationForKinError refines, the ones where the enclosing object is
// the right answer, and any type a later kin release adds, without oasdiff
// keeping a parallel list of kin's error types.
//
// The coupling is to the field's name and type. TestOriginKey_ReadsTheOriginField
// pins both, so a kin rename fails there rather than silently dropping every
// location.
//
// kin wraps the offending error in section, path and operation context, so the
// Origin is on a leaf rather than on what the caller holds. Walk the chain the
// way errors.AsType does and take the first Origin found, innermost wrappers
// last, so the nearest enclosing object wins over a distant ancestor.
func originKey(err error) *openapi3.Location {
	for err != nil {
		if origin := declaredOrigin(err); origin != nil && origin.Key != nil {
			return origin.Key
		}
		switch unwrapped := err.(type) {
		case interface{ Unwrap() error }:
			err = unwrapped.Unwrap()
		case interface{ Unwrap() []error }:
			for _, e := range unwrapped.Unwrap() {
				if loc := originKey(e); loc != nil {
					return loc
				}
			}
			return nil
		default:
			return nil
		}
	}
	return nil
}

// declaredOrigin reads an error's public `Origin *openapi3.Origin` field, or
// nil when it has none.
func declaredOrigin(err error) *openapi3.Origin {
	v := reflect.ValueOf(err)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return nil
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return nil
	}
	field := v.FieldByName("Origin")
	if !field.IsValid() || !field.CanInterface() {
		return nil
	}
	origin, _ := field.Interface().(*openapi3.Origin)
	return origin
}

// schemaFieldLocation returns the line/column of the named field inside a
// schema when origin tracking is enabled, falling back to the schema's own
// location, then 0. Shared by the native lints, which each report a finding
// against one schema field ("enum", "default", "format").
func schemaFieldLocation(s *openapi3.Schema, field string) (int, int) {
	if s == nil || s.Origin == nil {
		return 0, 0
	}
	if loc := fieldLoc(s.Origin, field); loc != nil {
		return loc.Line, loc.Column
	}
	return 0, 0
}

// fieldLoc returns the location of a specific scalar field inside an
// Origin's collection if present; otherwise the collection's Key.
// Lookup is by the leaf field name (e.g. "identifier" for license
// errors, "version" for info errors).
func fieldLoc(origin *openapi3.Origin, field string) *openapi3.Location {
	if origin == nil {
		return nil
	}
	if loc, ok := origin.Fields.Lookup(field); ok {
		return &loc
	}
	// Cluster errors carry dotted Field names (e.g. "info.version") for
	// disambiguation in the rule ID, but kin's Origin.Fields is keyed by
	// the leaf name as it appears in the YAML mapping ("version"). When
	// the full name doesn't match, fall back to the suffix after the
	// last dot so we still resolve to the precise field location instead
	// of the parent object's Key.
	if i := strings.LastIndex(field, "."); i >= 0 {
		if loc, ok := origin.Fields.Lookup(field[i+1:]); ok {
			return &loc
		}
	}
	return origin.Key
}

// paramTypeLocation returns the line/column of the schema's type field, adding
// one fallback beyond schemaFieldLocation: the parameter's own location, for a
// parameter whose schema carries no origin at all.
func paramTypeLocation(p *openapi3.Parameter) (int, int) {
	if line, column := schemaFieldLocation(p.Schema.Value, "type"); line != 0 || column != 0 {
		return line, column
	}
	if p.Origin != nil && p.Origin.Key != nil {
		return p.Origin.Key.Line, p.Origin.Key.Column
	}
	return 0, 0
}
