package report

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/oasdiff/oasdiff/diff"
)

type report struct {
	Writer io.Writer
	level  int
}

func (r *report) indent() *report {
	return &report{
		Writer: r.Writer,
		level:  r.level + 1,
	}
}

func (r *report) print(output ...interface{}) (n int, err error) {
	return fmt.Fprintln(r.Writer, addPrefix(r.level, output)...)
}

func addPrefix(level int, output []interface{}) []interface{} {
	return append(getPrefix(level), output...)
}

func getPrefix(level int) []interface{} {
	if level == 1 {
		return []interface{}{"-"}
	}

	if level > 1 {
		return []interface{}{strings.Repeat("  ", level-1) + "-"}
	}

	return []interface{}{}
}

// output prints the diff
// note that it may mutate diff by sorting its members
func (r *report) output(d *diff.Diff) {

	if d.Empty() {
		r.print("No changes")
		return
	}

	if d.EndpointsDiff.Empty() {
		r.print("No endpoint changes, but there are some other changes")
	} else {
		r.printEndpoints(d.EndpointsDiff)
	}

	if d.ExtensionsDiff.Empty() &&
		d.SecurityDiff.Empty() &&
		d.ServersDiff.Empty() {
		return
	}

	r.print("Other Changes")
	r.print("-------------")

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
		r.print("")
	}

	r.printValue(d.OpenAPIDiff, "Version")

	if !d.SecurityDiff.Empty() {
		r.print("Security Requirements changed")
		r.indent().printSecurityRequirements(d.SecurityDiff)
		r.print("")
	}

	if !d.ServersDiff.Empty() {
		r.print("Servers changed")
		r.indent().printServers(d.ServersDiff)
		r.print("")
	}
}

func (r *report) printEndpoints(d *diff.EndpointsDiff) {

	r.printTitle("New Endpoints", len(d.Added))
	sort.Sort(d.Added)
	for _, added := range d.Added {
		r.print(added.Method, added.Path, " ")
	}
	r.print("")

	r.printTitle("Deleted Endpoints", len(d.Deleted))
	sort.Sort(d.Deleted)
	for _, deleted := range d.Deleted {
		r.print(deleted.Method, deleted.Path, " ")
	}
	r.print("")

	r.printTitle("Modified Endpoints", len(d.Modified))
	keys := d.Modified.ToEndpoints()
	sort.Sort(keys)
	for _, endpoint := range keys {
		r.print(endpoint.Method, endpoint.Path)
		r.indent().printMethod(d.Modified[endpoint])
		r.print("")
	}
}

func (r *report) printServers(d *diff.ServersDiff) {
	if d.Empty() {
		return
	}

	sort.Sort(d.Added)
	for _, added := range d.Added {
		r.print("New server:", added)
	}

	sort.Sort(d.Deleted)
	for _, deleted := range d.Deleted {
		r.print("Deleted server:", deleted)
	}

	for _, server := range getKeys(d.Modified) {
		r.print("Modified server:", server)
		r.indent().printServer(d.Modified[server])
	}
}

func (r *report) printMethod(d *diff.MethodDiff) {
	if d.Empty() {
		return
	}

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	r.printStrings(d.TagsDiff, "Tags")
	r.printValue(d.SummaryDiff, "Summary")
	r.printValue(d.DescriptionDiff, "Description")
	r.printValue(d.OperationIDDiff, "OperationID")
	r.printParams(d.ParametersDiff)

	if !d.RequestBodyDiff.Empty() {
		r.print("Request body changed")
		r.indent().printRequestBody(d.RequestBodyDiff)
	}

	if !d.ResponsesDiff.Empty() {
		r.print("Responses changed")
		r.indent().printResponses(d.ResponsesDiff)
	}

	r.printMessage(d.CallbacksDiff, "Callbacks changed")
	r.printValue(d.DeprecatedDiff, "Deprecated")

	if !d.SecurityDiff.Empty() {
		r.print("Security changed")
		r.indent().printSecurityRequirements(d.SecurityDiff)
	}

	if !d.ServersDiff.Empty() {
		r.print("Servers changed")
		r.indent().printServers(d.ServersDiff)
	}
}

