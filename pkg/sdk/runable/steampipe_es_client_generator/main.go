package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/opengovern/og-describer-template/provider/configs"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"html/template"
	"os"
	"regexp"
	"strings"
)

var (
	file              = flag.String("file", "", "Location of the model file")
	output            = flag.String("output", "", "Location of the output file")
	resourceTypesFile = flag.String("resourceTypesFile", "", "Location of the resource types json file file")
	pluginPath        = flag.String("pluginPath", "", "Location of the steampipe plugin")
)

const PluginPath = "" // TODO: give the steampipe plugin path

type IntegrationType struct {
	Name            string
	Index           string
	IntegrationType string
	ListFilters     map[string]string
	GetFilters      map[string]string
}

type ResourceType struct {
	ResourceName   string
	ListDescriber  string
	GetDescriber   string
	SteampipeTable string
	Model          string
}

func main() {
	if output == nil || len(*output) == 0 {
		v := "../../es/resources_clients.go"
		output = &v
	}
	if file == nil || len(*file) == 0 {
		v := "../../../../provider/model/model.go"
		file = &v
	}

	if pluginPath == nil || len(*pluginPath) == 0 {
		pp := PluginPath
		pluginPath = &pp
	}

	if resourceTypesFile == nil || len(*resourceTypesFile) == 0 {
		rt := "../../../../provider/resource_types/resource-types.json"
		resourceTypesFile = &rt
	}

	b, err := os.ReadFile(*resourceTypesFile)
	if err != nil {
		panic(err)
	}
	var resourceTypes []ResourceType
	err = json.Unmarshal(b, &resourceTypes)
	if err != nil {
		panic(err)
	}

	flag.CommandLine.Init("gen", flag.ExitOnError)
	flag.Parse()

	tpl := template.New("types")
	_, err = tpl.Parse(`
// ==========================  START: {{ .Name }} =============================

type {{ .Name }} struct {
	ResourceID string ` + "`json:\"resource_id\"`" + `
	PlatformID string ` + "`json:\"platform_id\"`" + `
	Description   {{ .IntegrationType }}.{{ .Name }}Description 	` + "`json:\"description\"`" + `
	Metadata      {{ .IntegrationType }}.Metadata 					` + "`json:\"metadata\"`" + `
	DescribedBy 	   string ` + "`json:\"described_by\"`" + `
	ResourceType       string ` + "`json:\"resource_type\"`" + `
	IntegrationType    string ` + "`json:\"integration_type\"`" + `
	IntegrationID      string ` + "`json:\"integration_id\"`" + `
}

type {{ .Name }}Hit struct {
	ID      string            ` + "`json:\"_id\"`" + `
	Score   float64           ` + "`json:\"_score\"`" + `
	Index   string            ` + "`json:\"_index\"`" + `
	Type    string            ` + "`json:\"_type\"`" + `
	Version int64             ` + "`json:\"_version,omitempty\"`" + `
	Source  {{ .Name }}       ` + "`json:\"_source\"`" + `
	Sort    []interface{}     ` + "`json:\"sort\"`" + `
}

type {{ .Name }}Hits struct {
	Total essdk.SearchTotal       ` + "`json:\"total\"`" + `
	Hits  []{{ .Name }}Hit ` + "`json:\"hits\"`" + `
}

type {{ .Name }}SearchResponse struct {
	PitID string          ` + "`json:\"pit_id\"`" + `
	Hits  {{ .Name }}Hits ` + "`json:\"hits\"`" + `
}

type {{ .Name }}Paginator struct {
	paginator *essdk.BaseESPaginator
}

func (k Client) New{{ .Name }}Paginator(filters []essdk.BoolFilter, limit *int64) ({{ .Name }}Paginator, error) {
	paginator, err := essdk.NewPaginator(k.ES(), "{{ .Index }}", filters, limit)
	if err != nil {
		return {{ .Name }}Paginator{}, err
	}

	p := {{ .Name }}Paginator{
		paginator: paginator,
	}

	return p, nil
}

func (p {{ .Name }}Paginator) HasNext() bool {
	return !p.paginator.Done()
}

func (p {{ .Name }}Paginator) Close(ctx context.Context) error {
	return p.paginator.Deallocate(ctx)
}

func (p {{ .Name }}Paginator) NextPage(ctx context.Context) ([]{{ .Name }}, error) {
	var response {{ .Name }}SearchResponse
	err := p.paginator.Search(ctx, &response)
	if err != nil {
		return nil, err
	}

	var values []{{ .Name }}
	for _, hit := range response.Hits.Hits {
		values = append(values, hit.Source)
	}

	hits := int64(len(response.Hits.Hits))
	if hits > 0 {
		p.paginator.UpdateState(hits, response.Hits.Hits[hits-1].Sort, response.PitID)
	} else {
		p.paginator.UpdateState(hits, nil, "")
	}

	return values, nil
}

var list{{ .Name }}Filters = map[string]string{
	{{ range $key, $value := .ListFilters }}"{{ $key }}": "{{ $value }}",
	{{ end }}
}

func List{{ .Name }}(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("List{{ .Name }}")
    runtime.GC()

	// create service
	cfg := essdk.GetConfig(d.Connection)
	ke, err := essdk.NewClientCached(cfg, d.ConnectionCache, ctx)
	if err != nil {
		plugin.Logger(ctx).Error("List{{ .Name }} NewClientCached", "error", err)
		return nil, err
	}
	k := Client{Client: ke}

	sc, err := steampipesdk.NewSelfClientCached(ctx, d.ConnectionCache)
	if err != nil {
		plugin.Logger(ctx).Error("List{{ .Name }} NewSelfClientCached", "error", err)
		return nil, err
	}
	integrationID, err := sc.GetConfigTableValueOrNil(ctx, steampipesdk.OpenGovernanceConfigKeyIntegrationID)
	if err != nil {
		plugin.Logger(ctx).Error("List{{ .Name }} GetConfigTableValueOrNil for OpenGovernanceConfigKeyIntegrationID", "error", err)
		return nil, err
	}
	encodedResourceCollectionFilters, err := sc.GetConfigTableValueOrNil(ctx, steampipesdk.OpenGovernanceConfigKeyResourceCollectionFilters)
	if err != nil {
		plugin.Logger(ctx).Error("List{{ .Name }} GetConfigTableValueOrNil for OpenGovernanceConfigKeyResourceCollectionFilters", "error", err)
		return nil, err
	}
	clientType, err := sc.GetConfigTableValueOrNil(ctx, steampipesdk.OpenGovernanceConfigKeyClientType)
	if err != nil {
		plugin.Logger(ctx).Error("List{{ .Name }} GetConfigTableValueOrNil for OpenGovernanceConfigKeyClientType", "error", err)
		return nil, err
	}

	paginator, err := k.New{{ .Name }}Paginator(essdk.BuildFilter(ctx, d.QueryContext, list{{ .Name }}Filters, integrationID, encodedResourceCollectionFilters, clientType), d.QueryContext.Limit)
	if err != nil {
		plugin.Logger(ctx).Error("List{{ .Name }} New{{ .Name }}Paginator", "error", err)
		return nil, err
	}

	for paginator.HasNext() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			plugin.Logger(ctx).Error("List{{ .Name }} paginator.NextPage", "error", err)
			return nil, err
		}

		for _, v := range page {
			d.StreamListItem(ctx, v)
		}
	}

	err = paginator.Close(ctx)
	if err != nil {
		return nil, err
	}

	return nil, nil
}


var get{{ .Name }}Filters = map[string]string{
	{{ range $key, $value := .GetFilters }}"{{ $key }}": "{{ $value }}",
	{{ end }}
}

func Get{{ .Name }}(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("Get{{ .Name }}")
    runtime.GC()
	// create service
	cfg := essdk.GetConfig(d.Connection)
	ke, err := essdk.NewClientCached(cfg, d.ConnectionCache, ctx)
	if err != nil {
		return nil, err
	}
	k := Client{Client: ke}

	sc, err := steampipesdk.NewSelfClientCached(ctx, d.ConnectionCache)
	if err != nil {
		return nil, err
	}
	integrationID, err := sc.GetConfigTableValueOrNil(ctx, steampipesdk.OpenGovernanceConfigKeyIntegrationID)
	if err != nil {
		return nil, err
	}
	encodedResourceCollectionFilters, err := sc.GetConfigTableValueOrNil(ctx, steampipesdk.OpenGovernanceConfigKeyResourceCollectionFilters)
	if err != nil {
		return nil, err
	}
	clientType, err := sc.GetConfigTableValueOrNil(ctx, steampipesdk.OpenGovernanceConfigKeyClientType)
	if err != nil {
		return nil, err
	}

	limit := int64(1)
	paginator, err := k.New{{ .Name }}Paginator(essdk.BuildFilter(ctx, d.QueryContext, get{{ .Name }}Filters, integrationID, encodedResourceCollectionFilters, clientType), &limit)
	if err != nil {
		return nil, err
	}

	for paginator.HasNext() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page {
			return v, nil
		}
	}

	err = paginator.Close(ctx)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ==========================  END: {{ .Name }} =============================

`)
	if err != nil {
		panic(err)
	}

	out, err := os.Create(*output)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, *file, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(&buf, "// Code is generated by go generate. DO NOT EDIT.")
	fmt.Fprintf(&buf, "package opengovernance")

	var sources []IntegrationType

	ast.Inspect(node, func(n ast.Node) bool {
		decl, ok := n.(*ast.GenDecl)
		if !ok || decl.Tok != token.TYPE {
			return true
		}

		for _, spec := range decl.Specs {
			t := spec.(*ast.TypeSpec)

			if !strings.HasSuffix(t.Name.String(), "Description") {
				fmt.Println("Ignoring type " + t.Name.String())
				continue
			}

			s := IntegrationType{
				Name:            strings.TrimSuffix(t.Name.String(), "Description"),
				IntegrationType: configs.IntegrationTypeLower,
				GetFilters:      map[string]string{},
				ListFilters:     map[string]string{},
			}
			for _, resourceType := range resourceTypes {
				if resourceType.Model == s.Name {
					var stopWordsRe = regexp.MustCompile(`\W+`)
					index := stopWordsRe.ReplaceAllString(resourceType.ResourceName, "_")
					index = strings.ToLower(index)
					s.Index = index

					fileName := *pluginPath + "/table_" + resourceType.SteampipeTable + ".go"
					tableFileSet := token.NewFileSet()
					tableNode, err := parser.ParseFile(tableFileSet, fileName, nil, parser.ParseComments)
					if err != nil {
						panic(err)
					}

					ast.Inspect(tableNode, func(tnode ast.Node) bool {
						if c, ok := tnode.(*ast.CompositeLit); ok {

							var columnName, transformer string
							for _, arg := range c.Elts {
								if kv, ok := arg.(*ast.KeyValueExpr); ok {
									if i, ok := kv.Key.(*ast.Ident); ok {
										if i.Name == "Name" {
											if bl, ok := kv.Value.(*ast.BasicLit); ok {
												columnName = strings.Trim(bl.Value, "\"")
											}
										} else if i.Name == "Transform" {
											if cl, ok := kv.Value.(*ast.CallExpr); ok {
												transformer = extractTransformer(cl)
											}
										}
									}
								}
							}

							if columnName != "" && transformer != "" {
								s.GetFilters[columnName] = transformer
								s.ListFilters[columnName] = transformer
							}
							return true
						}
						return true
					})
				}
			}

			if decl.Doc != nil {
				for _, c := range decl.Doc.List {
					if strings.HasPrefix(c.Text, "//index:") {
						//s.Index = strings.TrimSpace(strings.TrimPrefix(c.Text, "//index:"))
					} else if strings.HasPrefix(c.Text, "//getfilter:") {
						f := strings.TrimSpace(strings.TrimPrefix(c.Text, "//getfilter:"))
						fparts := strings.Split(f, "=")
						s.GetFilters[fparts[0]] = fparts[1]
					} else if strings.HasPrefix(c.Text, "//listfilter:") {
						f := strings.TrimSpace(strings.TrimPrefix(c.Text, "//listfilter:"))
						fparts := strings.Split(f, "=")
						s.ListFilters[fparts[0]] = fparts[1]
					}
				}
			}

			if s.Index != "" {
				sources = append(sources, s)
			} else {
				fmt.Println("failed to find the index:", s.Name)
			}
		}
		return false
	})

	if len(sources) > 0 {
		fmt.Fprintln(&buf, `
		import (
			"context"
			essdk "github.com/opengovern/og-util/pkg/opengovernance-es-sdk"
			steampipesdk "github.com/opengovern/og-util/pkg/steampipe"
			"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
			`+configs.IntegrationTypeLower+` "`+configs.OGPluginRepoURL+`/provider/model"
            "runtime"
		)

		type Client struct {
			essdk.Client
		}

		`)
	}

	for _, source := range sources {
		err := tpl.Execute(&buf, source)
		if err != nil {
			panic(err)
		}
	}

	source, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	_, err = out.Write(source)
	if err != nil {
		panic(err)
	}
}

func extractTransformer(cl *ast.CallExpr) string {
	if sl, ok := cl.Fun.(*ast.SelectorExpr); ok {
		if sl.Sel.Name == "Transform" {
			return ""
		}
		if call, ok := sl.X.(*ast.CallExpr); ok {
			return extractTransformer(call)
		}
		if sl.Sel.Name == "FromField" {
			for _, arg := range cl.Args {
				if bl, ok := arg.(*ast.BasicLit); ok {
					return strings.Trim(bl.Value, "\"")
				}
			}
		}
	}
	return ""
}
