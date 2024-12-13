package provider

import (
	model "github.com/opengovern/og-describer-oci/pkg/sdk/models"
	"github.com/opengovern/og-describer-oci/provider/configs"
	"github.com/opengovern/og-describer-oci/provider/describer"
)

var ResourceTypes = map[string]model.ResourceType{

	"OCI::Image": {
		IntegrationType: configs.IntegrationName,
		ResourceName:    "OCI::Image",
		Tags:            map[string][]string{},
		Labels:          map[string]string{},
		Annotations:     map[string]string{},
		ListDescriber:   DescribeByIntegration(describer.OCIImage),
		GetDescriber:    nil,
	},

	"OCI::ImageTag": {
		IntegrationType: configs.IntegrationName,
		ResourceName:    "OCI::ImageTag",
		Tags:            map[string][]string{},
		Labels:          map[string]string{},
		Annotations:     map[string]string{},
		ListDescriber:   DescribeByIntegration(describer.OCIImageTag),
		GetDescriber:    nil,
	},
}