func (r *report) printParams(d *diff.ParametersDiffByLocation) {
	if d.Empty() {
		return
	}

	for _, location := range diff.ParamLocations {
		params := d.Added[location]
		sort.Strings(params)
		for _, param := range params {
			r.print("New", location, "param:", param)
		}
	}

	for _, location := range diff.ParamLocations {
		params := d.Deleted[location]
		sort.Strings(params)
		for _, param := range params {
			r.print("Deleted", location, "param:", param)
		}
	}

	for _, location := range diff.ParamLocations {
		paramDiffs := d.Modified[location]
		for _, param := range getKeys(paramDiffs) {
			r.print("Modified", location, "param:", param)
			r.indent().printParam(paramDiffs[param])
		}
	}
}

func (r *report) printParam(d *diff.ParameterDiff) {
	r.printValue(d.NameDiff, "Name")
	r.printValue(d.InDiff, "In")

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	r.printValue(d.DescriptionDiff, "Description")
	r.printValue(d.StyleDiff, "Style")
	r.printValue(d.ExplodeDiff, "Explode")
	r.printValue(d.AllowEmptyValueDiff, "AllowEmptyValue")
	r.printValue(d.AllowReservedDiff, "AllowReserved")
	r.printValue(d.DeprecatedDiff, "Deprecated")
	r.printValue(d.RequiredDiff, "Required")

	if !d.SchemaDiff.Empty() {
		r.print("Schema changed")
		r.indent().printSchema(d.SchemaDiff)
	}

	r.printValue(d.ExampleDiff, "Example")

	if !d.ExamplesDiff.Empty() {
		r.print("Examples changed")
		r.indent().printExamples(d.ExamplesDiff)
	}

	if !d.ContentDiff.Empty() {
		r.print("Content changed")
		r.indent().printContent(d.ContentDiff)
	}
}

func (r *report) printExamples(d *diff.ExamplesDiff) {
	if d.Empty() {
		return
	}

	sort.Sort(d.Added)
	for _, example := range d.Added {
		r.print("New example:", example)
	}

	sort.Sort(d.Deleted)
	for _, example := range d.Deleted {
		r.print("Deleted example:", example)
	}

	for _, example := range getKeys(d.Modified) {
		r.print("Modified example:", example)
		r.indent().printExample(d.Modified[example])
	}
}

func (r *report) printExample(d *diff.ExampleDiff) {
	if d.Empty() {
		return
	}

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	r.printValue(d.SummaryDiff, "Summary")
	r.printValue(d.DescriptionDiff, "Description")
	r.printValue(d.ValueDiff, "Value")
	r.printValue(d.ExternalValueDiff, "ExternalValue")
}

func (r *report) printRequiredProperties(d *diff.RequiredPropertiesDiff) {
	if d.Empty() {
		return
	}

	sort.Sort(d.Added)
	for _, added := range d.Added {
		r.print("New required property:", added)
	}

	sort.Sort(d.Deleted)
	for _, deleted := range d.Deleted {
		r.print("Deleted required property:", deleted)
	}
}

func (r *report) printServer(d *diff.ServerDiff) {
	if d.Empty() {
		return
	}

	r.printConditional(d.Added, "Server added")
	r.printConditional(d.Deleted, "Server deleted")

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	r.printValue(d.URLDiff, "URL")
	r.printValue(d.DescriptionDiff, "Description")
	if !d.VariablesDiff.Empty() {
		r.print("Variables changed")
		r.indent().printVariables(d.VariablesDiff)
	}
}

func (r *report) printVariables(d *diff.VariablesDiff) {
	if d.Empty() {
		return
	}

	sort.Sort(d.Added)
	for _, variable := range d.Added {
		r.print("New variable:", variable)
	}

	sort.Sort(d.Deleted)
	for _, variable := range d.Deleted {
		r.print("Deleted variable:", variable)
	}

	for _, variable := range getKeys(d.Modified) {
		r.print("Modified variable:", variable)
		r.indent().printVariable(d.Modified[variable])
	}
}

