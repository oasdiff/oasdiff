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
	// Non-nil empty so the formatters' nil guard renders `[]`, not empty
	// bytes, for a valid spec with no findings (oasdiff #1045/#1046).
	findings := formatters.Findings{}
	if verr := spec.Validate(context.Background(), openapi3.EnableMultiError()); verr != nil {
		findings = mapKinErrors(source, verr)
	}
	// oasdiff-native SHOULD-level lints that kin-openapi does not enforce.
	findings = append(findings, lintDuplicateEnums(spec, source)...)
	findings = append(findings, lintAmbiguousParamSerialization(spec, source)...)
	findings = append(findings, lintRequiredWithDefault(spec, source)...)
	return findings
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
		Id:        knownRuleID(ruleIDForKinError(err)),
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
	if sve, ok := errors.AsType[*openapi3.SchemaValueError](err); ok {
		if sve.ValueKind == "example" {
			return checker.INFO
		}
		return checker.WARN
	}

	// WARN: structurally valid but a portability or correctness risk.
	if _, ok := errors.AsType[*openapi3.FieldVersionMismatchError](err); ok {
		// e.g. a 3.1-only field in a doc that declares an older version.
		return checker.WARN
	}
	if _, ok := errors.AsType[*openapi3.ExtraSiblingFieldsError](err); ok {
		// Fields alongside a $ref are silently ignored in 3.0 rather than
		// breaking the spec; the author's intent is lost, not the document.
		return checker.WARN
	}
	if _, ok := errors.AsType[*openapi3.ConflictingPathsError](err); ok {
		return checker.WARN
	}
	if _, ok := errors.AsType[*openapi3.DuplicateParameterError](err); ok {
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
	if pve, ok := errors.AsType[*openapi3.PathValidationError](err); ok {
		path = pve.Path
	}
	if ove, ok := errors.AsType[*openapi3.OperationValidationError](err); ok {
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
	if secErr, ok := errors.AsType[*openapi3.SectionValidationError](err); ok {
		return secErr.Section
	}

	// Cluster types that have a structural section regardless of payload.
	if _, ok := errors.AsType[*openapi3.PathParametersError](err); ok {
		return "paths"
	}
	if _, ok := errors.AsType[*openapi3.WebhookNilError](err); ok {
		return "webhooks"
	}
	if _, ok := errors.AsType[*openapi3.ServerURLTemplateError](err); ok {
		return "servers"
	}
	if _, ok := errors.AsType[*openapi3.PathParameterRequiredError](err); ok {
		return "paths"
	}
	if _, ok := errors.AsType[*openapi3.DuplicateOperationIDError](err); ok {
		return "paths"
	}
	if _, ok := errors.AsType[*openapi3.InvalidParameterInError](err); ok {
		return "paths"
	}
	if _, ok := errors.AsType[*openapi3.InvalidSecuritySchemeTypeError](err); ok {
		return "components"
	}
	if _, ok := errors.AsType[*openapi3.InvalidHTTPSchemeError](err); ok {
		return "components"
	}
	if _, ok := errors.AsType[*openapi3.APIKeyInInvalidError](err); ok {
		return "components"
	}
	if _, ok := errors.AsType[*openapi3.PathMustStartWithSlashError](err); ok {
		return "paths"
	}
	if _, ok := errors.AsType[*openapi3.ConflictingPathsError](err); ok {
		return "paths"
	}
	if _, ok := errors.AsType[*openapi3.DuplicateParameterError](err); ok {
		return "paths"
	}

	// Cluster types with a Field that hints at the section.
	if rfe, ok := errors.AsType[*openapi3.RequiredFieldError](err); ok {
		return sectionFromField(rfe.Field)
	}
	if fvm, ok := errors.AsType[*openapi3.FieldVersionMismatchError](err); ok {
		return sectionFromField(fvm.Field)
	}

	// Schema-deep clusters: lean toward "paths" since most kin
	// validation surfaces from request/response schemas inside
	// operations. Inline component schemas miscount here, but the
	// section is a navigational hint, not a hard claim.
	if _, ok := errors.AsType[*openapi3.SchemaValueError](err); ok {
		return "paths"
	}
	if _, ok := errors.AsType[*openapi3.SchemaBothFormsExclusive](err); ok {
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
	if rfe, ok := errors.AsType[*openapi3.RequiredFieldError](err); ok {
		return []any{rfe.Field}
	}
	if fvm, ok := errors.AsType[*openapi3.FieldVersionMismatchError](err); ok {
		return []any{fvm.Field}
	}
	if mef, ok := errors.AsType[*openapi3.MutuallyExclusiveFieldsError](err); ok {
		return []any{mef.Field1, mef.Field2}
	}
	if ffe, ok := errors.AsType[*openapi3.ForbiddenFieldError](err); ok {
		return []any{ffe.Field}
	}
	if efr, ok := errors.AsType[*openapi3.EitherFieldRequiredError](err); ok {
		return []any{strings.Join(efr.Fields, "-or-")}
	}
	if sbf, ok := errors.AsType[*openapi3.SchemaBothFormsExclusive](err); ok {
		return []any{sbf.Field}
	}
	if eofe, ok := errors.AsType[*openapi3.ExactlyOneFieldError](err); ok {
		return []any{strings.Join(eofe.Fields, "-or-")}
	}
	if sec, ok := errors.AsType[*openapi3.SingleEntryContentError](err); ok {
		return []any{sec.Subject}
	}
	if sve, ok := errors.AsType[*openapi3.SchemaValueError](err); ok {
		return []any{sve.ValueKind}
	}
	if pre, ok := errors.AsType[*openapi3.PathParameterRequiredError](err); ok {
		return []any{pre.Param}
	}
	if doid, ok := errors.AsType[*openapi3.DuplicateOperationIDError](err); ok {
		return []any{doid.OperationID}
	}
	if esf, ok := errors.AsType[*openapi3.ExtraSiblingFieldsError](err); ok {
		return []any{strings.Join(esf.Fields, ",")}
	}
	if ste, ok := errors.AsType[*openapi3.SchemaTypeError](err); ok {
		return []any{ste.Type}
	}
	if ipe, ok := errors.AsType[*openapi3.InvalidParameterInError](err); ok {
		return []any{ipe.Value}
	}
	if spre, ok := errors.AsType[*openapi3.SchemaPatternRegexError](err); ok {
		return []any{spre.Pattern}
	}
	if isste, ok := errors.AsType[*openapi3.InvalidSecuritySchemeTypeError](err); ok {
		return []any{isste.Type}
	}
	if ihse, ok := errors.AsType[*openapi3.InvalidHTTPSchemeError](err); ok {
		return []any{ihse.Scheme}
	}
	if ure, ok := errors.AsType[*openapi3.UnresolvedRefError](err); ok {
		return []any{ure.Ref}
	}
	if akie, ok := errors.AsType[*openapi3.APIKeyInInvalidError](err); ok {
		return []any{akie.Value}
	}
	if pmss, ok := errors.AsType[*openapi3.PathMustStartWithSlashError](err); ok {
		return []any{pmss.Path}
	}
	if cpe, ok := errors.AsType[*openapi3.ConflictingPathsError](err); ok {
		// Fingerprint by both paths in sorted order so flipped
		// argument order produces a stable identity.
		p1, p2 := cpe.Path1, cpe.Path2
		if p1 > p2 {
			p1, p2 = p2, p1
		}
		return []any{p1 + "|" + p2}
	}
	if dpe, ok := errors.AsType[*openapi3.DuplicateParameterError](err); ok {
		return []any{dpe.In + ":" + dpe.Name}
	}
	if isme, ok := errors.AsType[*openapi3.InvalidSerializationMethodError](err); ok {
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

// ruleIDForKinError returns the stable code kin-openapi declares on the
// validation error (openapi3.CodedError, one per rule), or unknownValidationID
// for an error that carries no code. knownRuleID gates the result against the
// registry. See TestRuleIDs_MatchKinCatalog for the registry/kin contract.
func ruleIDForKinError(err error) string {
	var coded openapi3.CodedError
	if errors.As(err, &coded) {
		return coded.Code()
	}
	return unknownValidationID
}
