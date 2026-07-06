// Package validate exposes spec validation as a public library API.
// Validate runs kin-openapi's spec validator and returns a structured list
// of findings (rule IDs, severities, source locations) so any caller can
// surface validation results in the same shape as `oasdiff validate` does
// on the command line.
package validate

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
)

// unknownValidationID is the fallback rule ID for any spec-validation
// error our dispatcher (ruleIDForKinError) has no errors.As arm for.
// If we encounter this in the output, we should replace it with a more
// specific ID.
const unknownValidationID = "spec-validation-error"

// Validate validates the spec against the OpenAPI and JSON Schema rules
// (kin-openapi's validator), returning a flat list of findings. Each
// finding carries a stable rule ID, severity, source location (when origin
// tracking is enabled on the loader), and a fingerprint for cross-spec
// matching.
//
// source is the display name for the spec (typically its file path). It
// appears in each finding's Source.File so callers can render
// file:line:column anchors. Pass an empty string when there is no
// meaningful source name (e.g. specs loaded from memory).
//
// A valid spec yields a non-nil empty Findings (nil only for the nil-spec
// guard), so the formatters' nil guard doesn't collapse `[]` to empty bytes.
func Validate(spec *openapi3.T, source string) formatters.Findings {
	if spec == nil {
		return nil
	}
	verr := spec.Validate(context.Background(), openapi3.EnableMultiError())
	if verr == nil {
		return formatters.Findings{}
	}
	return mapKinErrors(source, verr)
}

// mapKinErrors flattens kin-openapi's MultiError tree into a list of
// findings. kin can return either a single error or a MultiError; the
// MultiError can itself contain MultiErrors, so we recurse.
func mapKinErrors(source string, err error) formatters.Findings {
	return dedupePreferringComponents(flattenKinErrors(source, err))
}

// flattenKinErrors walks the kin error tree (MultiError → leaves) and
// produces one finding per leaf. Findings may include duplicates of the
// same defect when a shared definition (e.g. a schema in components) is
// referenced from multiple operations — the validator re-visits the
// definition under each $ref. dedupePreferringComponents collapses those
// into a single finding.
func flattenKinErrors(source string, err error) formatters.Findings {
	if err == nil {
		return nil
	}
	if me, ok := err.(openapi3.MultiError); ok {
		var out formatters.Findings
		for _, sub := range me {
			out = append(out, flattenKinErrors(source, sub)...)
		}
		return out
	}
	path, operation := pathOperationForKinError(err)
	f := formatters.Finding{
		Id:        ruleIDForKinError(err),
		Text:      unwrapContext(err).Error(),
		Level:     severityForKinError(err),
		Operation: operation,
		Path:      path,
		Section:   sectionForKinError(err),
		Source: formatters.Source{
			File:   source,
			Line:   lineForKinError(err),
			Column: columnForKinError(err),
		},
	}
	// Fingerprint last so it hashes over the populated fields.
	f.Fingerprint = checker.ComputeFingerprint(f.Id, f.Operation, f.Path, argsForKinError(err))
	return formatters.Findings{f}
}

// severityForKinError classifies a kin validation error into a severity.
// The default is ERR: every kin Validate result is a spec violation, so
// errors are the safe default and any kin cluster we don't recognise (or
// a newly-typed one) stays an error. A few clusters are downgraded
// because the spec still parses and the issue is a portability or
// doc-accuracy concern rather than a structural break.
//
// The mapping is hardcoded for now. Per-rule severity customization (like
// the changelog command's --severity-levels) can be layered on later, so
// these classifications aren't engraved in stone.
//
// Deliberately kept as errors despite being downgrade candidates:
// duplicate-operation-id violates the spec's uniqueness MUST and breaks
// code generators that key method names off operationId.
func severityForKinError(err error) checker.Level {
	// INFO: an example that doesn't match its schema is a documentation
	// accuracy nit; the contract (the schema) is still valid. A default,
	// by contrast, is consumed at runtime by some tooling, so a mismatch
	// there is a real risk and stays a warning.
	var sve *openapi3.SchemaValueError
	if errors.As(err, &sve) {
		if sve.ValueKind == "example" {
			return checker.INFO
		}
		return checker.WARN
	}

	// WARN: structurally valid but a portability or correctness risk.
	var fvm *openapi3.FieldVersionMismatchError
	if errors.As(err, &fvm) {
		// e.g. a 3.1-only field in a doc that declares an older version.
		return checker.WARN
	}
	var esf *openapi3.ExtraSiblingFieldsError
	if errors.As(err, &esf) {
		// Fields alongside a $ref are silently ignored in 3.0 rather than
		// breaking the spec; the author's intent is lost, not the document.
		return checker.WARN
	}
	var cpe *openapi3.ConflictingPathsError
	if errors.As(err, &cpe) {
		return checker.WARN
	}
	var dpe *openapi3.DuplicateParameterError
	if errors.As(err, &dpe) {
		return checker.WARN
	}

	return checker.ERR
}

