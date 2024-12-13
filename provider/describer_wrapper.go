package provider

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	model "github.com/opengovern/og-describer-oci/pkg/sdk/models"
	"github.com/opengovern/og-describer-oci/provider/configs"
	"github.com/opengovern/og-describer-oci/provider/describer"
	"github.com/opengovern/og-util/pkg/describe/enums"
	configs2 "github.com/opengovern/opencomply/services/integration/integration-type/oci-repository/configs"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"net/url"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
	"strings"
	"time"
)

type AuthConfig struct {
	Registry string
	Username string
	Password string
}

func httpPostForm(ctx context.Context, urlStr string, data url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", urlStr, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST %s failed: %w", urlStr, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("non-2xx status: %d body: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func getDockerhubAuth(username, password string) (*AuthConfig, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("missing required Dockerhub credentials")
	}

	return &AuthConfig{
		Registry: "docker.io",
		Username: username,
		Password: password,
	}, nil
}

func getGHCRAuth(username, token, owner string) (*AuthConfig, error) {
	if username == "" || token == "" || owner == "" {
		return nil, fmt.Errorf("missing required GHCR credentials")
	}

	return &AuthConfig{
		Registry: fmt.Sprintf("ghcr.io/%s", owner),
		Username: username,
		Password: token,
	}, nil
}

func getECRAuth(accessKey, secretKey, accountID, region string) (*AuthConfig, error) {
	if accountID == "" || region == "" {
		return nil, fmt.Errorf("missing required ECR credentials")
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := ecr.NewFromConfig(cfg)
	resp, err := client.GetAuthorizationToken(context.Background(), &ecr.GetAuthorizationTokenInput{
		RegistryIds: []string{accountID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ECR auth token: %w", err)
	}

	if len(resp.AuthorizationData) == 0 || resp.AuthorizationData[0].AuthorizationToken == nil {
		return nil, fmt.Errorf("no authorization token received from ECR")
	}

	authData := resp.AuthorizationData[0]
	decoded, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decode auth token: %w", err)
	}

	registry := *authData.ProxyEndpoint
	registry = strings.TrimPrefix(registry, "https://")

	authStr := string(decoded) // "AWS:<token>"

	username, password := strings.Split(authStr, ":")[0], strings.Split(authStr, ":")[1]

	return &AuthConfig{
		Registry: registry,
		Username: username,
		Password: password,
	}, nil
}

func getACRAuth(loginServer, tenantID, clientID, clientSecret string) (*AuthConfig, error) {
	if loginServer == "" || tenantID == "" {
		return nil, fmt.Errorf("missing required ACR credentials")
	}

	cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get default azure credential: %w", err)
	}
	ctx := context.Background()
	aadToken, err := cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{"https://management.azure.com/.default"}})
	if err != nil {
		return nil, fmt.Errorf("failed to get AAD token: %w", err)
	}

	refreshToken, err := getACRRefreshToken(ctx, loginServer, tenantID, aadToken.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to get ACR refresh token: %w", err)
	}

	accessToken, err := getACRAccessToken(ctx, loginServer, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get ACR access token: %w", err)
	}

	return &AuthConfig{
		Registry: loginServer,
		Username: "00000000-0000-0000-0000-000000000000",
		Password: accessToken,
	}, nil
}

func getACRRefreshToken(ctx context.Context, acrService, tenantID, aadAccessToken string) (string, error) {
	formData := url.Values{
		"grant_type":   {"access_token"},
		"service":      {acrService},
		"tenant":       {tenantID},
		"access_token": {aadAccessToken},
	}

	urlStr := fmt.Sprintf("https://%s/oauth2/exchange", acrService)
	respBody, err := httpPostForm(ctx, urlStr, formData)
	if err != nil {
		return "", err
	}
	var response map[string]interface{}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("invalid exchange response: %w", err)
	}
	refreshToken, ok := response["refresh_token"].(string)
	if !ok || refreshToken == "" {
		return "", fmt.Errorf("no refresh_token in ACR exchange response")
	}
	return refreshToken, nil
}

func getACRAccessToken(ctx context.Context, acrService, refreshToken string) (string, error) {
	formData := url.Values{
		"grant_type":    {"refresh_token"},
		"service":       {acrService},
		"refresh_token": {refreshToken},
		"scope":         {"repository:*:pull,push"},
	}

	urlStr := fmt.Sprintf("https://%s/oauth2/token", acrService)
	respBody, err := httpPostForm(ctx, urlStr, formData)
	if err != nil {
		return "", err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("invalid token response: %w", err)
	}
	accessToken, ok := response["access_token"].(string)
	if !ok || accessToken == "" {
		return "", fmt.Errorf("no access_token in response")
	}
	return accessToken, nil
}

// DescribeByIntegration TODO: implement a wrapper to pass integration authorization to describer functions
func DescribeByIntegration(describe func(context.Context, *configs.IntegrationCredentials, string, *model.StreamSender) ([]model.Resource, error)) model.ResourceDescriber {
	return func(ctx context.Context, cfg configs.IntegrationCredentials, triggerType enums.DescribeTriggerType, additionalParameters map[string]string, stream *model.StreamSender) ([]model.Resource, error) {
		var creds *AuthConfig
		var err error
		switch cfg.RegistryType {
		case configs2.RegistryTypeDockerhub:
			creds, err = getDockerhubAuth(cfg.DockerhubCredentials.Username, cfg.DockerhubCredentials.Password)
		case configs2.RegistryTypeGHCR:
			creds, err = getGHCRAuth(cfg.GhcrCredentials.Username, cfg.GhcrCredentials.Token, cfg.GhcrCredentials.Owner)
			ctx = describer.WithOwner(ctx, cfg.GhcrCredentials.Owner)

		case configs2.RegistryTypeECR:
			creds, err = getECRAuth(cfg.EcrCredentials.AccessKey, cfg.EcrCredentials.SecretKey, cfg.EcrCredentials.AccountID, cfg.EcrCredentials.Region)
		case configs2.RegistryTypeACR:
			creds, err = getACRAuth(cfg.AcrCredentials.LoginServer, cfg.AcrCredentials.TenantID, cfg.AcrCredentials.ClientID, cfg.AcrCredentials.ClientSecret)
		}
		if err != nil {
			return nil, err
		}

		remoteClient := &auth.Client{
			Client: retry.DefaultClient,
			Cache:  auth.NewCache(),
			Credential: auth.StaticCredential(creds.Registry, auth.Credential{
				Username: creds.Username,
				Password: creds.Password,
			}),
		}

		ctx = describer.WithOrasClient(ctx, remoteClient)
		ctx = describer.WithRegistry(ctx, creds.Registry)

		return describe(ctx, &cfg, string(triggerType), stream)
	}
}
