package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/opengovern/og-describer-oci/provider/configs"
	"os"
	"strings"
)

// Define the ResourceType struct with Labels and Annotations
type ResourceType struct {
	ResourceName      string
	Tags              map[string][]string
	TagsString        string `json:"-"`
	ListDescriber     string
	GetDescriber      string
	SteampipeTable    string
	Model             string
	Annotations       map[string]string
	Labels            map[string]string
	AnnotationsString string `json:"-"`
	LabelsString      string `json:"-"`
}

var (
	resourceTypesFile = flag.String("resourceTypesFile", "", "Location of the resource types json file file")
	indexMap          = flag.String("index-map", "", "Path to the output file for index map")
)

func main() {
	flag.Parse()

	var resourceTypes []ResourceType

	if resourceTypesFile == nil || len(*resourceTypesFile) == 0 {
		rt := "../../../../provider/resource_types/resource-types.json"
		resourceTypesFile = &rt
	}

	if indexMap == nil || len(*indexMap) == 0 {
		v := "../../../../steampipe/table_index_map.go"
		indexMap = &v
	}

	resourceTypesFileJson, err := os.ReadFile(*resourceTypesFile)
	if err != nil {
		panic(err)
	}
	// Parse the embedded JSON into resourceTypes slice
	if err := json.Unmarshal(resourceTypesFileJson, &resourceTypes); err != nil {
		panic(err)
	}

	// Generate the index map file as before
	b := &strings.Builder{}
	b.WriteString(fmt.Sprintf(`package steampipe

import (
	"%[1]s/pkg/sdk/es"
)

var Map = map[string]string{
`, configs.OGPluginRepoURL))
	for _, resourceType := range resourceTypes {
		b.WriteString(fmt.Sprintf("  \"%s\": \"%s\",\n", resourceType.ResourceName, resourceType.SteampipeTable))
	}
	b.WriteString(fmt.Sprintf(`}

var DescriptionMap = map[string]interface{}{
`))
	for _, resourceType := range resourceTypes {
		b.WriteString(fmt.Sprintf("  \"%s\": opengovernance.%s{},\n", resourceType.ResourceName, resourceType.Model))
	}
	b.WriteString(fmt.Sprintf(`}

var ReverseMap = map[string]string{
`))

	// Build the reverse map
	for _, resourceType := range resourceTypes {
		b.WriteString(fmt.Sprintf("  \"%s\": \"%s\",\n", resourceType.SteampipeTable, resourceType.ResourceName))
	}
	b.WriteString("}\n")

	// Write the index map to the specified file
	err = os.WriteFile(*indexMap, []byte(b.String()), os.ModePerm)
	if err != nil {
		panic(err)
	}
}

// escapeString ensures that any quotes in the strings are properly escaped
func escapeString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}