// dedupePreferringComponents groups findings by their underlying
// defect identity (Id + Source location + Text — which carries the
// args-derived discriminator) and keeps one representative per group.
// When the group has a components-rooted finding (Section ==
// "components"), prefer it: the components-rooted version points at
// the definition site and has empty Operation/Path, giving a stable
// fingerprint across reference-graph changes.
//
// This covers the common case where a defect in components/schemas/X
// (or any components sub-section) is reported once from the components
// walk and once from each operation that $refs it. Path-level shared
// parameters don't need handling here because kin only validates them
// once at the PathItem level (no per-operation re-validation).
func dedupePreferringComponents(in formatters.Findings) formatters.Findings {
	type group struct {
		first  int // index into `in` of the first finding for this key
		chosen int // index of the current best representative
	}
	keyOf := func(f formatters.Finding) string {
		return f.Id + "\x00" + f.Source.File + "\x00" +
			strconv.Itoa(f.Source.Line) + "\x00" +
			strconv.Itoa(f.Source.Column) + "\x00" + f.Text
	}
	groups := make(map[string]*group)
	var order []string // preserve first-seen order for stable output
	for i, f := range in {
		k := keyOf(f)
		g, ok := groups[k]
		if !ok {
			groups[k] = &group{first: i, chosen: i}
			order = append(order, k)
			continue
		}
		// Already have a candidate; prefer this one only if it's
		// components-rooted and the current pick isn't.
		if f.Section == "components" && in[g.chosen].Section != "components" {
			g.chosen = i
		}
	}
	out := make(formatters.Findings, 0, len(order))
	for _, k := range order {
		out = append(out, in[groups[k].chosen])
	}
	return out
}

// pathOperationForKinError extracts the path template and HTTP method
// from kin's typed context wrappers (PathValidationError and
// OperationValidationError, added in getkin/kin-openapi #1183). Either
// return value is "" when the error chain carries no such scope, e.g.
// doc-root findings like info-version-required.
func pathOperationForKinError(err error) (path, operation string) {
	var pve *openapi3.PathValidationError
	if errors.As(err, &pve) {
		path = pve.Path
	}
	var ove *openapi3.OperationValidationError
	if errors.As(err, &ove) {
		operation = ove.Method
	}
	return path, operation
}

// unwrapContext strips kin's structural context wrappers
// (SectionValidationError / PathValidationError / OperationValidationError,
// kin #1183) from the front of the chain. That section/path/operation
// scope is captured in the Finding's typed fields, so Text should carry
// only the underlying message, without the redundant "invalid <scope>:"
// prefixes those wrappers add to Error().
func unwrapContext(err error) error {
	for {
		switch err.(type) {
		case *openapi3.SectionValidationError,
			*openapi3.PathValidationError,
			*openapi3.OperationValidationError,
			*openapi3.ComponentValidationError,
			*openapi3.ExternalDocsURLValidationError,
			*openapi3.HeaderFieldValidationError,
			*openapi3.MediaTypeExampleValidationError,
			*openapi3.WebhookValidationError,
			*openapi3.ParameterFieldValidationError,
			*openapi3.ParameterExampleValidationError,
			*openapi3.SecuritySchemeFlowValidationError,
			*openapi3.OAuthFlowValidationError,
			*openapi3.OAuthFlowFieldValidationError:
			u := errors.Unwrap(err)
			if u == nil {
				return err
			}
			err = u
		default:
			return err
		}
	}
}

