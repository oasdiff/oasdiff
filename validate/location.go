package validate

// Source locations for findings. kin-openapi's Origin model is the substrate:
// Origin.Key points at the start of the enclosing collection and
// Origin.Fields.Lookup(X) at a specific scalar field inside it, so pinning a finding
// to the offending line means preferring the field entry over the key. These
// helpers own that resolution for both the kin-error path (locationForKinError)
// and the native lints (schemaFieldLocation).

import (
	"errors"
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
// For clusters that carry a Field, we want Fields.Lookup(Field) (the
// offending line) rather than Key (the enclosing object's line).
// Falls back to Key when the per-field entry is missing, and finally
// to nil for clusters with no Origin at all (WebhookNilError).
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
	if ppe, ok := errors.AsType[*openapi3.PathParametersError](err); ok && ppe.Origin != nil {
		return ppe.Origin.Key
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
	if sute, ok := errors.AsType[*openapi3.ServerURLTemplateError](err); ok && sute.Origin != nil {
		return sute.Origin.Key
	}
	if efr, ok := errors.AsType[*openapi3.EitherFieldRequiredError](err); ok && efr.Origin != nil {
		// EitherFieldRequiredError fires when none of {Fields...} are
		// present, so per-field lookup wouldn't match anything — the
		// enclosing object's Key is the right pin.
		return efr.Origin.Key
	}
	if sbf, ok := errors.AsType[*openapi3.SchemaBothFormsExclusive](err); ok && sbf.Origin != nil {
		return fieldLoc(sbf.Origin, sbf.Field)
	}
	if eofe, ok := errors.AsType[*openapi3.ExactlyOneFieldError](err); ok && eofe.Origin != nil {
		// Same reasoning as EitherFieldRequiredError: cluster fires
		// when the constraint is violated across multiple fields; the
		// object Key is the cleanest pin.
		return eofe.Origin.Key
	}
	if sec, ok := errors.AsType[*openapi3.SingleEntryContentError](err); ok && sec.Origin != nil {
		return fieldLoc(sec.Origin, sec.Subject)
	}
	if pre, ok := errors.AsType[*openapi3.PathParameterRequiredError](err); ok && pre.Origin != nil {
		// PathParameterRequiredError fires on a parameter declared with
		// in: path but without required: true. The Key of the parameter
		// object pins the finding correctly; the `required` field would
		// be more precise but is absent (that's the whole bug).
		return pre.Origin.Key
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
	if esf, ok := errors.AsType[*openapi3.ExtraSiblingFieldsError](err); ok && esf.Origin != nil {
		// Origin points at the parent object carrying the unexpected
		// sibling fields. The fields themselves may not have Origin
		// entries (Yaml parser tracks structural keys, not the
		// offending ones), so the object Key is the best pin.
		return esf.Origin.Key
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
	if pmss, ok := errors.AsType[*openapi3.PathMustStartWithSlashError](err); ok && pmss.Origin != nil {
		// Origin is the paths object; the offending path key lives
		// inside it but Fields tracks struct fields, not map keys, so
		// fall back to the paths object's Key.
		return pmss.Origin.Key
	}
	if cpe, ok := errors.AsType[*openapi3.ConflictingPathsError](err); ok && cpe.Origin != nil {
		return cpe.Origin.Key
	}
	if dpe, ok := errors.AsType[*openapi3.DuplicateParameterError](err); ok && dpe.Origin != nil {
		// Origin is on the second (offending) parameter; pin to its Key.
		return dpe.Origin.Key
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
	// WebhookNilError carries no Origin (the offending key is on the
	// document root, which the loader doesn't track per-key).
	return nil
}

// additionalOperationsLoc pins an additionalOperations key error to the
// `additionalOperations:` field of the path item.
//
// The offending thing is a key inside that map, and the loader does record it
// (on the keyed Operation's own Origin), but the error carries the path item's
// Origin, so the map field is as close as the error itself reaches. That still
// lands in the right block rather than nowhere, which is what these two
// reported before. Falls back to the path item's Key.
func additionalOperationsLoc(origin *openapi3.Origin) *openapi3.Location {
	return fieldLoc(origin, "additionalOperations")
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
