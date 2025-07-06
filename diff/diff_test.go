package diff_test

import (
	"fmt"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/oasdiff/oasdiff/utils"
	"github.com/stretchr/testify/require"
)

const (
	securityScorePath  = "/api/{domain}/{project}/badges/security-score"
	installCommandPath = "/api/{domain}/{project}/install-command"
)

func l(t *testing.T, v int) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	oas, err := loader.LoadFromFile(fmt.Sprintf("../data/openapi-test%d.yaml", v))
	require.NoError(t, err)
	return oas
}

func d(t *testing.T, config *diff.Config, v1, v2 int) *diff.Diff {
	t.Helper()
	d, err := diff.Get(config, l(t, v1), l(t, v2))
	require.NoError(t, err)
	return d
}

func TestDiff_Same(t *testing.T) {
	require.Nil(t, d(t, diff.NewConfig(), 1, 1))
}

func TestDiff_Empty(t *testing.T) {
	require.True(t, (*diff.CallbacksDiff)(nil).Empty())
	require.True(t, (*diff.EncodingsDiff)(nil).Empty())
	require.True(t, (*diff.ExtensionsDiff)(nil).Empty())
	require.True(t, (*diff.HeadersDiff)(nil).Empty())
	require.True(t, (*diff.OperationsDiff)(nil).Empty())
	require.True(t, (*diff.ParametersDiffByLocation)(nil).Empty())
	require.True(t, (*diff.RequestBodiesDiff)(nil).Empty())
	require.True(t, (*diff.ResponsesDiff)(nil).Empty())
	require.True(t, (*diff.SchemasDiff)(nil).Empty())
	require.True(t, (*diff.ServersDiff)(nil).Empty())
	require.True(t, (*diff.StringsDiff)(nil).Empty())
	require.True(t, (*diff.StringMapDiff)(nil).Empty())
	require.True(t, (*diff.TagsDiff)(nil).Empty())
	require.True(t, (*diff.SecurityRequirementsDiff)(nil).Empty())
	require.True(t, (*diff.SecuritySchemesDiff)(nil).Empty())
	require.True(t, (*diff.ExamplesDiff)(nil).Empty())
}

func TestDiff_DeletedPaths(t *testing.T) {
	require.ElementsMatch(t,
		[]string{installCommandPath, "/register", "/subscribe"},
		d(t, diff.NewConfig(), 1, 2).PathsDiff.Deleted)
}

func TestDiff_AddedOperation(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 1, 2).PathsDiff.Modified[securityScorePath].OperationsDiff.Added,
		"POST")
}

func TestDiff_DeletedOperation(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 2, 1).PathsDiff.Modified[securityScorePath].OperationsDiff.Deleted,
		"POST")
}

func TestDiff_ModifiedOperation(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile("../data/simple1.yaml")
	require.NoError(t, err)

	s2, err := loader.LoadFromFile("../data/simple2.yaml")
	require.NoError(t, err)

	d, err := diff.Get(diff.NewConfig(), s2, s1)
	require.NoError(t, err)

	require.Equal(t, &diff.OperationsDiff{
		Added:    utils.StringList{"GET"},
		Deleted:  utils.StringList{"POST"},
		Modified: diff.ModifiedOperations{},
	},
		d.PathsDiff.Modified["/api/test"].OperationsDiff)
}

func TestAddedExtension(t *testing.T) {
	require.Contains(t, d(t, diff.NewConfig(), 3, 1).ExtensionsDiff.Added, "x-extension-test")
}

func TestDeletedExtension(t *testing.T) {
	require.Contains(t, d(t, diff.NewConfig(), 1, 3).ExtensionsDiff.Deleted, "x-extension-test")
}

func TestModifiedExtension(t *testing.T) {
	require.NotNil(t, d(t, diff.NewConfig(), 1, 3).ExtensionsDiff.Modified["x-extension-test2"])
}

func TestExcludedExtension(t *testing.T) {
	require.Nil(t, d(t, diff.NewConfig().WithExcludeElements([]string{diff.ExcludeExtensionsOption}), 1, 3).ExtensionsDiff)
}

func TestDiff_AddedGlobalTag(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 3, 1).TagsDiff.Added,
		"security")
}

func TestDiff_ModifiedGlobalTag(t *testing.T) {
	require.Equal(t,
		&diff.ValueDiff{
			From: "Harrison",
			To:   "harrison",
		},
		d(t, diff.NewConfig(), 1, 3).TagsDiff.Modified["reuven"].DescriptionDiff)
}

