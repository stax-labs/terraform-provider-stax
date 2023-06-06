package staxsdk

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stax-labs/terraform-provider-stax/internal/api/auth"
	"github.com/stax-labs/terraform-provider-stax/internal/api/auth/cognito"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/client"
)

func TestClient_Authenticate(t *testing.T) {
	ctx := context.Background()

	apiToken := &auth.APIToken{
		AccessKey: "accessKey",
		SecretKey: "secretKey",
	}

	sdkClient, err := client.NewClientWithResponses("http://localhost")
	if err != nil {
		t.Fatal(err)
	}

	client := &Client{
		apiToken: apiToken,
		client:   sdkClient,
		authFn: func(ctx context.Context, client client.ClientWithResponsesInterface, apiToken *auth.APIToken) (*auth.AuthResponse, error) {
			return &auth.AuthResponse{
				Tokens: &cognito.IDPTokens{
					AccessToken:  aws.String("test"),
					IdToken:      aws.String("test"),
					RefreshToken: aws.String("test"),
				},
			}, nil
		},
	}

	err = client.Authenticate(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if client.authRequestSigner == nil {
		t.Fatal("expected authRequestSigner to be set")
	}
}
