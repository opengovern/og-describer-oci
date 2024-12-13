package describer

import (
	"context"
	"fmt"
	"github.com/google/go-github/v67/github"
	"github.com/opengovern/og-describer-oci/pkg/sdk/models"
	"github.com/opengovern/og-describer-oci/provider/configs"
	"github.com/opengovern/og-describer-oci/provider/model"
	"github.com/opengovern/opencomply/pkg/utils"
	configs2 "github.com/opengovern/opencomply/services/integration/integration-type/oci-repository/configs"
	"io"
	"oras.land/oras-go/v2/registry/remote"
)

func listGithubImages(ctx context.Context, creds *configs.IntegrationCredentials, triggerType string, stream *models.StreamSender) ([]string, error) {
	owner := GetOwnerFromContext(ctx)
	if creds.GhcrCredentials == nil {
		return nil, fmt.Errorf("missing required GHCR credentials")
	}

	client := github.NewClient(nil).WithAuthToken(creds.GhcrCredentials.Token)
	packages, _, err := client.Organizations.ListPackages(ctx, owner, &github.PackageListOptions{PackageType: utils.GetPointer("container")})
	if err != nil {
		// TODO handle 404
		return nil, err
	}

	imagesMap := make(map[string]bool)
	for _, pkg := range packages {
		imagesMap[*pkg.Name] = true
	}

	var images []string
	for image := range imagesMap {
		images = append(images, image)
	}

	return images, nil
}

func listImages(ctx context.Context, creds *configs.IntegrationCredentials, triggerType string, stream *models.StreamSender) ([]string, error) {
	switch creds.RegistryType {
	case configs2.RegistryTypeDockerhub:
		//TODO
	case configs2.RegistryTypeGHCR:
		return listGithubImages(ctx, creds, triggerType, stream)
	case configs2.RegistryTypeECR:
		//TODO
	case configs2.RegistryTypeACR:
		//TODO
	}
	return nil, fmt.Errorf("unsupported registry type: %s", creds.RegistryType)
}

func OCIImage(ctx context.Context, creds *configs.IntegrationCredentials, triggerType string, stream *models.StreamSender) ([]models.Resource, error) {
	regHost := GetRegistryFromContext(ctx)
	var resources []models.Resource

	images, err := listImages(ctx, creds, triggerType, stream)
	if err != nil {
		return nil, err
	}

	for _, image := range images {
		resource := models.Resource{
			ID:   fmt.Sprintf("%s/%s", regHost, image),
			Name: image,
			Description: model.OCIImageDescription{
				RegistryType: creds.RegistryType,
				Repository:   regHost,
				Image:        image,
			},
		}

		if stream != nil {
			if err := (*stream)(resource); err != nil {
				return nil, err
			}
		} else {
			resources = append(resources, resource)
		}
	}
	return resources, nil
}

func OCIImageTag(ctx context.Context, creds *configs.IntegrationCredentials, triggerType string, stream *models.StreamSender) ([]models.Resource, error) {
	regHost := GetRegistryFromContext(ctx)
	client := GetOrasClientFromContext(ctx)

	var resources []models.Resource

	images, err := listImages(ctx, creds, triggerType, stream)
	if err != nil {
		return nil, err
	}

	for _, imageName := range images {
		repo, err := remote.NewRepository(fmt.Sprintf("%s/%s", regHost, imageName))
		if err != nil {
			return nil, err
		}
		repo.Client = client

		lastTag := ""
		isMoreTags := true
		for isMoreTags {
			err = repo.Tags(ctx, lastTag, func(t []string) error {
				if len(t) == 0 {
					isMoreTags = false
					return nil
				}
				for _, tag := range t {
					lastTag = tag
					ref, manifestBlob, err := repo.FetchReference(ctx, tag)
					if err != nil {
						return err
					}

					manifest, err := io.ReadAll(manifestBlob)
					if err != nil {
						return fmt.Errorf("failed to read manifest: %v", err)
					}

					resource := models.Resource{
						ID:   fmt.Sprintf("%s/%s:%s", regHost, imageName, tag),
						Name: fmt.Sprintf("%s:%s", imageName, tag),
						Description: model.OCIImageTagDescription{
							RegistryType: creds.RegistryType,
							Repository:   regHost,
							Image:        imageName,
							Tag:          tag,
							Manifest:     string(manifest),
							Descriptor:   ref,
						},
					}

					if stream != nil {
						if err := (*stream)(resource); err != nil {
							return err
						}
					} else {
						resources = append(resources, resource)
					}
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		}

		resource := models.Resource{
			Name: imageName,
			Description: model.OCIImageDescription{
				RegistryType: creds.RegistryType,
				Repository:   regHost,
				Image:        imageName,
			},
		}

		if stream != nil {
			if err := (*stream)(resource); err != nil {
				return nil, err
			}
		} else {
			resources = append(resources, resource)
		}
	}

	return resources, nil
}
