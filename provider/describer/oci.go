package describer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v67/github"
	"github.com/opengovern/og-describer-oci/pkg/sdk/models"
	"github.com/opengovern/og-describer-oci/provider/configs"
	"github.com/opengovern/og-describer-oci/provider/model"
	"github.com/opengovern/opencomply/pkg/utils"
	configs2 "github.com/opengovern/opencomply/services/integration/integration-type/oci-repository/configs"
	"google.golang.org/api/artifactregistry/v1"
	"google.golang.org/api/option"
	"io"
	"net/http"
	"oras.land/oras-go/v2/registry/remote"
	"strings"
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

func listDockerhubImages(ctx context.Context, creds *configs.IntegrationCredentials, triggerType string, stream *models.StreamSender) ([]string, error) {
	if creds.DockerhubCredentials == nil {
		return nil, fmt.Errorf("missing required Dockerhub credentials")
	}

	images := make([]string, 0)

	// Login to dockerhub
	//TOKEN=$(curl -s -H "Content-Type: application/json" -X POST -d '{"username": "'${UNAME}'", "password": "'${UPASS}'"}' https://hub.docker.com/v2/users/login/ | jq -r .token)
	tokenRequest, err := http.NewRequest("POST", "https://hub.docker.com/v2/users/login/", strings.NewReader(fmt.Sprintf(`{"username": "%s", "password": "%s"}`, creds.DockerhubCredentials.Username, creds.DockerhubCredentials.Password)))
	if err != nil {
		return nil, err
	}
	tokenRequest = tokenRequest.WithContext(ctx)
	tokenRequest.Header.Set("Content-Type", "application/json")
	tokenResponse, err := http.DefaultClient.Do(tokenRequest)
	if err != nil {
		return nil, err
	}
	defer tokenResponse.Body.Close()
	if tokenResponse.StatusCode < 200 || tokenResponse.StatusCode > 299 {
		body, _ := io.ReadAll(tokenResponse.Body)
		return nil, fmt.Errorf("non-2xx status: %d, %s", tokenResponse.StatusCode, string(body))
	}
	tokenBody, err := io.ReadAll(tokenResponse.Body)
	if err != nil {
		return nil, err
	}
	tokenStruct := struct {
		Token string `json:"token"`
	}{}
	if err := json.Unmarshal(tokenBody, &tokenStruct); err != nil {
		return nil, err
	}
	token := tokenStruct.Token

	// Get the list of repositories
	//REPO_LIST=$(curl -s -H "Authorization: JWT ${TOKEN}" https://hub.docker.com/v2/repositories/${UNAME}/?page_size=100 | jq -r '.results|.[]|.name')
	repoListRequest, err := http.NewRequest("GET", fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/?page_size=100", creds.DockerhubCredentials.Owner), nil)
	if err != nil {
		return nil, err
	}
	repoListRequest = repoListRequest.WithContext(ctx)
	repoListRequest.Header.Set("Authorization", fmt.Sprintf("JWT %s", token))
	repoListResponse, err := http.DefaultClient.Do(repoListRequest)
	if err != nil {
		return nil, err
	}
	defer repoListResponse.Body.Close()
	if repoListResponse.StatusCode < 200 || repoListResponse.StatusCode > 299 {
		return nil, fmt.Errorf("non-2xx status: %d", repoListResponse.StatusCode)
	}
	repoListBody, err := io.ReadAll(repoListResponse.Body)
	if err != nil {
		return nil, err
	}
	repoListStruct := struct {
		Results []struct {
			Name string `json:"name"`
		}
	}{}
	if err := json.Unmarshal(repoListBody, &repoListStruct); err != nil {
		return nil, err
	}
	for _, repo := range repoListStruct.Results {
		images = append(images, fmt.Sprintf("%s/%s", creds.DockerhubCredentials.Owner, repo.Name))
	}

	return images, nil
}

func listGCRImages(ctx context.Context, creds *configs.IntegrationCredentials, triggerType string, stream *models.StreamSender) ([]string, error) {
	service, err := artifactregistry.NewService(ctx, option.WithCredentialsJSON([]byte(creds.GcrCredentials.JSONKey)))
	if err != nil {
		return nil, err
	}
	parent := fmt.Sprintf("projects/%s/locations/%s", creds.GcrCredentials.ProjectID, creds.GcrCredentials.Location)
	res, err := service.Projects.Locations.Repositories.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	images := make([]string, 0)
	for _, repo := range res.Repositories {
		repoRes, err := service.Projects.Locations.Repositories.DockerImages.List(repo.Name).Context(ctx).Do()
		if err != nil {
			return nil, err
		}

		for _, img := range repoRes.DockerImages {
			parts := strings.Split(img.Uri, "/")
			// registry, project, imageRepo, image
			_, project, imageRepo, image := parts[0], parts[1], parts[2], strings.Split(parts[3], "@")[0]
			images = append(images, fmt.Sprintf("%s/%s/%s", project, imageRepo, image))
		}
	}

	return images, nil
}

func listImages(ctx context.Context, creds *configs.IntegrationCredentials, triggerType string, stream *models.StreamSender) ([]string, error) {
	switch creds.GetRegistryType() {
	case configs2.RegistryTypeGHCR:
		return listGithubImages(ctx, creds, triggerType, stream)
	case configs2.RegistryTypeDockerhub:
		return listDockerhubImages(ctx, creds, triggerType, stream)
	case configs2.RegistryTypeGCR:
		return listGCRImages(ctx, creds, triggerType, stream)
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

const maxTagPerImage = 20

func OCIArtifact(ctx context.Context, creds *configs.IntegrationCredentials, triggerType string, stream *models.StreamSender) ([]models.Resource, error) {
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
				if len(t) == 0 || lastTag == t[len(t)-1] {
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
		artifacts := make(map[string]model.OCIArtifactDescription)
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
			if v, ok := artifacts[ref.Digest.String()]; ok {
				v.Tags = append(v.Tags, tag)
				artifacts[ref.Digest.String()] = v
			} else {
				artifacts[ref.Digest.String()] = model.OCIArtifactDescription{
					RegistryType: creds.GetRegistryType(),
					Repository:   regHost,
					Digest:       ref.Digest.String(),
					MediaType:    ref.MediaType,
					Size:         ref.Size,
					Image:        imageName,
					Manifest:     string(manifest),
					Tags:         []string{tag},
				}
			}
		}

		for digest, artifact := range artifacts {
			artifact := artifact
			resource := models.Resource{
				ID:          fmt.Sprintf("%s/%s@%s", regHost, imageName, digest),
				Name:        fmt.Sprintf("%s@%s", imageName, digest),
				Description: artifact,
			}

			if stream != nil {
				if err := (*stream)(resource); err != nil {
					return nil, err
				}
			} else {
				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}
