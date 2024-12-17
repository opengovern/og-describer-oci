// Code is generated by go generate. DO NOT EDIT.
package opengovernance

import (
	"context"
	oci_repository "github.com/opengovern/og-describer-oci/provider/model"
	essdk "github.com/opengovern/og-util/pkg/opengovernance-es-sdk"
	steampipesdk "github.com/opengovern/og-util/pkg/steampipe"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"runtime"
)

type Client struct {
	essdk.Client
}

// ==========================  START: OCIArtifact =============================

type OCIArtifact struct {
	ResourceID      string                                `json:"resource_id"`
	PlatformID      string                                `json:"platform_id"`
	Description     oci_repository.OCIArtifactDescription `json:"description"`
	Metadata        oci_repository.Metadata               `json:"metadata"`
	DescribedBy     string                                `json:"described_by"`
	ResourceType    string                                `json:"resource_type"`
	IntegrationType string                                `json:"integration_type"`
	IntegrationID   string                                `json:"integration_id"`
}

type OCIArtifactHit struct {
	ID      string        `json:"_id"`
	Score   float64       `json:"_score"`
	Index   string        `json:"_index"`
	Type    string        `json:"_type"`
	Version int64         `json:"_version,omitempty"`
	Source  OCIArtifact   `json:"_source"`
	Sort    []interface{} `json:"sort"`
}

type OCIArtifactHits struct {
	Total essdk.SearchTotal `json:"total"`
	Hits  []OCIArtifactHit  `json:"hits"`
}

type OCIArtifactSearchResponse struct {
	PitID string          `json:"pit_id"`
	Hits  OCIArtifactHits `json:"hits"`
}

type OCIArtifactPaginator struct {
	paginator *essdk.BaseESPaginator
}

func (k Client) NewOCIArtifactPaginator(filters []essdk.BoolFilter, limit *int64) (OCIArtifactPaginator, error) {
	paginator, err := essdk.NewPaginator(k.ES(), "oci_artifact", filters, limit)
	if err != nil {
		return OCIArtifactPaginator{}, err
	}

	p := OCIArtifactPaginator{
		paginator: paginator,
	}

	return p, nil
}

func (p OCIArtifactPaginator) HasNext() bool {
	return !p.paginator.Done()
}

func (p OCIArtifactPaginator) Close(ctx context.Context) error {
	return p.paginator.Deallocate(ctx)
}

func (p OCIArtifactPaginator) NextPage(ctx context.Context) ([]OCIArtifact, error) {
	var response OCIArtifactSearchResponse
	err := p.paginator.Search(ctx, &response)
	if err != nil {
		return nil, err
	}

	var values []OCIArtifact
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

var listOCIArtifactFilters = map[string]string{
	"digest":        "Description.Digest",
	"image":         "Description.Image",
	"manifest":      "Description.Manifest",
	"media_type":    "Description.MediaType",
	"registry_type": "Description.RegistryType",
	"repository":    "Description.Repository",
	"size":          "Description.Size",
	"tags":          "Description.Tags",
}

func ListOCIArtifact(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("ListOCIArtifact")
	runtime.GC()

	// create service
	cfg := essdk.GetConfig(d.Connection)
	ke, err := essdk.NewClientCached(cfg, d.ConnectionCache, ctx)
	if err != nil {
		plugin.Logger(ctx).Error("ListOCIArtifact NewClientCached", "error", err)
		return nil, err
	}
	k := Client{Client: ke}

	sc, err := steampipesdk.NewSelfClientCached(ctx, d.ConnectionCache)
	if err != nil {
		plugin.Logger(ctx).Error("ListOCIArtifact NewSelfClientCached", "error", err)
		return nil, err
	}
	integrationID, err := sc.GetConfigTableValueOrNil(ctx, steampipesdk.OpenGovernanceConfigKeyIntegrationID)
	if err != nil {
		plugin.Logger(ctx).Error("ListOCIArtifact GetConfigTableValueOrNil for OpenGovernanceConfigKeyIntegrationID", "error", err)
		return nil, err
	}
	encodedResourceCollectionFilters, err := sc.GetConfigTableValueOrNil(ctx, steampipesdk.OpenGovernanceConfigKeyResourceCollectionFilters)
	if err != nil {
		plugin.Logger(ctx).Error("ListOCIArtifact GetConfigTableValueOrNil for OpenGovernanceConfigKeyResourceCollectionFilters", "error", err)
		return nil, err
	}
	clientType, err := sc.GetConfigTableValueOrNil(ctx, steampipesdk.OpenGovernanceConfigKeyClientType)
	if err != nil {
		plugin.Logger(ctx).Error("ListOCIArtifact GetConfigTableValueOrNil for OpenGovernanceConfigKeyClientType", "error", err)
		return nil, err
	}

	paginator, err := k.NewOCIArtifactPaginator(essdk.BuildFilter(ctx, d.QueryContext, listOCIArtifactFilters, integrationID, encodedResourceCollectionFilters, clientType), d.QueryContext.Limit)
	if err != nil {
		plugin.Logger(ctx).Error("ListOCIArtifact NewOCIArtifactPaginator", "error", err)
		return nil, err
	}

	for paginator.HasNext() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			plugin.Logger(ctx).Error("ListOCIArtifact paginator.NextPage", "error", err)
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

var getOCIArtifactFilters = map[string]string{
	"digest":        "Description.Digest",
	"image":         "Description.Image",
	"manifest":      "Description.Manifest",
	"media_type":    "Description.MediaType",
	"registry_type": "Description.RegistryType",
	"repository":    "Description.Repository",
	"size":          "Description.Size",
	"tags":          "Description.Tags",
}

func GetOCIArtifact(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("GetOCIArtifact")
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
	paginator, err := k.NewOCIArtifactPaginator(essdk.BuildFilter(ctx, d.QueryContext, getOCIArtifactFilters, integrationID, encodedResourceCollectionFilters, clientType), &limit)
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

// ==========================  END: OCIArtifact =============================
