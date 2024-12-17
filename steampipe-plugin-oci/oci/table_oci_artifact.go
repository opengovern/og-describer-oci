package oci

import (
	"context"
	opengovernance "github.com/opengovern/og-describer-oci/pkg/sdk/es"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableOCIArtifact(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "oci_artifact",
		Description: "Retrieve information about oci artifacts across multiple namespaces",
		List: &plugin.ListConfig{
			Hydrate: opengovernance.ListOCIArtifact,
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
			{
				Name:      "digest",
				Type:      proto.ColumnType_STRING,
				Transform: transform.FromField("Description.Digest"),
			},
			{
				Name:      "media_type",
				Type:      proto.ColumnType_STRING,
				Transform: transform.FromField("Description.MediaType"),
			},
			{
				Name:      "size",
				Type:      proto.ColumnType_INT,
				Transform: transform.FromField("Description.Size"),
			},
			{
				Name:      "tags",
				Type:      proto.ColumnType_JSON,
				Transform: transform.FromField("Description.Tags"),
			},
			{
				Name:      "manifest",
				Type:      proto.ColumnType_JSON,
				Transform: transform.FromField("Description.Manifest"),
			},
		}),
	}
}
