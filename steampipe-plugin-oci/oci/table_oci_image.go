package oci

import (
	"context"
	opengovernance "github.com/opengovern/og-describer-oci/pkg/sdk/es"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableOCIImage(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "oci_image",
		Description: "Retrieve information about images in the repository",
		List: &plugin.ListConfig{
			Hydrate: opengovernance.ListOCIImage,
		},
		Columns: integrationColumns([]*plugin.Column{
			{
				Name:      "registry_type",
				Type:      proto.ColumnType_STRING,
				Transform: transform.FromField("Description.RegistryType"),
			},
			{
				Name:      "repository",
				Type:      proto.ColumnType_STRING,
				Transform: transform.FromField("Description.Repository"),
			},
			{
				Name:      "image",
				Type:      proto.ColumnType_STRING,
				Transform: transform.FromField("Description.Image"),
			},
		}),
	}
}