func TestDiff_AddedTag(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 3, 1).PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].TagsDiff.Added,
		"security")
}

func TestDiff_DeletedEncoding(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 1, 3).PathsDiff.Modified["/subscribe"].OperationsDiff.Modified["POST"].CallbacksDiff.Modified["myEvent"].Modified["hi"].OperationsDiff.Modified["POST"].RequestBodyDiff.ContentDiff.MediaTypeModified["application/json"].EncodingsDiff.Deleted,
		"historyMetadata")
}

func TestDiff_ModifiedEncodingHeaders(t *testing.T) {
	require.NotNil(t,
		d(t, diff.NewConfig(), 3, 1).PathsDiff.Modified["/subscribe"].OperationsDiff.Modified["POST"].CallbacksDiff.Modified["myEvent"].Modified["hi"].OperationsDiff.Modified["POST"].RequestBodyDiff.ContentDiff.MediaTypeModified["application/json"].EncodingsDiff.Modified["profileImage"].HeadersDiff,
		"profileImage")
}

func TestDiff_AddedParam(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 2, 1).PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ParametersDiff.Added[openapi3.ParameterInHeader],
		"X-Auth-Name")
}

func TestDiff_DeletedParam(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 1, 2).PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ParametersDiff.Deleted[openapi3.ParameterInHeader],
		"X-Auth-Name")
}

func TestDiff_ModifiedParam(t *testing.T) {
	require.Equal(t,
		&diff.ValueDiff{
			From: true,
			To:   (interface{})(nil),
		},
		d(t, diff.NewConfig(), 2, 1).PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ParametersDiff.Modified[openapi3.ParameterInQuery]["image"].ExplodeDiff)
}

func TestSchemaDiff_TypeDiff(t *testing.T) {
	dd := d(t, diff.NewConfig(), 1, 2)

	require.True(t,
		dd.PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ParametersDiff.Modified[openapi3.ParameterInPath]["domain"].SchemaDiff.TypeDiff.Deleted.Is("string"))
	require.True(t,
		dd.PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ParametersDiff.Modified[openapi3.ParameterInPath]["domain"].SchemaDiff.TypeDiff.Added.Is("integer"))
}

func TestSchemaDiff_EnumDiff(t *testing.T) {
	require.Equal(t,
		&diff.EnumDiff{
			Added:   diff.EnumValues{"test1"},
			Deleted: diff.EnumValues{},
		},
		d(t, diff.NewConfig(), 1, 3).PathsDiff.Modified[installCommandPath].OperationsDiff.Modified["GET"].ParametersDiff.Modified[openapi3.ParameterInPath]["project"].SchemaDiff.EnumDiff)
}

func TestSchemaDiff_RequiredAdded(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 1, 5).PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ParametersDiff.Modified[openapi3.ParameterInQuery]["filter"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.RequiredDiff.Added,
		"type")
}

func TestSchemaDiff_RequiredDeleted(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 5, 1).PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ParametersDiff.Modified[openapi3.ParameterInQuery]["filter"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.RequiredDiff.Deleted,
		"type")
}

func TestSchemaDiff_NotDiff(t *testing.T) {
	require.Equal(t,
		true,
		d(t, diff.NewConfig(), 1, 3).PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ParametersDiff.Modified[openapi3.ParameterInQuery]["image"].SchemaDiff.NotDiff.SchemaAdded)
}

