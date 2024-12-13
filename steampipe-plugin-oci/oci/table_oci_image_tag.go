package oci

import (
	"context"
	opengovernance "github.com/opengovern/og-describer-oci/pkg/sdk/es"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func tableOCIImageTag(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "oci_image_tag",
		Description: "Retrieve information about images in the repository",
		List: &plugin.ListConfig{
			Hydrate: opengovernance.ListOCIImageTag,
		},
		Columns: integrationColumns([]*plugin.Column{
			// Top columns
		}),
	}
}