// sectionForKinError maps a typed kin error to its logical doc section,
// matching the values used by ApiChange / ComponentChange / SecurityChange
// in the existing changelog output (`paths`, `components`, `security`).
//
// The mapping is per-cluster + a light Field-prefix check on the cluster
// types that carry one (RequiredFieldError, FieldVersionMismatchError).
// Doc-root findings without a natural section (e.g. *RequiredFieldError
// {Field: "openapi"}) get the empty string.
func sectionForKinError(err error) string {
	// SectionValidationError (kin #1183) names the section directly and
	// authoritatively; prefer it over the cluster heuristics below, which
	// predate it and only approximate (e.g. they miscount inline component
	// schemas as "paths"). The cluster logic remains the fallback for
	// doc-root errors that aren't wrapped in a section at all.
	var secErr *openapi3.SectionValidationError
	if errors.As(err, &secErr) {
		return secErr.Section
	}

	// Cluster types that have a structural section regardless of payload.
	var ppe *openapi3.PathParametersError
	if errors.As(err, &ppe) {
		return "paths"
	}
	var wne *openapi3.WebhookNilError
	if errors.As(err, &wne) {
		return "webhooks"
	}
	var sute *openapi3.ServerURLTemplateError
	if errors.As(err, &sute) {
		return "servers"
	}
	var pre *openapi3.PathParameterRequiredError
	if errors.As(err, &pre) {
		return "paths"
	}
	var doid *openapi3.DuplicateOperationIDError
	if errors.As(err, &doid) {
		return "paths"
	}
	var ipe *openapi3.InvalidParameterInError
	if errors.As(err, &ipe) {
		return "paths"
	}
	var isste *openapi3.InvalidSecuritySchemeTypeError
	if errors.As(err, &isste) {
		return "components"
	}
	var ihse *openapi3.InvalidHTTPSchemeError
	if errors.As(err, &ihse) {
		return "components"
	}
	var akie *openapi3.APIKeyInInvalidError
	if errors.As(err, &akie) {
		return "components"
	}
	var pmss *openapi3.PathMustStartWithSlashError
	if errors.As(err, &pmss) {
		return "paths"
	}
	var cpe *openapi3.ConflictingPathsError
	if errors.As(err, &cpe) {
		return "paths"
	}
	var dpe *openapi3.DuplicateParameterError
	if errors.As(err, &dpe) {
		return "paths"
	}

	// Cluster types with a Field that hints at the section.
	var rfe *openapi3.RequiredFieldError
	if errors.As(err, &rfe) {
		return sectionFromField(rfe.Field)
	}
	var fvm *openapi3.FieldVersionMismatchError
	if errors.As(err, &fvm) {
		return sectionFromField(fvm.Field)
	}

	// Schema-deep clusters: lean toward "paths" since most kin
	// validation surfaces from request/response schemas inside
	// operations. Inline component schemas miscount here, but the
	// section is a navigational hint, not a hard claim.
	var sve *openapi3.SchemaValueError
	if errors.As(err, &sve) {
		return "paths"
	}
	var sbf *openapi3.SchemaBothFormsExclusive
	if errors.As(err, &sbf) {
		return "paths"
	}
	return ""
}

// sectionFromField returns the section a kin Field name lives in,
// based on the field's top-level prefix. Anything not recognised
// returns empty.
func sectionFromField(field string) string {
	switch {
	case strings.HasPrefix(field, "info"):
		return "info"
	case strings.HasPrefix(field, "paths"):
		return "paths"
	case strings.HasPrefix(field, "components"):
		return "components"
	case strings.HasPrefix(field, "webhooks"):
		return "webhooks"
	case strings.HasPrefix(field, "servers"):
		return "servers"
	case strings.HasPrefix(field, "security"):
		return "security"
	case strings.HasPrefix(field, "tags"):
		return "tags"
	default:
		return ""
	}
}

