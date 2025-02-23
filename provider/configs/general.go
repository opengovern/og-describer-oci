package configs

import (
	"github.com/opengovern/opensecurity/services/integration/integration-type/oci-repository/configs"
)

const (
	IntegrationTypeLower = "oci_repository"                         // example: aws, azure
	IntegrationName      = configs.IntegrationTypeOciRepository     // example: AWS_ACCOUNT, AZURE_SUBSCRIPTION
	OGPluginRepoURL      = "github.com/opengovern/og-describer-oci" // example: github.com/opengovern/og-describer-aws
)