func TestSchemaDiff_ContentDiff(t *testing.T) {
	dd := d(t, diff.NewConfig(), 2, 1)

	require.True(t,
		dd.PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ParametersDiff.Modified[openapi3.ParameterInQuery]["filter"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.PropertiesDiff.Modified["color"].TypeDiff.Deleted.Is("number"))

	require.True(t,
		dd.PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ParametersDiff.Modified[openapi3.ParameterInQuery]["filter"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.PropertiesDiff.Modified["color"].TypeDiff.Added.Is("string"))
}

func TestSchemaDiff_MediaTypeAdded(t *testing.T) {
	require.Equal(t,
		utils.StringList([]string{"application/json"}),
		d(t, diff.NewConfig(), 5, 1).PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ParametersDiff.Modified[openapi3.ParameterInHeader]["user"].ContentDiff.MediaTypeAdded)
}

func TestSchemaDiff_MediaTypeDeleted(t *testing.T) {
	require.Equal(t,
		utils.StringList([]string{"application/json"}),
		d(t, diff.NewConfig(), 1, 5).PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ParametersDiff.Modified[openapi3.ParameterInHeader]["user"].ContentDiff.MediaTypeDeleted)
}

func TestSchemaDiff_MediaTypeModified(t *testing.T) {
	dd := d(t, diff.NewConfig(), 1, 5)

	require.True(t,
		dd.PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ParametersDiff.Modified[openapi3.ParameterInCookie]["test"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.TypeDiff.Deleted.Is("object"))

	require.True(t,
		dd.PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ParametersDiff.Modified[openapi3.ParameterInCookie]["test"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.TypeDiff.Added.Is("string"))
}

func TestSchemaDiff_MediaType_MultiEntries(t *testing.T) {

	s5 := l(t, 5)
	s5.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/json"] = openapi3.NewMediaType()
	s5.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["text/plain"] = openapi3.NewMediaType()

	s1 := l(t, 1)
	s1.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/json"] = openapi3.NewMediaType()
	s1.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["text/plain"] = openapi3.NewMediaType()

	_, err := diff.Get(diff.NewConfig(), s5, s1)

	require.NoError(t, err)
}

func TestSchemaDiff_AnyOfModified(t *testing.T) {
	require.False(t, d(t, &diff.Config{PathPrefixRevision: "/prefix"}, 4, 2).PathsDiff.Modified["/prefix"+securityScorePath].OperationsDiff.Modified["GET"].ParametersDiff.Modified[openapi3.ParameterInQuery]["token"].SchemaDiff.AnyOfDiff.Empty())
}

func TestSchemaDiff_WithExamples(t *testing.T) {

	require.Equal(t,
		&diff.ValueDiff{
			From: "26734565-dbcc-449a-a370-0beaaf04b0e8",
			To:   "26734565-dbcc-449a-a370-0beaaf04b0e7",
		},
		d(t, diff.NewConfig(), 1, 3).PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ParametersDiff.Modified["query"]["token"].SchemaDiff.ExampleDiff)
}

func TestSchemaDiff_MinDiff(t *testing.T) {

	dd := d(t, &diff.Config{PathPrefixRevision: "/prefix"}, 4, 2)
	path := "/prefix" + securityScorePath

	require.Nil(t,
		dd.PathsDiff.Modified[path].OperationsDiff.Modified["GET"].ParametersDiff.Modified[openapi3.ParameterInPath]["domain"].SchemaDiff.MinDiff.From)

	require.Equal(t,
		float64(7),
		dd.PathsDiff.Modified[path].OperationsDiff.Modified["GET"].ParametersDiff.Modified[openapi3.ParameterInPath]["domain"].SchemaDiff.MinDiff.To)
}

func TestResponseAdded(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 1, 3).PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ResponsesDiff.Added,
		"default")
}

func TestResponseDeleted(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 3, 1).PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ResponsesDiff.Deleted,
		"default")
}

func TestResponseDescriptionModified(t *testing.T) {

	require.Equal(t,
		&diff.ValueDiff{
			From: "Tufin",
			To:   "Tufin1",
		},
		d(t, diff.NewConfig(), 3, 1).PathsDiff.Modified[installCommandPath].OperationsDiff.Modified["GET"].ResponsesDiff.Modified["default"].DescriptionDiff)
}

func TestResponseHeadersModified(t *testing.T) {
	require.Equal(t,
		&diff.ValueDiff{
			From: "Request limit per min.",
			To:   "Request limit per hour.",
		},
		d(t, diff.NewConfig(), 3, 1).PathsDiff.Modified[installCommandPath].OperationsDiff.Modified["GET"].ResponsesDiff.Modified["default"].HeadersDiff.Modified["X-RateLimit-Limit"].DescriptionDiff)
}

func TestServerAdded(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 5, 3).PathsDiff.Modified[installCommandPath].OperationsDiff.Modified["GET"].ServersDiff.Added,
		"https://api.oasdiff.com")
}

func TestServerDeleted(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 3, 5).PathsDiff.Modified[installCommandPath].OperationsDiff.Modified["GET"].ServersDiff.Deleted,
		"https://api.oasdiff.com")
}

func TestServerModified(t *testing.T) {

	require.Contains(t,
		d(t, diff.NewConfig(), 5, 3).PathsDiff.Modified[installCommandPath].OperationsDiff.Modified["GET"].ServersDiff.Modified,
		"https://www.oasdiff.com")
}

func TestServerVariableAdded(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 3, 5).PathsDiff.Modified[installCommandPath].OperationsDiff.Modified["GET"].ServersDiff.Modified["https://www.oasdiff.com"].VariablesDiff.Added,
		"name")
}

