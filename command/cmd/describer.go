package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/opengovern/og-describer-oci/pkg/describer"
	model "github.com/opengovern/og-describer-oci/pkg/sdk/models"
	"github.com/opengovern/og-describer-oci/provider"
	"github.com/opengovern/og-describer-oci/provider/configs"
	"github.com/opengovern/og-describer-oci/steampipe"
	"github.com/opengovern/og-util/pkg/describe"
	"github.com/opengovern/og-util/pkg/es"
	configs2 "github.com/opengovern/opensecurity/services/integration/integration-type/oci-repository/configs"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	resourceType string
	outputFile   string
)

// describerCmd represents the describer command
var describerCmd = &cobra.Command{
	Use:   "describer",
	Short: "A brief description of your command",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Open the output file
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close() // Ensure the file is closed at the end

		job := describe.DescribeJob{
			JobID:                  uint(uuid.New().ID()),
			ResourceType:           resourceType,
			IntegrationID:          "",
			ProviderID:             "",
			DescribedAt:            time.Now().UnixMilli(),
			IntegrationType:        configs.IntegrationTypeLower,
			CipherText:             "",
			IntegrationLabels:      nil,
			IntegrationAnnotations: nil,
		}

		ctx := context.Background()
		logger, _ := zap.NewProduction()

		// TODO: Set the credentials
		creds := configs.IntegrationCredentials{
			IntegrationCredentials: configs2.IntegrationCredentials{},
		}

		additionalParameters, err := provider.GetAdditionalParameters(job)
		if err != nil {
			return err
		}
		plg := steampipe.Plugin()

		f := func(resource model.Resource) error {
			if resource.Description == nil {
				return nil
			}
			descriptionJSON, err := json.Marshal(resource.Description)
			if err != nil {
				return fmt.Errorf("failed to marshal description: %w", err)
			}
			descriptionJSON, err = trimJsonFromEmptyObjects(descriptionJSON)
			if err != nil {
				return fmt.Errorf("failed to trim json: %w", err)
			}

			metadata, err := provider.GetResourceMetadata(job, resource)
			if err != nil {
				return fmt.Errorf("failed to get resource metadata")
			}
			err = provider.AdjustResource(job, &resource)
			if err != nil {
				return fmt.Errorf("failed to adjust resource metadata")
			}

			desc := resource.Description
			err = json.Unmarshal(descriptionJSON, &desc)
			if err != nil {
				return fmt.Errorf("unmarshal description: %v", err.Error())
			}

			if plg != nil {
				_, _, err = steampipe.ExtractTagsAndNames(logger, plg, job.ResourceType, resource)
				if err != nil {
					logger.Error("failed to build tags for service", zap.Error(err), zap.String("resourceType", job.ResourceType), zap.Any("resource", resource))
				}
			}

			var description any
			err = json.Unmarshal([]byte(descriptionJSON), &description)
			if err != nil {
				logger.Error("failed to parse resource description json", zap.Error(err))
				return fmt.Errorf("failed to parse resource description json")
			}

			res := es.Resource{
				PlatformID:      fmt.Sprintf("%s:::%s:::%s", job.IntegrationID, job.ResourceType, resource.UniqueID()),
				ResourceID:      resource.UniqueID(),
				ResourceName:    resource.Name,
				Description:     description,
				IntegrationType: configs.IntegrationName,
				ResourceType:    strings.ToLower(job.ResourceType),
				IntegrationID:   job.IntegrationID,
				Metadata:        metadata,
				DescribedAt:     job.DescribedAt,
				DescribedBy:     strconv.FormatUint(uint64(job.JobID), 10),
			}

			// Write the resource JSON to the file
			resJSON, err := json.Marshal(res)
			if err != nil {
				return fmt.Errorf("failed to marshal resource JSON: %w", err)
			}
			_, err = file.Write(resJSON)
			if err != nil {
				return fmt.Errorf("failed to write to file: %w", err)
			}
			_, err = file.Write([]byte(",\n")) // Add a newline for readability
			if err != nil {
				return fmt.Errorf("failed to write newline to file: %w", err)
			}

			return nil
		}
		clientStream := (*model.StreamSender)(&f)

		err = describer.GetResources(
			ctx,
			logger,
			job.ResourceType,
			job.TriggerType,
			creds,
			additionalParameters,
			clientStream,
		)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	describerCmd.Flags().StringVar(&resourceType, "resourceType", "", "Resource type")
	describerCmd.Flags().StringVar(&outputFile, "outputFile", "output.json", "File to write JSON outputs")
}

func trimJsonFromEmptyObjects(input []byte) ([]byte, error) {
	unknownData := map[string]any{}
	err := json.Unmarshal(input, &unknownData)
	if err != nil {
		return nil, err
	}
	trimEmptyMaps(unknownData)
	return json.Marshal(unknownData)
}

func trimEmptyMaps(input map[string]any) {
	for key, value := range input {
		switch value.(type) {
		case map[string]any:
			if len(value.(map[string]any)) != 0 {
				trimEmptyMaps(value.(map[string]any))
			}
			if len(value.(map[string]any)) == 0 {
				delete(input, key)
			}
		}
	}
}
