package describer

import (
	"context"
	"github.com/opengovern/og-describer-oci/pkg/sdk/models"
	"github.com/opengovern/opencomply/services/integration/integration-type/oci-repository/configs"
	"oras.land/oras-go/v2/registry/remote"
)

func OCIImage(ctx context.Context, creds *configs.IntegrationCredentials, triggerType string, stream *models.StreamSender) ([]models.Resource, error) {
	reg, err := remote.NewRegistry(creds.Host)
	if err != nil {
		return nil, err
	}
	reg.Client =

}