func TestServerVariableModified(t *testing.T) {

	dd := d(t, diff.NewConfig(), 3, 5)

	require.Equal(t,
		"CEO",
		dd.PathsDiff.Modified[installCommandPath].OperationsDiff.Modified["GET"].ServersDiff.Modified["https://www.oasdiff.com"].VariablesDiff.Modified["title"].DefaultDiff.From)

	require.Equal(t,
		"developer",
		dd.PathsDiff.Modified[installCommandPath].OperationsDiff.Modified["GET"].ServersDiff.Modified["https://www.oasdiff.com"].VariablesDiff.Modified["title"].DefaultDiff.To)
}

func TestServerAddedToPathItem(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 5, 3).PathsDiff.Modified[installCommandPath].ServersDiff.Added,
		"https://api.oasdiff.com")
}

func TestParamAddedToPathItem(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 5, 3).PathsDiff.Modified[installCommandPath].ParametersDiff.Added[openapi3.ParameterInHeader],
		"name")
}

func TestParamDeletedFromPathItem(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 1, 2).PathsDiff.Modified[securityScorePath].ParametersDiff.Deleted[openapi3.ParameterInPath],
		"domain")
}

func TestHeaderAdded(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 5, 1).ComponentsDiff.HeadersDiff.Added,
		"new")
}

func TestHeaderDeleted(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 1, 5).ComponentsDiff.HeadersDiff.Deleted,
		"new")
}

func TestRequestBodyModified(t *testing.T) {
	dd := d(t, diff.NewConfig(), 1, 3)

	require.True(t,
		dd.ComponentsDiff.RequestBodiesDiff.Modified["reuven"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.PropertiesDiff.Modified["meter_value"].TypeDiff.Deleted.Is("number"),
	)

	require.True(t,
		dd.ComponentsDiff.RequestBodiesDiff.Modified["reuven"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.PropertiesDiff.Modified["meter_value"].TypeDiff.Added.Is("integer"),
	)
}

func TestHeaderModifiedSchema(t *testing.T) {
	dd := d(t, diff.NewConfig(), 5, 1)

	require.Equal(t,
		false,
		dd.ComponentsDiff.HeadersDiff.Modified["test"].SchemaDiff.AdditionalPropertiesAllowedDiff.From)

	require.Equal(t,
		true,
		dd.ComponentsDiff.HeadersDiff.Modified["test"].SchemaDiff.AdditionalPropertiesAllowedDiff.To)
}

func TestHeaderModifiedContent(t *testing.T) {
	dd := d(t, diff.NewConfig(), 5, 1)

	require.True(t,
		dd.ComponentsDiff.HeadersDiff.Modified["testc"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.TypeDiff.Deleted.Is("string"))

	require.True(t,
		dd.ComponentsDiff.HeadersDiff.Modified["testc"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.TypeDiff.Added.Is("object"))
}

func TestResponseContentModified(t *testing.T) {
	dd := d(t, diff.NewConfig(), 5, 1)

	require.True(t,
		dd.PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ResponsesDiff.Modified["201"].ContentDiff.MediaTypeModified["application/xml"].SchemaDiff.TypeDiff.Deleted.Is("object"))

	require.True(t,
		dd.PathsDiff.Modified[securityScorePath].OperationsDiff.Modified["GET"].ResponsesDiff.Modified["201"].ContentDiff.MediaTypeModified["application/xml"].SchemaDiff.TypeDiff.Added.Is("string"))
}

func TestResponseDespcriptionNil(t *testing.T) {
	s3 := l(t, 3)
	s3.Paths.Value(installCommandPath).Get.Responses.Value("default").Value.Description = nil

	d, err := diff.Get(diff.NewConfig(), s3, l(t, 1))
	require.NoError(t, err)

	require.Equal(t,
		&diff.ValueDiff{
			From: interface{}(nil),
			To:   "Tufin1",
		},
		d.PathsDiff.Modified[installCommandPath].OperationsDiff.Modified["GET"].ResponsesDiff.Modified["default"].DescriptionDiff)
}

func TestSchemaDiff_DeletedCallback(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 3, 1).PathsDiff.Modified["/register"].OperationsDiff.Modified["POST"].CallbacksDiff.Deleted,
		"myEvent")
}

func TestSchemaDiff_ModifiedCallback(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 3, 1).PathsDiff.Modified["/subscribe"].OperationsDiff.Modified["POST"].CallbacksDiff.Modified["myEvent"].Deleted,
		"{$request.body#/callbackUrl}")
}

