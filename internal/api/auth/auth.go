package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stax-labs/terraform-provider-stax/internal/api/auth/cognito"
	"github.com/stax-labs/terraform-provider-stax/internal/api/auth/sts"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/client"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/models"
)

type CredentialsRetrieverFn func(ctx context.Context) (aws.Credentials, error)

func AWSCredentialsRetrieverFn(awscfg aws.Config) CredentialsRetrieverFn {
	return func(ctx context.Context) (aws.Credentials, error) {
		return awscfg.Credentials.Retrieve(ctx)
	}
}

// APIToken stax api token used for authentication.
type APIToken struct {
	AccessKey string
	SecretKey string
}

func (a *APIToken) IsValid() bool {
	return a.AccessKey != "" && a.SecretKey != ""
}

// AuthResponse authentication response containing the aws sdk configuration loaded with credentials.
type AuthResponse struct {
	AWSConfig aws.Config
	Tokens    *cognito.IDPTokens
}

func AuthAPIToken(ctx context.Context, client client.ClientWithResponsesInterface, apiToken *APIToken) (*AuthResponse, error) {
	publicConfig, err := getPublicConfig(ctx, client)
	if err != nil {
		return nil, err
	}

	result, err := cognito.Auth(ctx,
		cognito.WithRegion(string(publicConfig.ApiAuth.Region)),
		cognito.WithUserPool(publicConfig.ApiAuth.UserPoolId, publicConfig.ApiAuth.UserPoolWebClientId),
		cognito.WithUsernamePassword(apiToken.AccessKey, apiToken.SecretKey))
	if err != nil {
		return nil, err
	}

	cfg, err := sts.GetCredentials(ctx,
		sts.WithRegion(string(publicConfig.ApiAuth.Region)),
		sts.WithUserPool(publicConfig.ApiAuth.UserPoolId, publicConfig.ApiAuth.IdentityPoolId),
		sts.WithIDToken(*result.IDPTokens.IdToken))
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AWSConfig: cfg,
		Tokens:    result.IDPTokens,
	}, nil
}

func getPublicConfig(ctx context.Context, client client.ClientWithResponsesInterface) (*models.PublicReadConfig, error) {
	res, err := client.PublicReadConfigWithResponse(ctx)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("request failed, returned non 200 status: %s", res.Status())
	}

	return res.JSON200, nil
}
