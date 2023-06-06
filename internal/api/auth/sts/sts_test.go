package sts

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity/types"
	"github.com/stretchr/testify/assert"
)

func TestGetCredentials(t *testing.T) {
	t.Run("gets credentials successfully", func(t *testing.T) {
		userPoolID := "user-pool-id"
		identityPoolID := "identity-pool-id"
		idToken := "id-token"
		region := "us-east-1"

		cisvc := &mockCognitoIdentityClient{
			getIdOutput: &cognitoidentity.GetIdOutput{
				IdentityId: aws.String("identity-id"),
			},
			getCredentialsForIdentityOutput: &cognitoidentity.GetCredentialsForIdentityOutput{
				Credentials: &types.Credentials{
					AccessKeyId:  aws.String("access-key-id"),
					SecretKey:    aws.String("secret-key"),
					SessionToken: aws.String("session-token"),
					Expiration:   aws.Time(time.Now()),
				},
			},
		}

		cfg, err := GetCredentials(context.Background(),
			WithUserPool(userPoolID, identityPoolID),
			WithRegion(region),
			WithIDToken(idToken),
			WithCognitoIdentityClient(cisvc))
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, region, cfg.Region)
		assert.NotNil(t, cfg.Credentials)
	})
}

type mockCognitoIdentityClient struct {
	getIdOutput                     *cognitoidentity.GetIdOutput
	getCredentialsForIdentityOutput *cognitoidentity.GetCredentialsForIdentityOutput
}

func (m *mockCognitoIdentityClient) GetId(ctx context.Context, params *cognitoidentity.GetIdInput, optFns ...func(*cognitoidentity.Options)) (*cognitoidentity.GetIdOutput, error) {
	return m.getIdOutput, nil
}

func (m *mockCognitoIdentityClient) GetCredentialsForIdentity(ctx context.Context, params *cognitoidentity.GetCredentialsForIdentityInput, optFns ...func(*cognitoidentity.Options)) (*cognitoidentity.GetCredentialsForIdentityOutput, error) {
	return m.getCredentialsForIdentityOutput, nil
}