func TestSchemaDiff_AddedRequestBody(t *testing.T) {
	require.True(t,
		d(t, diff.NewConfig(), 3, 1).PathsDiff.Modified["/subscribe"].OperationsDiff.Modified["POST"].CallbacksDiff.Modified["myEvent"].Modified["bye"].OperationsDiff.Modified["POST"].RequestBodyDiff.Added)
}

func TestSchemaDiff_AddedSchemas(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 1, 5).ComponentsDiff.SchemasDiff.Added,
		"requests")
}

func TestSchemaDiff_DeletedSchemas(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 5, 1).ComponentsDiff.SchemasDiff.Deleted,
		"requests")
}

func TestSchemaDiff_ModifiedSchemas(t *testing.T) {
	dd := d(t, diff.NewConfig(), 1, 5)

	require.Equal(t,
		true,
		dd.ComponentsDiff.SchemasDiff.Modified["network-policies"].AdditionalPropertiesAllowedDiff.From)

	require.Equal(t,
		false,
		dd.ComponentsDiff.SchemasDiff.Modified["network-policies"].AdditionalPropertiesAllowedDiff.To)
}

func TestSchemaDiff_ModifiedSchemasOldNil(t *testing.T) {
	dd := d(t, diff.NewConfig(), 1, 5)

	require.Equal(t,
		nil,
		dd.ComponentsDiff.SchemasDiff.Modified["rules"].AdditionalPropertiesAllowedDiff.From)

	require.Equal(t,
		false,
		dd.ComponentsDiff.SchemasDiff.Modified["rules"].AdditionalPropertiesAllowedDiff.To)
}

func TestSchemaDiff_ModifiedSchemasNewNil(t *testing.T) {
	dd := d(t, diff.NewConfig(), 5, 1)

	require.Equal(t,
		false,
		dd.ComponentsDiff.SchemasDiff.Modified["rules"].AdditionalPropertiesAllowedDiff.From)

	require.Nil(t,
		dd.ComponentsDiff.SchemasDiff.Modified["rules"].AdditionalPropertiesAllowedDiff.To)
}

func TestSummary(t *testing.T) {

	dd := d(t, diff.NewConfig(), 1, 2).GetSummary()

	require.Equal(t, diff.SummaryDetails{0, 3, 1}, dd.GetSummaryDetails(diff.PathsDetail))
	require.Equal(t, diff.SummaryDetails{0, 1, 0}, dd.GetSummaryDetails(diff.SecurityDetail))
	require.Equal(t, diff.SummaryDetails{0, 1, 0}, dd.GetSummaryDetails(diff.ServersDetail))
	require.Equal(t, diff.SummaryDetails{0, 2, 0}, dd.GetSummaryDetails(diff.TagsDetail))
	require.Equal(t, diff.SummaryDetails{0, 2, 0}, dd.GetSummaryDetails(diff.SchemasDetail))
	require.Equal(t, diff.SummaryDetails{0, 1, 0}, dd.GetSummaryDetails(diff.ParametersDetail))
	require.Equal(t, diff.SummaryDetails{0, 3, 0}, dd.GetSummaryDetails(diff.HeadersDetail))
	require.Equal(t, diff.SummaryDetails{0, 1, 0}, dd.GetSummaryDetails(diff.RequestBodiesDetail))
	require.Equal(t, diff.SummaryDetails{0, 1, 0}, dd.GetSummaryDetails(diff.ResponsesDetail))
	require.Equal(t, diff.SummaryDetails{0, 3, 0}, dd.GetSummaryDetails(diff.SecuritySchemesDetail))
	require.Equal(t, diff.SummaryDetails{}, dd.GetSummaryDetails(diff.ExamplesDetail))
	require.Equal(t, diff.SummaryDetails{}, dd.GetSummaryDetails(diff.LinksDetail))
	require.Equal(t, diff.SummaryDetails{}, dd.GetSummaryDetails(diff.CallbacksDetail))
}

