package describer

import (
	"context"
	"fmt"
	"github.com/opengovern/og-describer-oci/pkg/sdk/models"
	"github.com/opengovern/og-describer-oci/provider/configs"
	"github.com/opengovern/og-describer-oci/provider/model"
	"io"
	"oras.land/oras-go/v2/registry/remote"
)

func OCIImage(ctx context.Context, creds *configs.IntegrationCredentials, triggerType string, stream *models.StreamSender) ([]models.Resource, error) {
	regHost := GetRegistryFromContext(ctx)
	client := GetOrasClientFromContext(ctx)

	reg, err := remote.NewRegistry(regHost)
	if err != nil {
		return nil, err
	}
	reg.Client = client

	var resources []models.Resource
	last := ""
	isMore := true
	for isMore {
		err = reg.Repositories(ctx, last, func(r []string) error {
			if len(r) == 0 {
				isMore = false
				return nil
			}
			for _, repo := range r {
				resource := models.Resource{
					Name: repo,
					Description: model.OCIImageDescription{
						RegistryType: creds.RegistryType,
						Repository:   regHost,
						Image:        repo,
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

	return resources, nil
}

func OCIImageTag(ctx context.Context, creds *configs.IntegrationCredentials, triggerType string, stream *models.StreamSender) ([]models.Resource, error) {
	regHost := GetRegistryFromContext(ctx)
	client := GetOrasClientFromContext(ctx)

	reg, err := remote.NewRegistry(regHost)
	if err != nil {
		return nil, err
	}
	reg.Client = client

	var resources []models.Resource
	last := ""
	isMore := true
	for isMore {
		err = reg.Repositories(ctx, last, func(r []string) error {
			if len(r) == 0 {
				isMore = false
				return nil
			}
			for _, repoName := range r {
				repo, err := remote.NewRepository(fmt.Sprintf("%s/%s", regHost, repoName))
				if err != nil {
					return err
				}

				lastTag := ""
				isMoreTags := true
				for isMoreTags {
					err = repo.Tags(ctx, lastTag, func(t []string) error {
						if len(t) == 0 {
							isMoreTags = false
							return nil
						}
						for _, tag := range t {
							ref, manifestBlob, err := repo.FetchReference(ctx, tag)
							if err != nil {
								return err
							}

							manifest, err := io.ReadAll(manifestBlob)
							if err != nil {
								return fmt.Errorf("failed to read manifest: %v", err)
							}

							resource := models.Resource{
								Name: fmt.Sprintf("%s:%s", repoName, tag),
								Description: model.OCIImageTagDescription{
									RegistryType: creds.RegistryType,
									Repository:   regHost,
									Image:        repoName,
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
						return err
					}
				}

				resource := models.Resource{
					Name: repoName,
					Description: model.OCIImageDescription{
						RegistryType: creds.RegistryType,
						Repository:   regHost,
						Image:        repoName,
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

	return resources, nil
}
