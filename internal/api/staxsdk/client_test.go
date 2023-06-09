package staxsdk

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stax-labs/terraform-provider-stax/internal/api/auth"
	"github.com/stax-labs/terraform-provider-stax/internal/api/auth/cognito"
	"github.com/stax-labs/terraform-provider-stax/internal/api/mocks"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/client"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
		authFn:   testAuthFn,
	}

	err = client.Authenticate(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if client.authRequestSigner == nil {
		t.Fatal("expected authRequestSigner to be set")
	}
}

func TestClient_PublicReadConfig(t *testing.T) {
	assert := require.New(t)

	ctx := context.Background()
	testClient, clientWithResponsesMock := NewTestClient(t)

	clientWithResponsesMock.On("PublicReadConfigWithResponse", mock.AnythingOfType("*context.emptyCtx")).
		Return(&client.PublicReadConfigResp{
			JSON200:      &models.PublicReadConfig{},
			HTTPResponse: &http.Response{StatusCode: http.StatusOK},
		}, nil)

	publicConfigResp, err := testClient.PublicReadConfig(ctx)
	assert.NoError(err)

	assert.Equal(&models.PublicReadConfig{}, publicConfigResp.JSON200)
}

func TestClient_AccountCreate(t *testing.T) {
	assert := require.New(t)

	ctx := context.Background()
	testClient, clientWithResponsesMock := NewTestClient(t)

	ac := models.AccountsCreateAccount_AccountType{}
	err := ac.FromRoUuidv4("production")
	assert.NoError(err)

	createAccount := models.AccountsCreateAccount{
		Name:        "test-account",
		AccountType: ac,
	}

	clientWithResponsesMock.On("AccountsCreateAccountWithResponse",
		mock.AnythingOfType("*context.emptyCtx"),
		createAccount,
		mock.AnythingOfType("client.RequestEditorFn"),
	).Return(&client.AccountsCreateAccountResp{
		JSON200: &models.AccountsCreateAccountResponse{
			TaskId: aws.String("test"),
		},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}, nil)

	createResp, err := testClient.AccountCreate(ctx, createAccount)
	assert.NoError(err)

	assert.Equal("test", *createResp.JSON200.TaskId)
}

func TestClient_GroupRead(t *testing.T) {
	assert := require.New(t)

	groupId := "b549185e-0fd7-44cf-a7b5-0751c720c0f0"

	ctx := context.Background()
	testClient, clientWithResponsesMock := NewTestClient(t)

	params := &models.TeamsReadGroupsParams{
		IdFilter: aws.String(groupId),
	}

	clientWithResponsesMock.On("TeamsReadGroupsWithResponse",
		mock.AnythingOfType("*context.emptyCtx"),
		params,
		mock.AnythingOfType("client.RequestEditorFn"),
	).Return(&client.TeamsReadGroupsResp{
		JSON200:      &models.TeamsReadGroupsResponse{},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}, nil)

	groupsResp, err := testClient.GroupRead(ctx, []string{groupId})
	assert.NoError(err)
	assert.Equal(&models.TeamsReadGroupsResponse{}, groupsResp.JSON200)
}

func NewTestClient(t *testing.T) (*Client, *mocks.ClientWithResponsesInterface) {

	clientWithResponses := mocks.NewClientWithResponsesInterface(t)

	c := Client{
		client: clientWithResponses,
		authFn: testAuthFn,
	}

	err := c.Authenticate(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	return &c, clientWithResponses
}

func testAuthFn(ctx context.Context, client client.ClientWithResponsesInterface, apiToken *auth.APIToken) (*auth.AuthResponse, error) {
	return &auth.AuthResponse{
		Tokens: &cognito.IDPTokens{
			AccessToken:  aws.String("test"),
			IdToken:      aws.String("test"),
			RefreshToken: aws.String("test"),
		},
	}, nil
}