func (r *report) printVariable(d *diff.VariableDiff) {
	if d.Empty() {
		return
	}

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	if !d.EnumDiff.Empty() {
		r.printConditional(len(d.EnumDiff.Added) > 0, "New enum values:", d.EnumDiff.Added)
		r.printConditional(len(d.EnumDiff.Deleted) > 0, "Deleted enum values:", d.EnumDiff.Deleted)
	}
	r.printValue(d.DefaultDiff, "Default")
	r.printValue(d.DescriptionDiff, "Description")
}

func (r *report) printExtensions(d *diff.ExtensionsDiff) {
	if d.Empty() {
		return
	}

	sort.Sort(d.Added)
	for _, added := range d.Added {
		r.print("New extension:", added)
	}

	sort.Sort(d.Deleted)
	for _, deleted := range d.Deleted {
		r.print("Deleted extension:", deleted)
	}

	for extension, patch := range d.Modified {
		r.print("Modified extension:", extension)
		r.indent().printExtension(patch)
	}
}

func (r *report) printExtension(d diff.JsonPatch) {
	if d.Empty() {
		return
	}

	for _, op := range d {
		r.print(op.String())
	}
}

func (r *report) printSchema(d *diff.SchemaDiff) {
	if d.Empty() {
		return
	}

	r.printConditional(d.SchemaAdded, "Schema added")
	r.printConditional(d.SchemaDeleted, "Schema deleted")
	r.printConditional(d.CircularRefDiff, "Schema circular referecnce changed")

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	if !d.OneOfDiff.Empty() {
		r.print("Property 'OneOf' changed")
		r.indent().printSchemaListDiff(d.OneOfDiff)
	}
	if !d.AnyOfDiff.Empty() {
		r.print("Property 'AnyOf' changed")
		r.indent().printSchemaListDiff(d.AnyOfDiff)
	}
	if !d.AllOfDiff.Empty() {
		r.print("Property 'AllOf' changed")
		r.indent().printSchemaListDiff(d.AllOfDiff)
	}

	if !d.NotDiff.Empty() {
		r.print("Property 'Not' changed")
		r.indent().printSchema(d.NotDiff)
	}

	r.printStrings(d.TypeDiff, "Type")
	r.printValue(d.TitleDiff, "Title")
	r.printValue(d.FormatDiff, "Format")
	r.printValue(d.DescriptionDiff, "Description")

	if !d.EnumDiff.Empty() {
		r.printConditional(len(d.EnumDiff.Added) > 0, "New enum values:", d.EnumDiff.Added)
		r.printConditional(len(d.EnumDiff.Deleted) > 0, "Deleted enum values:", d.EnumDiff.Deleted)
	}

	r.printValue(d.DefaultDiff, "Default")
	r.printValue(d.ExampleDiff, "Example")
	r.printValue(d.AdditionalPropertiesAllowedDiff, "AdditionalProperties")
	r.printValue(d.UniqueItemsDiff, "UniqueItems")
	r.printValue(d.ExclusiveMinDiff, "ExclusiveMin")
	r.printValue(d.ExclusiveMaxDiff, "ExclusiveMax")
	r.printValue(d.NullableDiff, "Nullable")
	r.printValue(d.ReadOnlyDiff, "ReadOnly")
	r.printValue(d.WriteOnlyDiff, "WriteOnly")
	r.printValue(d.AllowEmptyValueDiff, "AllowEmptyValue")
	r.printValue(d.XMLDiff, "XML")
	r.printValue(d.DeprecatedDiff, "Deprecated")
	r.printValue(d.MinDiff, "Min")
	r.printValue(d.MaxDiff, "Max")
	r.printValue(d.MultipleOfDiff, "MultipleOf")
	r.printValue(d.MinLengthDiff, "MinLength")
	r.printValue(d.MaxLengthDiff, "MaxLength")
	r.printValue(d.PatternDiff, "Pattern")
	r.printValue(d.MinItemsDiff, "MinItems")
	r.printValue(d.MaxItemsDiff, "MaxItems")

	if !d.ItemsDiff.Empty() {
		r.print("Items changed")
		r.indent().printSchema(d.ItemsDiff)
	}

	if !d.RequiredDiff.Empty() {
		r.print("Required changed")
		r.indent().printRequiredProperties(d.RequiredDiff)
	}

	r.printValue(d.MinPropsDiff, "MinProps")
	r.printValue(d.MaxPropsDiff, "MaxProps")

	if !d.PropertiesDiff.Empty() {
		r.print("Properties changed")
		r.indent().printProperties(d.PropertiesDiff)
	}

	if !d.AdditionalPropertiesDiff.Empty() {
		r.print("AdditionalProperties changed")
		r.indent().printSchema(d.AdditionalPropertiesDiff)
	}

	r.printMessage(d.DiscriminatorDiff, "Discriminator changed")
}