// argsForKinError returns the disambiguating args used in fingerprint
// computation. For most validate clusters the args list is the
// cluster's structured Field (or Fields); for clusters that carry no
// per-finding field, the args are empty and identity is already
// captured by the rule ID + Source.
func argsForKinError(err error) []any {
	var rfe *openapi3.RequiredFieldError
	if errors.As(err, &rfe) {
		return []any{rfe.Field}
	}
	var fvm *openapi3.FieldVersionMismatchError
	if errors.As(err, &fvm) {
		return []any{fvm.Field}
	}
	var mef *openapi3.MutuallyExclusiveFieldsError
	if errors.As(err, &mef) {
		return []any{mef.Field1, mef.Field2}
	}
	var ffe *openapi3.ForbiddenFieldError
	if errors.As(err, &ffe) {
		return []any{ffe.Field}
	}
	var efr *openapi3.EitherFieldRequiredError
	if errors.As(err, &efr) {
		return []any{strings.Join(efr.Fields, "-or-")}
	}
	var sbf *openapi3.SchemaBothFormsExclusive
	if errors.As(err, &sbf) {
		return []any{sbf.Field}
	}
	var eofe *openapi3.ExactlyOneFieldError
	if errors.As(err, &eofe) {
		return []any{strings.Join(eofe.Fields, "-or-")}
	}
	var sec *openapi3.SingleEntryContentError
	if errors.As(err, &sec) {
		return []any{sec.Subject}
	}
	var sve *openapi3.SchemaValueError
	if errors.As(err, &sve) {
		return []any{sve.ValueKind}
	}
	var pre *openapi3.PathParameterRequiredError
	if errors.As(err, &pre) {
		return []any{pre.Param}
	}
	var doid *openapi3.DuplicateOperationIDError
	if errors.As(err, &doid) {
		return []any{doid.OperationID}
	}
	var esf *openapi3.ExtraSiblingFieldsError
	if errors.As(err, &esf) {
		return []any{strings.Join(esf.Fields, ",")}
	}
	var ste *openapi3.SchemaTypeError
	if errors.As(err, &ste) {
		return []any{ste.Type}
	}
	var ipe *openapi3.InvalidParameterInError
	if errors.As(err, &ipe) {
		return []any{ipe.Value}
	}
	var spre *openapi3.SchemaPatternRegexError
	if errors.As(err, &spre) {
		return []any{spre.Pattern}
	}
	var isste *openapi3.InvalidSecuritySchemeTypeError
	if errors.As(err, &isste) {
		return []any{isste.Type}
	}
	var ihse *openapi3.InvalidHTTPSchemeError
	if errors.As(err, &ihse) {
		return []any{ihse.Scheme}
	}
	var ure *openapi3.UnresolvedRefError
	if errors.As(err, &ure) {
		return []any{ure.Ref}
	}
	var akie *openapi3.APIKeyInInvalidError
	if errors.As(err, &akie) {
		return []any{akie.Value}
	}
	var pmss *openapi3.PathMustStartWithSlashError
	if errors.As(err, &pmss) {
		return []any{pmss.Path}
	}
	var cpe *openapi3.ConflictingPathsError
	if errors.As(err, &cpe) {
		// Fingerprint by both paths in sorted order so flipped
		// argument order produces a stable identity.
		p1, p2 := cpe.Path1, cpe.Path2
		if p1 > p2 {
			p1, p2 = p2, p1
		}
		return []any{p1 + "|" + p2}
	}
	var dpe *openapi3.DuplicateParameterError
	if errors.As(err, &dpe) {
		return []any{dpe.In + ":" + dpe.Name}
	}
	var isme *openapi3.InvalidSerializationMethodError
	if errors.As(err, &isme) {
		return []any{isme.Subject, isme.Style, isme.Explode}
	}
	return nil
}

// lineForKinError extracts the line number from the typed cluster
// errors' Origin. Returns 0 when origin metadata isn't available
// (untyped error, doc-root field, or loader.IncludeOrigin = false).
func lineForKinError(err error) int {
	if k := locationForKinError(err); k != nil {
		return k.Line
	}
	return 0
}

// columnForKinError extracts the column number from the typed cluster
// errors' Origin. Returns 0 when origin metadata isn't available.
func columnForKinError(err error) int {
	if k := locationForKinError(err); k != nil {
		return k.Column
	}
	return 0
}