func TestSummary2(t *testing.T) {

	dd := d(t, diff.NewConfig(), 1, 3).GetSummary()

	require.Equal(t, diff.SummaryDetails{0, 0, 4}, dd.GetSummaryDetails(diff.PathsDetail))
	require.Equal(t, diff.SummaryDetails{0, 1, 0}, dd.GetSummaryDetails(diff.SecurityDetail))
	require.Equal(t, diff.SummaryDetails{0, 1, 0}, dd.GetSummaryDetails(diff.ServersDetail))
	require.Equal(t, diff.SummaryDetails{0, 1, 1}, dd.GetSummaryDetails(diff.TagsDetail))
	require.Equal(t, diff.SummaryDetails{0, 2, 0}, dd.GetSummaryDetails(diff.SchemasDetail))
	require.Equal(t, diff.SummaryDetails{0, 1, 0}, dd.GetSummaryDetails(diff.ParametersDetail))
	require.Equal(t, diff.SummaryDetails{0, 3, 0}, dd.GetSummaryDetails(diff.HeadersDetail))
	require.Equal(t, diff.SummaryDetails{0, 0, 1}, dd.GetSummaryDetails(diff.RequestBodiesDetail))
	require.Equal(t, diff.SummaryDetails{0, 1, 0}, dd.GetSummaryDetails(diff.ResponsesDetail))
	require.Equal(t, diff.SummaryDetails{0, 2, 1}, dd.GetSummaryDetails(diff.SecuritySchemesDetail))
	require.Equal(t, diff.SummaryDetails{1, 0, 0}, dd.GetSummaryDetails(diff.ExamplesDetail))
	require.Equal(t, diff.SummaryDetails{1, 0, 0}, dd.GetSummaryDetails(diff.LinksDetail))
	require.Equal(t, diff.SummaryDetails{1, 0, 0}, dd.GetSummaryDetails(diff.CallbacksDetail))
}
func TestSummaryInvalidComponent(t *testing.T) {
	require.Equal(t, diff.SummaryDetails{
		Added:    0,
		Deleted:  0,
		Modified: 0,
	}, d(t, diff.NewConfig(), 1, 2).GetSummary().GetSummaryDetails("invalid"))
}

func TestFilterByRegex(t *testing.T) {
	d, err := diff.Get(&diff.Config{MatchPath: "x"}, l(t, 1), l(t, 2))
	require.NoError(t, err)
	require.Nil(t, d.GetSummary().Details[diff.PathsDetail])
}

func TestFilterByInvertedRegex(t *testing.T) {
	d, err := diff.Get(&diff.Config{UnmatchPath: "api|subscribe|register"}, l(t, 1), l(t, 2))
	require.NoError(t, err)
	require.Nil(t, d.GetSummary().Details[diff.PathsDetail])
}

func TestFilterPathsByExtension(t *testing.T) {
	d, err := diff.Get(&diff.Config{FilterExtension: "x-extension-test"}, l(t, 1), l(t, 2))
	require.NoError(t, err)
	require.Equal(t,
		&diff.SummaryDetails{Added: 1, Deleted: 3, Modified: 0},
		d.GetSummary().Details[diff.PathsDetail])
}

func TestFilterByInvalidRegex(t *testing.T) {
	_, err := diff.Get(&diff.Config{MatchPath: "["}, l(t, 1), l(t, 2))
	require.EqualError(t, err, "failed to compile filter regex \"[\": error parsing regexp: missing closing ]: `[`")
}

func TestFilterByInvalidInvertedRegex(t *testing.T) {
	_, err := diff.Get(&diff.Config{UnmatchPath: "["}, l(t, 1), l(t, 2))
	require.EqualError(t, err, "failed to compile filter regex \"[\": error parsing regexp: missing closing ]: `[`")
}

func TestFilterOperationssByExtension(t *testing.T) {
	d, err := diff.Get(&diff.Config{FilterExtension: "x-beta"}, l(t, 1), l(t, 3))
	require.NoError(t, err)
	require.Equal(t,
		&diff.SummaryDetails{Added: 0, Deleted: 0, Modified: 3},
		d.GetSummary().Details[diff.EndpointsDetail])
}

func TestAddedSecurityRequirement(t *testing.T) {
	require.Contains(t,
		d(t, diff.NewConfig(), 3, 1).PathsDiff.Modified["/register"].OperationsDiff.Modified["POST"].SecurityDiff.Added,
		"bearerAuth")
}

func TestSecurityRequirementScopesDeleted(t *testing.T) {
	securityScopesDiff := d(t, diff.NewConfig(), 3, 1).PathsDiff.Modified["/register"].OperationsDiff.Modified["POST"].SecurityDiff.Modified["OAuth"]
	require.NotEmpty(t, securityScopesDiff)

	require.Contains(t,
		securityScopesDiff["OAuth"].Deleted,
		"write:pets")
}

func TestSecurityRequirementModified(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile("../data/security-requirements/spec_1.yaml")
	require.NoError(t, err)

	s2, err := loader.LoadFromFile("../data/security-requirements/spec_2.yaml")
	require.NoError(t, err)

	dd, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotEmpty(t, dd)
}

func TestAddedSecurityOAuthFlows(t *testing.T) {
	require.True(t,
		d(t, diff.NewConfig(), 1, 5).ComponentsDiff.SecuritySchemesDiff.Modified["AccessToken"].OAuthFlowsDiff.Added)
}

