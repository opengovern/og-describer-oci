package oci

import (
	"context"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func tableOCIImage(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "oci_image",
		Description: "Retrieve information about images in the repository",
		List:        &plugin.ListConfig{
			//Hydrate: opengovernance.ListDigitalOceanAccount,
		},
		Columns: integrationColumns([]*plugin.Column{
			// Top columns

		}),
	}
}
