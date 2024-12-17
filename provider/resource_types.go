package provider
import (
	"github.com/opengovern/og-describer-oci/provider/describer"
	"github.com/opengovern/og-describer-oci/provider/configs"
	model "github.com/opengovern/og-describer-oci/pkg/sdk/models"
)
var ResourceTypes = map[string]model.ResourceType{

	"OCI::Artifact": {
		IntegrationType:      configs.IntegrationName,
		ResourceName:         "OCI::Artifact",
		Tags:                 map[string][]string{
        },
		Labels:               map[string]string{
        },
		Annotations:          map[string]string{
        },
		ListDescriber:        DescribeByIntegration(describer.OCIArtifact),
		GetDescriber:         nil,
	},
}