func TestOAS31(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile("../data/openapi31-test1.yaml")
	require.NoError(t, err)

	s2, err := loader.LoadFromFile("../data/openapi31-test2.yaml")
	require.NoError(t, err)

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	// while specific 3.1 features, such as webhooks, are not yet supported by kin-openapi, the diff still works
	require.Contains(t,
		d.ComponentsDiff.SchemasDiff.Modified["Pet"].RequiredDiff.Added,
		"tag")
}

func TestCircularSchema_Diff(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile("../data/circular1.yaml")
	require.NoError(t, err)

	s2, err := loader.LoadFromFile("../data/circular2.yaml")
	require.NoError(t, err)

	_, err = diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	// TODO: fix circular checks and re-enable this test
	// require.True(t, dd.SchemasDiff.Modified["circular1"].PropertiesDiff.Modified["children"].ItemsDiff.CircularRefDiff)
}

func TestCircularSchemaRefs(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile("../data/circular1.yaml")
	require.NoError(t, err)

	s2, err := loader.LoadFromFile("../data/circular2.yaml")
	require.NoError(t, err)

	dd, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	require.Contains(t, dd.ComponentsDiff.SchemasDiff.Modified, "circular1")
	require.NotContains(t, dd.ComponentsDiff.SchemasDiff.Modified, "circular2")
	require.Contains(t, dd.ComponentsDiff.SchemasDiff.Modified, "circular3")
	require.NotContains(t, dd.ComponentsDiff.SchemasDiff.Modified, "circular4")
	require.Contains(t, dd.ComponentsDiff.SchemasDiff.Modified, "circular5")
	require.Contains(t, dd.ComponentsDiff.SchemasDiff.Modified, "circular6")
}

func TestCallbacks(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile("../data/callbacks/spec_1.yaml")
	require.NoError(t, err)

	s2, err := loader.LoadFromFile("../data/callbacks/spec_2.yaml")
	require.NoError(t, err)

	_, err = diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
}

func TestDiff_InfoNil(t *testing.T) {
	s1 := &openapi3.T{}
	d, err := diff.Get(diff.NewConfig(), s1, s1)
	require.NoError(t, err)
	require.Empty(t, d)
}

func TestDiff_InfoAdded(t *testing.T) {
	s1 := &openapi3.T{}
	s2 := &openapi3.T{
		Info: &openapi3.Info{},
	}
	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.True(t, d.InfoDiff.Added)
}

func TestDiff_InfoDeleted(t *testing.T) {
	s1 := &openapi3.T{
		Info: &openapi3.Info{},
	}
	s2 := &openapi3.T{}
	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.True(t, d.InfoDiff.Deleted)
}

func TestDiff_PathParamInMethodRenamed(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/param-rename/method-base.yaml"))
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/param-rename/method-revision.yaml"))
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	dd := d.PathsDiff.Modified["/books/{bookId}"].ParametersDiff.Modified["path"]["bookId"].NameDiff
	require.Equal(t, "bookId", dd.From)
	require.Equal(t, "id", dd.To)
}

func TestDiff_PathParamInOperationRenamed(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/param-rename/op-base.yaml"))
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/param-rename/op-revision.yaml"))
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	dd := d.PathsDiff.Modified["/books/{bookId}"].OperationsDiff.Modified["GET"].ParametersDiff.Modified["path"]["bookId"].NameDiff
	require.Equal(t, "bookId", dd.From)
	require.Equal(t, "id", dd.To)
}

func TestDiff_PathParamRefInOperationRenamed(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/param-rename/op-base.yaml"))
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/param-rename/op-revision-ref.yaml"))
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	dd := d.PathsDiff.Modified["/books/{bookId}"].OperationsDiff.Modified["GET"].ParametersDiff.Modified["path"]["bookId"].NameDiff
	require.Equal(t, "bookId", dd.From)
	require.Equal(t, "id", dd.To)
}

func TestDiff_TwoPathParamsRenamed(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/param-rename/two-base.yaml"))
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/param-rename/two-revision.yaml"))
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	dd := d.PathsDiff.Modified["/books/{bookId}/{libraryId}"].OperationsDiff.Modified["GET"].ParametersDiff.Modified["path"]["bookId"].NameDiff
	require.Equal(t, "bookId", dd.From)
	require.Equal(t, "id", dd.To)

	dd = d.PathsDiff.Modified["/books/{bookId}/{libraryId}"].OperationsDiff.Modified["GET"].ParametersDiff.Modified["path"]["libraryId"].NameDiff
	require.Equal(t, "libraryId", dd.From)
	require.Equal(t, "otherId", dd.To)
}