func (r *report) printSchemaListDiff(d *diff.SubschemasDiff) {
	if d.Empty() {
		return
	}

	r.printConditional(len(d.Added) > 0, "Schemas added:", d.Added)
	r.printConditional(len(d.Deleted) > 0, "Schemas deleted:", d.Deleted)

	if len(d.Modified) > 0 {
		for _, schemaDiff := range d.Modified {
			r.print("Modified schema:", schemaDiff.String())
			r.indent().printSchema(schemaDiff.Diff)
		}
	}
}

func (r *report) printProperties(d *diff.SchemasDiff) {
	if d.Empty() {
		return
	}

	sort.Sort(d.Added)
	for _, property := range d.Added {
		r.print("New property:", property)
	}

	sort.Sort(d.Deleted)
	for _, property := range d.Deleted {
		r.print("Deleted property:", property)
	}

	for _, property := range getKeys(d.Modified) {
		r.print("Modified property:", property)
		r.indent().printSchema(d.Modified[property])
	}
}

func quote(value interface{}) interface{} {
	if value == nil {
		return "null"
	}
	if reflect.ValueOf(value).Kind() == reflect.String {
		return "'" + value.(string) + "'"
	}
	return value
}

func (r *report) printResponses(d *diff.ResponsesDiff) {
	if d.Empty() {
		return
	}

	sort.Sort(d.Added)
	for _, added := range d.Added {
		r.print("New response:", added)
	}

	sort.Sort(d.Deleted)
	for _, deleted := range d.Deleted {
		r.print("Deleted response:", deleted)
	}

	for _, response := range getKeys(d.Modified) {
		r.print("Modified response:", response)
		r.indent().printResponse(d.Modified[response])
	}
}

func (r *report) printResponse(d *diff.ResponseDiff) {
	if d.Empty() {
		return
	}

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	r.printValue(d.DescriptionDiff, "Description")

	if !d.ContentDiff.Empty() {
		r.print("Content changed")
		r.indent().printContent(d.ContentDiff)
	}

	if !d.HeadersDiff.Empty() {
		r.print("Headers changed")
		r.indent().printHeaders(d.HeadersDiff)
	}
}

func (r *report) printRequestBody(d *diff.RequestBodyDiff) {
	if d.Empty() {
		return
	}

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	r.printValue(d.DescriptionDiff, "Description")

	if !d.ContentDiff.Empty() {
		r.print("Content changed")
		r.indent().printContent(d.ContentDiff)
	}
}

func (r *report) printContent(d *diff.ContentDiff) {
	if d.Empty() {
		return
	}

	sort.Sort(d.MediaTypeAdded)
	for _, name := range d.MediaTypeAdded {
		r.print("New media type:", name)
	}

	sort.Sort(d.MediaTypeDeleted)
	for _, name := range d.MediaTypeDeleted {
		r.print("Deleted media type:", name)
	}

	for _, name := range getKeys(d.MediaTypeModified) {
		r.print("Modified media type:", name)
		r.indent().printMediaType(d.MediaTypeModified[name])
	}
}