// locationForKinError returns the most-specific *Location available
// for a typed kin error. kin's Origin model:
//
//   - Origin.Key       points at the start of the enclosing collection
//     (e.g. for a LicenseIdentifierFieldFor31Plus, Key is the line of
//     the parent "license:" key, not "identifier:").
//   - Origin.Fields[X] points at the specific scalar field X inside
//     that collection.
//
// For clusters that carry a Field, we want Fields[Field] (the
// offending line) rather than Key (the enclosing object's line).
// Falls back to Key when the per-field entry is missing, and finally
// to nil for clusters with no Origin at all (WebhookNilError).
func locationForKinError(err error) *openapi3.Location {
	var rfe *openapi3.RequiredFieldError
	if errors.As(err, &rfe) && rfe.Origin != nil {
		return fieldLoc(rfe.Origin, rfe.Field)
	}
	var fvm *openapi3.FieldVersionMismatchError
	if errors.As(err, &fvm) && fvm.Origin != nil {
		return fieldLoc(fvm.Origin, fvm.Field)
	}
	var sve *openapi3.SchemaValueError
	if errors.As(err, &sve) && sve.Origin != nil {
		// SchemaValueError carries ValueKind (e.g. "example", "default")
		// — the per-field key under the schema where the value lives.
		return fieldLoc(sve.Origin, sve.ValueKind)
	}
	var ppe *openapi3.PathParametersError
	if errors.As(err, &ppe) && ppe.Origin != nil {
		return ppe.Origin.Key
	}
	var mef *openapi3.MutuallyExclusiveFieldsError
	if errors.As(err, &mef) && mef.Origin != nil {
		// Prefer Field1's location; either offender pins the finding to
		// the right object. We don't carry both since a single Source
		// is the contract.
		return fieldLoc(mef.Origin, mef.Field1)
	}
	var ffe *openapi3.ForbiddenFieldError
	if errors.As(err, &ffe) && ffe.Origin != nil {
		return fieldLoc(ffe.Origin, ffe.Field)
	}
	var sute *openapi3.ServerURLTemplateError
	if errors.As(err, &sute) && sute.Origin != nil {
		return sute.Origin.Key
	}
	var efr *openapi3.EitherFieldRequiredError
	if errors.As(err, &efr) && efr.Origin != nil {
		// EitherFieldRequiredError fires when none of {Fields...} are
		// present, so per-field lookup wouldn't match anything — the
		// enclosing object's Key is the right pin.
		return efr.Origin.Key
	}
	var sbf *openapi3.SchemaBothFormsExclusive
	if errors.As(err, &sbf) && sbf.Origin != nil {
		return fieldLoc(sbf.Origin, sbf.Field)
	}
	var eofe *openapi3.ExactlyOneFieldError
	if errors.As(err, &eofe) && eofe.Origin != nil {
		// Same reasoning as EitherFieldRequiredError: cluster fires
		// when the constraint is violated across multiple fields; the
		// object Key is the cleanest pin.
		return eofe.Origin.Key
	}
	var sec *openapi3.SingleEntryContentError
	if errors.As(err, &sec) && sec.Origin != nil {
		return fieldLoc(sec.Origin, sec.Subject)
	}
	var pre *openapi3.PathParameterRequiredError
	if errors.As(err, &pre) && pre.Origin != nil {
		// PathParameterRequiredError fires on a parameter declared with
		// in: path but without required: true. The Key of the parameter
		// object pins the finding correctly; the `required` field would
		// be more precise but is absent (that's the whole bug).
		return pre.Origin.Key
	}
	var ste *openapi3.SchemaTypeError
	if errors.As(err, &ste) && ste.Origin != nil {
		// SchemaTypeError fires on the offending `type:` field of a
		// schema. Pin to the type field if the Origin tracks it,
		// otherwise to the schema's Key.
		return fieldLoc(ste.Origin, "type")
	}
	var doid *openapi3.DuplicateOperationIDError
	if errors.As(err, &doid) && doid.Origin != nil {
		// Pin to the offending operationId scalar inside the second
		// operation (not the operation's start), since the duplicate
		// is the field value, not the surrounding block. Falls back
		// to the operation Key if the loader didn't track the field.
		return fieldLoc(doid.Origin, "operationId")
	}
	var esf *openapi3.ExtraSiblingFieldsError
	if errors.As(err, &esf) && esf.Origin != nil {
		// Origin points at the parent object carrying the unexpected
		// sibling fields. The fields themselves may not have Origin
		// entries (Yaml parser tracks structural keys, not the
		// offending ones), so the object Key is the best pin.
		return esf.Origin.Key
	}
	var ipe *openapi3.InvalidParameterInError
	if errors.As(err, &ipe) && ipe.Origin != nil {
		// Pin to the parameter's `in` field if the loader tracks it,
		// otherwise the parameter object's Key.
		return fieldLoc(ipe.Origin, "in")
	}
	var spre *openapi3.SchemaPatternRegexError
	if errors.As(err, &spre) && spre.Origin != nil {
		// Pin to the schema's `pattern` field where the bad regex
		// lives, otherwise the schema's Key.
		return fieldLoc(spre.Origin, "pattern")
	}
	var isste *openapi3.InvalidSecuritySchemeTypeError
	if errors.As(err, &isste) && isste.Origin != nil {
		return fieldLoc(isste.Origin, "type")
	}
	var ihse *openapi3.InvalidHTTPSchemeError
	if errors.As(err, &ihse) && ihse.Origin != nil {
		return fieldLoc(ihse.Origin, "scheme")
	}
	var ure *openapi3.UnresolvedRefError
	if errors.As(err, &ure) && ure.Origin != nil {
		// Pin to the $ref field if the loader tracks it, otherwise
		// the ref-bearing object's Key.
		return fieldLoc(ure.Origin, "$ref")
	}
	var akie *openapi3.APIKeyInInvalidError
	if errors.As(err, &akie) && akie.Origin != nil {
		return fieldLoc(akie.Origin, "in")
	}
	var pmss *openapi3.PathMustStartWithSlashError
	if errors.As(err, &pmss) && pmss.Origin != nil {
		// Origin is the paths object; the offending path key lives
		// inside it but Fields tracks struct fields, not map keys, so
		// fall back to the paths object's Key.
		return pmss.Origin.Key
	}
	var cpe *openapi3.ConflictingPathsError
	if errors.As(err, &cpe) && cpe.Origin != nil {
		return cpe.Origin.Key
	}
	var dpe *openapi3.DuplicateParameterError
	if errors.As(err, &dpe) && dpe.Origin != nil {
		// Origin is on the second (offending) parameter; pin to its Key.
		return dpe.Origin.Key
	}
	var isme *openapi3.InvalidSerializationMethodError
	if errors.As(err, &isme) && isme.Origin != nil {
		// Pin to the `style` field if the loader tracks it on the
		// encoding/parameter/header object.
		return fieldLoc(isme.Origin, "style")
	}
	// WebhookNilError carries no Origin (the offending key is on the
	// document root, which the loader doesn't track per-key).
	return nil
}