func TestDiff_TwoPathParamsOneRenamed(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/param-rename/one-of-two-base.yaml"))
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/param-rename/one-of-two-revision.yaml"))
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	dd := d.PathsDiff.Modified["/books/{bookId}/{libraryId}"].OperationsDiff.Modified["GET"].ParametersDiff.Modified["path"]["libraryId"].NameDiff
	require.Equal(t, "libraryId", dd.From)
	require.Equal(t, "otherId", dd.To)
}

func TestDiff_DifferentComponentSameParam(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/different_component_same_parameter.yaml"))
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s1)
	require.NoError(t, err)
	require.Empty(t, d)
}

func TestDiff_DifferentComponentModifiedParam(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/different_component_same_parameter.yaml"))
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/different_component_modified_parameter.yaml"))
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	dd := d.ComponentsDiff.ParametersDiff.Modified["differentComponentName_A"].InDiff
	require.Equal(t, "header", dd.From)
	require.Equal(t, "query", dd.To)
}

func TestDiff_DifferentComponentSameSchema(t *testing.T) {
	s1, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/different_component_same_schema.yaml"))
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s1)
	require.NoError(t, err)
	require.Empty(t, d)
}

func TestDiff_DifferentComponentSameHeader(t *testing.T) {
	s1, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/different_component_same_header.yaml"))
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s1)
	require.NoError(t, err)
	require.Empty(t, d)
}

func TestDiff_Extensions(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/extensions/base.yaml"))
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/extensions/revision.yaml"))
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	dd := d.PathsDiff.Modified["/example/callback"].OperationsDiff.Modified["POST"].ExtensionsDiff.Modified["x-amazon-apigateway-integration"]
	require.Len(t, dd, 2)
	require.Equal(t, "200", dd[0].Value)
	require.Equal(t, "201", dd[0].OldValue)
	require.Equal(t, "http://api.example.com/v1/example/calllllllllback", dd[1].Value)
	require.Equal(t, "http://api.example.com/v1/example/callback", dd[1].OldValue)
}

func TestDiff_ExtensionsExcluded(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/extensions/base.yaml"))
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/extensions/revision.yaml"))
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig().WithExcludeElements([]string{diff.ExcludeExtensionsOption}), s1, s2)
	require.NoError(t, err)
	require.Empty(t, d)
}

func TestDiff_ExtensionsInvalid(t *testing.T) {
	s1, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/extensions/base.yaml"))
	require.NoError(t, err)

	// Add invalid extension
	newPathItem := s1.Spec.Paths.Find("/example/callback")
	newPathItem.Post.Extensions["x-amazon-apigateway-integration"] = interface{}(make(chan int))
	s1.Spec.Paths.Set("/example/callback", newPathItem)

	_, _, err = diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s1)
	require.EqualError(t, err, "json: unsupported type: chan int")
}

func TestDiff_ExtensionsInvalidSecuritySchemes(t *testing.T) {
	s1, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/security-requirements/spec_1.yaml"))
	require.NoError(t, err)

	// Add invalid extension
	s1.Spec.Components.SecuritySchemes["petstore_auth"].Value.Extensions = map[string]interface{}{"invalid": interface{}(make(chan int))}

	_, _, err = diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s1)
	require.EqualError(t, err, "json: unsupported type: chan int")
}

func TestDiff_ExtensionsInvalidFlows(t *testing.T) {
	s1, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/security-requirements/spec_1.yaml"))
	require.NoError(t, err)

	// Add invalid extension
	s1.Spec.Components.SecuritySchemes["petstore_auth"].Value.Flows.Extensions = map[string]interface{}{"invalid": interface{}(make(chan int))}

	_, _, err = diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s1)
	require.EqualError(t, err, "json: unsupported type: chan int")
}

func TestDiff_ExtensionsInvalidImplicit(t *testing.T) {
	s1, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/security-requirements/spec_1.yaml"))
	require.NoError(t, err)

	// Add invalid extension
	s1.Spec.Components.SecuritySchemes["petstore_auth"].Value.Flows.Implicit.Extensions = map[string]interface{}{"invalid": interface{}(make(chan int))}

	_, _, err = diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s1)
	require.EqualError(t, err, "json: unsupported type: chan int")
}

func TestDiff_ExtensionsIssue519(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/extensions/openapi-2.json"))
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s1)
	require.NoError(t, err)
	require.Empty(t, d)
}
