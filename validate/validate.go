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
	findings = append(findings, lintTypeFormatMismatch(spec, source)...)
	findings = append(findings, lintSchemaConstraints(spec, source)...)
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
	line, column := lineColumnForKinError(err)
	f := formatters.Finding{
		Id:        knownRuleID(ruleIDForKinError(err)),
		Text:      unwrapContext(err).Error(),
		Level:     severityForKinError(err),
		Operation: operation,
		Path:      path,
		Section:   sectionForKinError(err),
		Source: formatters.Source{
			File:   source,
			Line:   line,
			Column: column,
		},
	}
	// Fingerprint last so it hashes over the populated fields.
	f.Fingerprint = checker.ComputeFingerprint(f.Id, f.Operation, f.Path, argsForKinError(err))
	return formatters.Findings{f}
}

// severityForKinError classifies a kin validation error into a severity by
// its rule id, so a finding carries exactly the severity `oasdiff checks
// validate` lists for that rule. See ruleLevels for the classifications.
func severityForKinError(err error) checker.Level {
	return RuleLevel(knownRuleID(ruleIDForKinError(err)))
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
