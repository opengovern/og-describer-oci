//go:generate go run ../../pkg/sdk/runable/steampipe_es_client_generator/main.go -pluginPath ../../steampipe-plugin-oci/oci -file $GOFILE -output ../../pkg/sdk/es/resources_clients.go -resourceTypesFile ../resource_types/resource-types.json

// Implement types for each resource

package model

import (
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opengovern/opencomply/services/integration/integration-type/oci-repository/configs"
)

type Metadata struct {
}

type OCIImageDescription struct {
	RegistryType configs.RegistryType
	Repository   string
	Image        string
	Tag          string
	Manifest     string
	Descriptor   ocispec.Descriptor
}

type OCIImageTagDescription struct {
}