// fieldLoc returns the location of a specific scalar field inside an
// Origin's collection if present; otherwise the collection's Key.
// Lookup is by the leaf field name (e.g. "identifier" for license
// errors, "version" for info errors).
func fieldLoc(origin *openapi3.Origin, field string) *openapi3.Location {
	if origin == nil {
		return nil
	}
	if loc, ok := origin.Fields[field]; ok {
		return &loc
	}
	// Cluster errors carry dotted Field names (e.g. "info.version") for
	// disambiguation in the rule ID, but kin's Origin.Fields is keyed by
	// the leaf name as it appears in the YAML mapping ("version"). When
	// the full name doesn't match, fall back to the suffix after the
	// last dot so we still resolve to the precise field location instead
	// of the parent object's Key.
	if i := strings.LastIndex(field, "."); i >= 0 {
		if loc, ok := origin.Fields[field[i+1:]]; ok {
			return &loc
		}
	}
	return origin.Key
}

// ruleIDForKinError dispatches a kin-openapi error to a stable
// kebab-case rule ID. Uses the typed cluster wrappers from kin's
// openapi3 package; one arm per cluster covers all the call-site
// leaves wrapped by that cluster.
//
// kin errors not yet migrated to a cluster fall through to
// unknownValidationID. Where a cluster carries field-name metadata,
// the rule ID is derived from that — the per-leaf type isn't
// consulted because (a) the cluster carries enough metadata, and (b)
// deriving from a single field keeps the dispatch stable as kin adds
// new leaves. Where a cluster has no useful field for derivation
// (ServerURLTemplateError carries only the offending URL,
// PathParametersError carries Path+Method+Missing), a static rule ID
// is returned for the whole cluster.
func ruleIDForKinError(err error) string {
	var rfe *openapi3.RequiredFieldError
	if errors.As(err, &rfe) {
		return ruleIDFromField(rfe.Field) + "-required"
	}

	var fvm *openapi3.FieldVersionMismatchError
	if errors.As(err, &fvm) {
		return ruleIDFromField(fvm.Field) + "-field-for-3-1-plus"
	}

	var sve *openapi3.SchemaValueError
	if errors.As(err, &sve) {
		return ruleIDFromField(sve.ValueKind) + "-violates-schema"
	}

	var ppe *openapi3.PathParametersError
	if errors.As(err, &ppe) {
		return "path-parameters-mismatch"
	}

	var mef *openapi3.MutuallyExclusiveFieldsError
	if errors.As(err, &mef) {
		return ruleIDFromField(mef.Field1) + "-" + ruleIDFromField(mef.Field2) + "-mutually-exclusive"
	}

	var ffe *openapi3.ForbiddenFieldError
	if errors.As(err, &ffe) {
		return ruleIDFromField(ffe.Field) + "-forbidden"
	}

	var sute *openapi3.ServerURLTemplateError
	if errors.As(err, &sute) {
		return "server-url-template-invalid"
	}

	var efr *openapi3.EitherFieldRequiredError
	if errors.As(err, &efr) {
		return joinFieldsForRuleID(efr.Fields) + "-required"
	}

	var sbf *openapi3.SchemaBothFormsExclusive
	if errors.As(err, &sbf) {
		return ruleIDFromField(sbf.Field) + "-both-forms-exclusive"
	}

	var eofe *openapi3.ExactlyOneFieldError
	if errors.As(err, &eofe) {
		return joinFieldsForRuleID(eofe.Fields) + "-exactly-one"
	}

	var sec *openapi3.SingleEntryContentError
	if errors.As(err, &sec) {
		return ruleIDFromField(sec.Subject) + "-content-single-entry"
	}

	var wne *openapi3.WebhookNilError
	if errors.As(err, &wne) {
		return "webhook-nil"
	}

	var pre *openapi3.PathParameterRequiredError
	if errors.As(err, &pre) {
		return "path-parameter-required"
	}

	var doid *openapi3.DuplicateOperationIDError
	if errors.As(err, &doid) {
		return "duplicate-operation-id"
	}

	var esf *openapi3.ExtraSiblingFieldsError
	if errors.As(err, &esf) {
		return "extra-sibling-fields"
	}

	var ste *openapi3.SchemaTypeError
	if errors.As(err, &ste) {
		return "schema-type-unsupported"
	}

	var ipe *openapi3.InvalidParameterInError
	if errors.As(err, &ipe) {
		return "parameter-in-invalid"
	}

	var spre *openapi3.SchemaPatternRegexError
	if errors.As(err, &spre) {
		return "schema-pattern-regex-invalid"
	}

	var isste *openapi3.InvalidSecuritySchemeTypeError
	if errors.As(err, &isste) {
		return "security-scheme-type-invalid"
	}

	var ihse *openapi3.InvalidHTTPSchemeError
	if errors.As(err, &ihse) {
		return "security-scheme-http-scheme-invalid"
	}

	var ure *openapi3.UnresolvedRefError
	if errors.As(err, &ure) {
		return "unresolved-ref"
	}

	var akie *openapi3.APIKeyInInvalidError
	if errors.As(err, &akie) {
		return "security-scheme-apikey-in-invalid"
	}

	var pmss *openapi3.PathMustStartWithSlashError
	if errors.As(err, &pmss) {
		return "path-must-start-with-slash"
	}

	var cpe *openapi3.ConflictingPathsError
	if errors.As(err, &cpe) {
		return "conflicting-paths"
	}

	var dpe *openapi3.DuplicateParameterError
	if errors.As(err, &dpe) {
		return "duplicate-parameter"
	}

	var isme *openapi3.InvalidSerializationMethodError
	if errors.As(err, &isme) {
		return "serialization-method-invalid"
	}

	return unknownValidationID
}

// joinFieldsForRuleID renders an N-field "any/exactly one of" rule ID
// fragment as kebab-case fields joined by "-or-" (e.g. ["value",
// "externalValue"] → "value-or-external-value"). The caller appends
// the cluster-specific suffix ("-required", "-exactly-one", ...).
func joinFieldsForRuleID(fields []string) string {
	parts := make([]string, len(fields))
	for i, f := range fields {
		parts[i] = ruleIDFromField(f)
	}
	return strings.Join(parts, "-or-")
}

// ruleIDFromField normalises a field path into a kebab-case identifier.
// Strips a leading "$" (for JSON Schema keywords like "$defs"),
// replaces "." with "-" (for paths like "info.version"), and inserts
// "-" before each uppercase letter (for camelCase like "prefixItems").
func ruleIDFromField(field string) string {
	field = strings.TrimPrefix(field, "$")
	field = strings.ReplaceAll(field, ".", "-")
	var b strings.Builder
	for i, r := range field {
		if i > 0 && unicode.IsUpper(r) {
			b.WriteByte('-')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}
