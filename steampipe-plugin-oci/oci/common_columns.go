package oci

import (
	"context"
	"encoding/json"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func commonColumns() []*plugin.Column {
	return []*plugin.Column{
		{
			Name:        "platform_account_id",
			Type:        proto.ColumnType_STRING,
			Description: "The Platform Account ID in which the resource is located.",
			Transform:   transform.FromField("IntegrationID"),
		},
		{
			Name:        "platform_resource_id",
			Type:        proto.ColumnType_STRING,
			Description: "The unique ID of the resource in opengovernance.",
			Transform:   transform.FromField("PlatformID"),
		},
		{
			Name:        "platform_metadata",
			Type:        proto.ColumnType_JSON,
			Description: "The metadata of the resource",
			Transform:   transform.FromField("Metadata").Transform(marshalJSON),
		},
		{
			Name:        "platform_resource_description",
			Type:        proto.ColumnType_JSON,
			Description: "The full model description of the resource",
			Transform:   transform.FromField("Description").Transform(marshalJSON),
		},
	}
}

func integrationColumns(columns []*plugin.Column) []*plugin.Column {
	return append(columns, commonColumns()...)
}

func marshalJSON(_ context.Context, d *transform.TransformData) (interface{}, error) {
	b, err := json.Marshal(d.Value)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}
