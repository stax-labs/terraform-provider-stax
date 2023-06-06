package sts

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/aws/smithy-go/logging"
)

type CognitoIdentityClient interface {
	GetId(ctx context.Context, params *cognitoidentity.GetIdInput, optFns ...func(*cognitoidentity.Options)) (*cognitoidentity.GetIdOutput, error)
	GetCredentialsForIdentity(ctx context.Context, params *cognitoidentity.GetCredentialsForIdentityInput, optFns ...func(*cognitoidentity.Options)) (*cognitoidentity.GetCredentialsForIdentityOutput, error)
}

type CredsConfig struct {
	region         string
	userPoolID     string
	identityPoolID string
	idToken        string
	cisvc          CognitoIdentityClient
	logger         logging.Logger
}

type CredsOption func(*CredsConfig)

func WithUserPool(userPoolID string, identityPoolID string) CredsOption {
	return func(cfg *CredsConfig) {
		cfg.userPoolID = userPoolID
		cfg.identityPoolID = identityPoolID
	}
}

func WithRegion(region string) CredsOption {
	return func(cfg *CredsConfig) {
		cfg.region = region
	}
}

func WithIDToken(idToken string) CredsOption {
	return func(cfg *CredsConfig) {
		cfg.idToken = idToken
	}
}

func WithCognitoIdentityClient(cisvc CognitoIdentityClient) CredsOption {
	return func(cfg *CredsConfig) {
		cfg.cisvc = cisvc
	}
}

func WithLogger(logger logging.Logger) CredsOption {
	return func(cfg *CredsConfig) {
		cfg.logger = logger
	}
}

func GetCredentials(ctx context.Context, opts ...CredsOption) (aws.Config, error) {

	cfg := CredsConfig{}

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
		return aws.Config{}, err
	}

	if cfg.cisvc == nil {
		cfg.cisvc = cognitoidentity.NewFromConfig(awscfg)
	}

	logins := map[string]string{
		fmt.Sprintf("cognito-idp.%s.amazonaws.com/%s", cfg.region, cfg.userPoolID): cfg.idToken,
	}

	getIdRes, err := cfg.cisvc.GetId(ctx, &cognitoidentity.GetIdInput{
		IdentityPoolId: aws.String(cfg.identityPoolID),
		Logins:         logins,
	})
	if err != nil {
		return aws.Config{}, err
	}

	credsRes, err := cfg.cisvc.GetCredentialsForIdentity(ctx, &cognitoidentity.GetCredentialsForIdentityInput{
		IdentityId: getIdRes.IdentityId,
		Logins:     logins,
	})
	if err != nil {
		return aws.Config{}, err
	}

	return config.LoadDefaultConfig(
		ctx,
		config.WithRegion(cfg.region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			aws.ToString(credsRes.Credentials.AccessKeyId),
			aws.ToString(credsRes.Credentials.SecretKey),
			aws.ToString(credsRes.Credentials.SessionToken),
		)),
	)
}
