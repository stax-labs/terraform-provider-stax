package cognito

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	t.Run("gets auth successfully", func(t *testing.T) {
		userPoolID := "eu-west-1_myPool"
		userPoolWebClientID := "123abd"
		username := "username"
		password := "password"
		region := "us-east-1"

		cipsvc := &mockCognitoIdentityProviderClient{
			initiateAuthOutput: &cognitoidentityprovider.InitiateAuthOutput{
				ChallengeName: types.ChallengeNameTypePasswordVerifier,
			},
			respondToAuthChallengeOutput: &cognitoidentityprovider.RespondToAuthChallengeOutput{
				AuthenticationResult: &types.AuthenticationResultType{
					AccessToken:  aws.String("access-token"),
					ExpiresIn:    3600,
					IdToken:      aws.String("id-token"),
					RefreshToken: aws.String("refresh-token"),
				},
			},
		}

		csrp := &mockCognitoSRP{}

		result, err := Auth(context.Background(),
			WithUserPool(userPoolID, userPoolWebClientID),
			WithUsernamePassword(username, password),
			WithRegion(region),
			WithCognitoIdentityProviderClient(cipsvc),
			WithCognitoSRP(csrp))
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.IDPTokens)
		assert.Equal(t, "access-token", aws.ToString(result.IDPTokens.AccessToken))
		assert.Equal(t, int32(3600), result.IDPTokens.ExpiresIn)
		assert.Equal(t, "id-token", aws.ToString(result.IDPTokens.IdToken))
		assert.Equal(t, "refresh-token", aws.ToString(result.IDPTokens.RefreshToken))
	})
}

type mockCognitoIdentityProviderClient struct {
	initiateAuthOutput           *cognitoidentityprovider.InitiateAuthOutput
	respondToAuthChallengeOutput *cognitoidentityprovider.RespondToAuthChallengeOutput
}

func (m *mockCognitoIdentityProviderClient) InitiateAuth(ctx context.Context, params *cognitoidentityprovider.InitiateAuthInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.InitiateAuthOutput, error) {
	return m.initiateAuthOutput, nil
}

func (m *mockCognitoIdentityProviderClient) RespondToAuthChallenge(ctx context.Context, params *cognitoidentityprovider.RespondToAuthChallengeInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.RespondToAuthChallengeOutput, error) {
	return m.respondToAuthChallengeOutput, nil
}

type mockCognitoSRP struct {
	clientID      string
	authParams    map[string]string
	challengeResp map[string]string
}

func (m *mockCognitoSRP) GetClientId() string {
	return m.clientID
}

func (m *mockCognitoSRP) GetAuthParams() map[string]string {
	return m.authParams
}

func (m *mockCognitoSRP) PasswordVerifierChallenge(challengeParms map[string]string, ts time.Time) (map[string]string, error) {
	return m.challengeResp, nil
}