func (r *report) printMediaType(d *diff.MediaTypeDiff) {
	if d.Empty() {
		return
	}

	if !d.NameDiff.Empty() {
		r.printValue(d.NameDiff.NameDiff, "Name")
	}

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	if !d.SchemaDiff.Empty() {
		r.print("Schema changed")
		r.indent().printSchema(d.SchemaDiff)
	}

	r.printValue(d.ExampleDiff, "Example")

	if !d.ExamplesDiff.Empty() {
		r.print("Examples changed")
		r.indent().printExamples(d.ExamplesDiff)
	}

	r.printMessage(d.EncodingsDiff, "Encodings changed")
}

func (r *report) printValue(d *diff.ValueDiff, title string) {
	if d.Empty() {
		return
	}

	r.print(title, "changed from", quote(d.From), "to", quote(d.To))
}

func (r *report) printStrings(d *diff.StringsDiff, title string) {
	if d.Empty() {
		return
	}

	r.print(title, "changed from", quote(d.Deleted.String()), "to", quote(d.Added.String()))
}

func (r *report) printHeaders(d *diff.HeadersDiff) {
	if d.Empty() {
		return
	}

	sort.Sort(d.Added)
	for _, added := range d.Added {
		r.print("New header:", added)
	}

	sort.Sort(d.Deleted)
	for _, deleted := range d.Deleted {
		r.print("Deleted header:", deleted)
	}

	for _, header := range getKeys(d.Modified) {
		r.print("Modified header:", header)
		r.indent().printHeader(d.Modified[header])
	}
}

func (r *report) printHeader(d *diff.HeaderDiff) {
	if d.Empty() {
		return
	}

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	r.printValue(d.DescriptionDiff, "Description")
	r.printValue(d.DeprecatedDiff, "Deprecated")
	r.printValue(d.RequiredDiff, "Required")

	r.printValue(d.ExampleDiff, "Example")

	if !d.ExamplesDiff.Empty() {
		r.print("Examples changed")
		r.indent().printExamples(d.ExamplesDiff)
	}

	if !d.SchemaDiff.Empty() {
		r.print("Schema changed")
		r.indent().printSchema(d.SchemaDiff)
	}

	if !d.ContentDiff.Empty() {
		r.print("Content changed")
		r.indent().printContent(d.ContentDiff)
	}
}

func (r *report) printSecurityRequirements(d *diff.SecurityRequirementsDiff) {
	if d.Empty() {
		return
	}

	sort.Sort(d.Added)
	for _, added := range d.Added {
		r.print("New security requirements:", added)
	}

	sort.Sort(d.Deleted)
	for _, deleted := range d.Deleted {
		r.print("Deleted security requirements:", deleted)
	}

	for _, securityRequirementID := range getKeys(d.Modified) {
		r.print("Modified security requirements:", securityRequirementID)
		r.indent().printSecurityScopes(d.Modified[securityRequirementID])
	}
}

func (r *report) printSecurityScopes(d diff.SecurityScopesDiff) {
	for _, scheme := range getKeys(d) {
		scopeDiff := d[scheme]
		r.printConditional(len(scopeDiff.Added) > 0, "Scheme", scheme, "Added scopes:", scopeDiff.Added)
		r.printConditional(len(scopeDiff.Deleted) > 0, "Scheme", scheme, "Deleted scopes:", scopeDiff.Deleted)
	}
}

func (r *report) printTitle(title string, count int) {
	text := ""
	if count == 0 {
		text = fmt.Sprintf("### %s: None", title)
	} else {
		text = fmt.Sprintf("### %s: %d", title, count)
	}

	r.print(text)
	r.print(strings.Repeat("-", len(text)))
}

func (r *report) printMessage(d diff.IDiff, output ...interface{}) {
	r.printConditional(!d.Empty(), output...)
}

func (r *report) printConditional(b bool, output ...interface{}) {
	if b {
		r.print(output...)
	}
}
