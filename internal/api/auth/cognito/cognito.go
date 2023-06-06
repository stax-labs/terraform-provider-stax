package cognito

import (
	"context"
	"fmt"
	"time"

	cognitosrp "github.com/alexrudd/cognito-srp/v4"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/aws/smithy-go/logging"
)

type CognitoIdentityProviderClient interface {
	InitiateAuth(ctx context.Context, params *cognitoidentityprovider.InitiateAuthInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.InitiateAuthOutput, error)
	RespondToAuthChallenge(ctx context.Context, params *cognitoidentityprovider.RespondToAuthChallengeInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.RespondToAuthChallengeOutput, error)
}

type CognitoSRP interface {
	GetClientId() string
	GetAuthParams() map[string]string
	PasswordVerifierChallenge(challengeParms map[string]string, ts time.Time) (map[string]string, error)
}

// AuthCognitoResult auth config results.
type AuthResult struct {
	IDPTokens *IDPTokens
	AWSConfig aws.Config
}

// IDPTokens contains the cognito "oidc" tokens.
type IDPTokens struct {
	// The access token.
	AccessToken *string

	// The expiration period of the authentication result in seconds.
	ExpiresIn int32

	// The ID token.
	IdToken *string

	// The refresh token.
	RefreshToken *string

	// The token type.
	TokenType *string
}

type CognitoConfig struct {
	region              string
	userPoolID          string
	userPoolWebClientID string
	username, password  string
	cipsvc              CognitoIdentityProviderClient
	csrp                CognitoSRP
	logger              logging.Logger
}

// AuthCognitoOption configures the AuthCognito function.
type AuthOption func(*CognitoConfig)

// WithRegion sets the AWS region.
func WithRegion(region string) AuthOption {
	return func(cfg *CognitoConfig) {
		cfg.region = region
	}
}

// WithUserPoolID sets the Cognito user pool ID and the Cognito user pool web client ID.
func WithUserPool(userPoolID string, userPoolWebClientID string) AuthOption {
	return func(cfg *CognitoConfig) {
		cfg.userPoolID = userPoolID
		cfg.userPoolWebClientID = userPoolWebClientID
	}
}

// WithUsername sets the username and password.
func WithUsernamePassword(username, password string) AuthOption {
	return func(cfg *CognitoConfig) {
		cfg.username = username
		cfg.password = password
	}
}

// WithCognitoIdentityProviderClient sets the CognitoIdentityProvider client.
func WithCognitoIdentityProviderClient(cipsvc CognitoIdentityProviderClient) AuthOption {
	return func(cfg *CognitoConfig) {
		cfg.cipsvc = cipsvc
	}
}

func WithCognitoSRP(csrp CognitoSRP) AuthOption {
	return func(cfg *CognitoConfig) {
		cfg.csrp = csrp
	}
}

func WithLogger(logger logging.Logger) AuthOption {
	return func(cfg *CognitoConfig) {
		cfg.logger = logger
	}
}

func Auth(ctx context.Context, opts ...AuthOption) (*AuthResult, error) {
	cfg := CognitoConfig{}

	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.logger == nil {
		cfg.logger = logging.Nop{}
	}

	awscfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(cfg.region),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
		config.WithLogger(cfg.logger),
		config.WithClientLogMode(aws.LogRetries|aws.LogRequest),
	)
	if err != nil {
		return nil, err
	}

	if cfg.cipsvc == nil {
		cfg.cipsvc = cognitoidentityprovider.NewFromConfig(awscfg)
	}

	if cfg.csrp == nil {
		cfg.csrp, err = cognitosrp.NewCognitoSRP(cfg.username, cfg.password, cfg.userPoolID, cfg.userPoolWebClientID, nil)
		if err != nil {
			return nil, err
		}
	}

	resp, err := cfg.cipsvc.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow:       types.AuthFlowTypeUserSrpAuth,
		ClientId:       aws.String(cfg.csrp.GetClientId()),
		AuthParameters: cfg.csrp.GetAuthParams(),
	})
	if err != nil {
		return nil, err
	}

	if resp.ChallengeName != types.ChallengeNameTypePasswordVerifier {
		return nil, fmt.Errorf("failed authentication, unhandled challenge: %s", resp.ChallengeName)
	}

	challengeResponses, err := cfg.csrp.PasswordVerifierChallenge(resp.ChallengeParameters, time.Now())
	if err != nil {
		return nil, err
	}

	authResp, err := cfg.cipsvc.RespondToAuthChallenge(ctx, &cognitoidentityprovider.RespondToAuthChallengeInput{
		ChallengeName:      types.ChallengeNameTypePasswordVerifier,
		ChallengeResponses: challengeResponses,
		ClientId:           aws.String(cfg.csrp.GetClientId()),
	})
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		AWSConfig: awscfg,
		IDPTokens: &IDPTokens{
			IdToken:      authResp.AuthenticationResult.IdToken,
			AccessToken:  authResp.AuthenticationResult.AccessToken,
			RefreshToken: authResp.AuthenticationResult.RefreshToken,
			ExpiresIn:    authResp.AuthenticationResult.ExpiresIn,
			TokenType:    authResp.AuthenticationResult.TokenType,
		},
	}, nil
}
