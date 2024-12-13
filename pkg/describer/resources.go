//go:generate go run ../SDK/runable/resourceType/resource_types_generator.go  --output resource_types.go --index-map ../steampipe/table_index_map.go && gofmt -w -s resource_types.go  && goimports -w resource_types.go

package describer

import (
	"context"
	"fmt"
	model "github.com/opengovern/og-describer-oci/pkg/sdk/models"
	"github.com/opengovern/og-describer-oci/provider"
	"github.com/opengovern/og-describer-oci/provider/configs"
	"github.com/opengovern/og-describer-oci/provider/describer"
	"github.com/opengovern/og-util/pkg/describe/enums"
	"go.uber.org/zap"
	"sort"
	"strings"
)

func ListResourceTypes() []string {
	var list []string
	for k := range provider.ResourceTypes {
		list = append(list, k)
	}

	sort.Strings(list)
	return list
}

func GetResourceType(resourceType string) (*model.ResourceType, error) {
	if r, ok := provider.ResourceTypes[resourceType]; ok {
		return &r, nil
	}
	resourceType = strings.ToLower(resourceType)
	for k, v := range provider.ResourceTypes {
		k := strings.ToLower(k)
		v := v
		if k == resourceType {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("resource type %s not found", resourceType)
}

func GetResourceTypesMap() map[string]model.ResourceType {
	return provider.ResourceTypes
}

func GetResources(
	ctx context.Context,
	logger *zap.Logger,
	resourceType string,
	triggerType enums.DescribeTriggerType,
	cfg configs.IntegrationCredentials,
	additionalParameters map[string]string,
	stream *model.StreamSender,
) error {
	_, err := describe(ctx, logger, cfg, resourceType, triggerType, additionalParameters, stream)
	if err != nil {
		return err
	}
	return nil
}

func describe(ctx context.Context, logger *zap.Logger, accountCfg configs.IntegrationCredentials, resourceType string, triggerType enums.DescribeTriggerType, additionalParameters map[string]string, stream *model.StreamSender) ([]model.Resource, error) {
	resourceTypeObject, ok := provider.ResourceTypes[resourceType]
	if !ok {
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
	ctx = describer.WithLogger(ctx, logger)

	return resourceTypeObject.ListDescriber(ctx, accountCfg, triggerType, additionalParameters, stream)
}

func GetSingleResource(
	ctx context.Context,
	logger *zap.Logger,
	resourceType string,
	triggerType enums.DescribeTriggerType,
	cfg configs.IntegrationCredentials,
	additionalParameters map[string]string,
	resourceId string,
	stream *model.StreamSender,
) error {
	_, err := describeSingle(ctx, logger, cfg, resourceType, resourceId, triggerType, additionalParameters, stream)
	if err != nil {
		return err
	}
	return nil
}

func describeSingle(ctx context.Context, logger *zap.Logger, accountCfg configs.IntegrationCredentials, resourceType string, resourceID string, triggerType enums.DescribeTriggerType, additionalParameters map[string]string, stream *model.StreamSender) (*model.Resource, error) {
	resourceTypeObject, ok := provider.ResourceTypes[resourceType]
	if !ok {
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
	ctx = describer.WithLogger(ctx, logger)

	return resourceTypeObject.GetDescriber(ctx, accountCfg, triggerType, additionalParameters, resourceID)
}
