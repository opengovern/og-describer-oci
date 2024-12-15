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
	switch creds.GetRegistryType() {
	case configs2.RegistryTypeGHCR:
		return listGithubImages(ctx, creds, triggerType, stream)
	case configs2.RegistryTypeDockerhub:
		//TODO
	case configs2.RegistryTypeECR:
		//TODO
	case configs2.RegistryTypeACR:
		fallthrough
	default:
		last := ""
		isMore := true
		regHost := GetRegistryFromContext(ctx)
		client := GetOrasClientFromContext(ctx)

		reg, err := remote.NewRegistry(regHost)
		if err != nil {
			return nil, err
		}
		reg.Client = client

		var images []string
		for isMore {
			err = reg.Repositories(ctx, last, func(r []string) error {
				if len(r) == 0 {
					isMore = false
					return nil
				}
				images = append(images, r...)
				last = r[len(r)-1]
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
		return images, err
	}
	return nil, fmt.Errorf("unsupported registry type")
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
				RegistryType: creds.GetRegistryType(),
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

const maxTagPerImage = 20

func OCIImageTag(ctx context.Context, creds *configs.IntegrationCredentials, triggerType string, stream *models.StreamSender) ([]models.Resource, error) {
	regHost := GetRegistryFromContext(ctx)
	client := GetOrasClientFromContext(ctx)

	var resources []models.Resource

	images, err := listImages(ctx, creds, triggerType, stream)
	if err != nil {
		return nil, err
	}

imageLabel:
	for _, imageName := range images {
		repo, err := remote.NewRepository(fmt.Sprintf("%s/%s", regHost, imageName))
		if err != nil {
			return nil, err
		}
		repo.Client = client
		repo.TagListPageSize = 10000

		lastTag := ""
		isMoreTags := true
		var tags []string
		for isMoreTags {
			err = repo.Tags(ctx, lastTag, func(t []string) error {
				if len(t) == 0 {
					isMoreTags = false
					return nil
				}
				tags = append(tags, t...)
				lastTag = t[len(t)-1]
				return nil
			})
			if err != nil {
				continue imageLabel
			}
		}
		if len(tags) > maxTagPerImage {
			tags = tags[len(tags)-maxTagPerImage:]
		}
		for _, tag := range tags {
			lastTag = tag
			ref, manifestBlob, err := repo.FetchReference(ctx, tag)
			if err != nil {
				return nil, err
			}

			manifest, err := io.ReadAll(manifestBlob)
			if err != nil {
				return nil, fmt.Errorf("failed to read manifest: %v", err)
			}

			resource := models.Resource{
				ID:   fmt.Sprintf("%s/%s:%s", regHost, imageName, tag),
				Name: fmt.Sprintf("%s:%s", imageName, tag),
				Description: model.OCIImageTagDescription{
					RegistryType: creds.GetRegistryType(),
					Repository:   regHost,
					Image:        imageName,
					Tag:          tag,
					Manifest:     string(manifest),
					Descriptor:   ref,
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

		resource := models.Resource{
			Name: imageName,
			Description: model.OCIImageDescription{
				RegistryType: creds.GetRegistryType(),
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
