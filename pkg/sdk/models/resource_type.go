package models

import (
	"github.com/opengovern/og-describer-oci/provider/configs"
	"github.com/opengovern/og-util/pkg/describe/enums"
	"github.com/opengovern/og-util/pkg/integration"
	"golang.org/x/net/context"
)

// any types are used to load your provider configuration.
type ResourceDescriber func(context.Context, configs.IntegrationCredentials, enums.DescribeTriggerType, map[string]string, *StreamSender) ([]Resource, error)
type SingleResourceDescriber func(context.Context, configs.IntegrationCredentials, enums.DescribeTriggerType, map[string]string, string) (*Resource, error)

type ResourceType struct {
	IntegrationType integration.Type
	ResourceName    string

	ListDescriber ResourceDescriber
	GetDescriber  SingleResourceDescriber

	Annotations map[string]string
	Labels      map[string]string
	Tags        map[string][]string
}

func (r ResourceType) GetIntegrationType() integration.Type {
	return r.IntegrationType
}

func (r ResourceType) GetResourceName() string {
	return r.ResourceName
}

func (r ResourceType) GetTags() map[string][]string {
	return r.Tags
}
