package staxsdk

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stax-labs/terraform-provider-stax/internal/api/auth"
	"github.com/stax-labs/terraform-provider-stax/internal/api/auth/cognito"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/client"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/mocks"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/models"
	permissionssetsclient "github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/client"
	permissionssetsmocks "github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/mocks"
	permissionssetsmodels "github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/models"
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

	testClient, clientWithResponsesMock := NewTestClient(t)

	clientWithResponsesMock.On("PublicReadConfigWithResponse", mock.AnythingOfType("*context.emptyCtx")).
		Return(&client.PublicReadConfigResp{
			JSON200:      &models.PublicReadConfig{},
			HTTPResponse: &http.Response{StatusCode: http.StatusOK},
		}, nil)

	publicConfigResp, err := testClient.PublicReadConfig(context.TODO())
	assert.NoError(err)

	assert.Equal(&models.PublicReadConfig{}, publicConfigResp.JSON200)
}

func TestClient_AccountReadByID(t *testing.T) {
	assert := require.New(t)
	accountID := "b549185e-0fd7-44cf-a7b5-0751c720c0f0"

	testClient, clientWithResponsesMock := NewTestClient(t)

	accounts := &models.AccountsReadAccounts{
		Accounts: []models.Account{
			{Id: &accountID},
		},
	}

	clientWithResponsesMock.On("AccountsReadAccountWithResponse",
		mock.AnythingOfType("*context.emptyCtx"),
		accountID,
		mock.AnythingOfType("*models.AccountsReadAccountParams"),
		mock.AnythingOfType("client.RequestEditorFn"),
	).Return(&client.AccountsReadAccountResp{
		JSON200:      accounts,
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}, nil)

	accountResp, err := testClient.AccountReadByID(context.TODO(), accountID)
	assert.NoError(err)
	assert.Equal(accounts, accountResp.JSON200)
}

func TestClient_AccountCreate(t *testing.T) {
	assert := require.New(t)

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

	createResp, err := testClient.AccountCreate(context.TODO(), createAccount)
	assert.NoError(err)

	assert.Equal("test", *createResp.JSON200.TaskId)
}

func TestClient_AccountTypeReadById(t *testing.T) {
	assert := require.New(t)
	accountTypeID := "b549185e-0fd7-44cf-a7b5-0751c720c0f0"

	testClient, clientWithResponsesMock := NewTestClient(t)

	accountTypes := &models.AccountsReadAccountTypes{
		AccountTypes: []models.AccountType{
			{Id: &accountTypeID},
		},
	}

	clientWithResponsesMock.On("AccountsReadAccountTypeWithResponse",
		mock.AnythingOfType("*context.emptyCtx"),
		accountTypeID,
		mock.AnythingOfType("*models.AccountsReadAccountTypeParams"),
		mock.AnythingOfType("client.RequestEditorFn"),
	).Return(&client.AccountsReadAccountTypeResp{
		JSON200:      accountTypes,
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}, nil)

	accountTypeResp, err := testClient.AccountTypeReadById(context.TODO(), accountTypeID)
	assert.NoError(err)
	assert.Equal(accountTypes, accountTypeResp.JSON200)
}

func TestClient_GroupRead(t *testing.T) {
	assert := require.New(t)

	groupId := "b549185e-0fd7-44cf-a7b5-0751c720c0f0"

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

	groupsResp, err := testClient.GroupRead(context.TODO(), []string{groupId})
	assert.NoError(err)
	assert.Equal(&models.TeamsReadGroupsResponse{}, groupsResp.JSON200)
}

func TestClient_GroupCreate(t *testing.T) {
	assert := require.New(t)
	groupId := "b549185e-0fd7-44cf-a7b5-0751c720c0f0"
	groupName := "group-new-name"

	testClient, clientWithResponsesMock := NewTestClient(t)

	clientWithResponsesMock.On("TeamsCreateGroupWithResponse",
		mock.AnythingOfType("*context.emptyCtx"),
		models.TeamsCreateGroup{Name: groupName},
		mock.AnythingOfType("client.RequestEditorFn"),
	).Return(&client.TeamsCreateGroupResp{
		JSON200:      &models.TeamsCreateGroupEvent{GroupId: &groupId},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}, nil)

	createResp, err := testClient.GroupCreate(context.TODO(), groupName)
	assert.NoError(err)
	assert.Equal(&models.TeamsCreateGroupEvent{GroupId: &groupId}, createResp.JSON200)
}

func TestClient_GroupUpdate(t *testing.T) {
	assert := require.New(t)
	groupName := "group-updated-name"
	groupId := "b549185e-0fd7-44cf-a7b5-0751c720c0f0"

	testClient, clientWithResponsesMock := NewTestClient(t)

	clientWithResponsesMock.On("TeamsUpdateGroupWithResponse",
		mock.AnythingOfType("*context.emptyCtx"),
		groupId,
		models.TeamsUpdateGroup{Name: groupName},
		mock.AnythingOfType("client.RequestEditorFn"),
	).Return(&client.TeamsUpdateGroupResp{
		JSON200:      &models.TeamsUpdateGroupEvent{},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}, nil)

	updateResp, err := testClient.GroupUpdate(context.TODO(), groupId, groupName)
	assert.NoError(err)
	assert.Equal(&models.TeamsUpdateGroupEvent{}, updateResp.JSON200)
}

func TestClient_GroupUpdate_AddMembers(t *testing.T) {
	assert := require.New(t)
	groupId := "b549185e-0fd7-44cf-a7b5-0751c720c0f0"

	addMemberUserIDs := []string{
		"95319c06-db97-481d-b355-7ff093d0c867",
		"18bd55b8-4785-41f8-b6df-d682eefd89b7",
	}

	testClient, clientWithResponsesMock := NewTestClient(t)

	clientWithResponsesMock.On("TeamsUpdateGroupMembersWithResponse",
		mock.AnythingOfType("*context.emptyCtx"),
		models.TeamsUpdateGroupMembers{
			AddMembers: &[]models.UserGroupMemberMap{
				{
					GroupId: groupId,
					UserId:  addMemberUserIDs[0],
				},
				{
					GroupId: groupId,
					UserId:  addMemberUserIDs[1],
				},
			},
		},
		mock.AnythingOfType("client.RequestEditorFn"),
	).Return(&client.TeamsUpdateGroupMembersResp{
		JSON200:      &models.TeamsUpdateGroupMembersEvent{},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}, nil)

	updateResp, err := testClient.GroupAssignUsers(context.TODO(), groupId, addMemberUserIDs, nil)
	assert.NoError(err)
	assert.Equal(&models.TeamsUpdateGroupMembersEvent{}, updateResp.JSON200)
}

func TestClient_GroupDelete(t *testing.T) {
	assert := require.New(t)
	groupId := "b549185e-0fd7-44cf-a7b5-0751c720c0f0"

	testClient, clientWithResponsesMock := NewTestClient(t)

	clientWithResponsesMock.On("TeamsDeleteGroupWithResponse",
		mock.AnythingOfType("*context.emptyCtx"),
		groupId,
		mock.AnythingOfType("client.RequestEditorFn"),
	).Return(&client.TeamsDeleteGroupResp{
		JSON200:      &models.TeamsDeleteGroupEvent{},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}, nil)

	deleteResp, err := testClient.GroupDelete(context.TODO(), groupId)
	assert.NoError(err)
	assert.Equal(&models.TeamsDeleteGroupEvent{}, deleteResp.JSON200)
}

func TestClient_UserCreate(t *testing.T) {
	assert := require.New(t)
	taskId := "b549185e-0fd7-44cf-a7b5-0751c720c0f0"
	email := "test@example.com"

	readonlyRole := models.IdamUserRole(models.Readonly)

	params := models.TeamsCreateUser{
		Email:     openapi_types.Email(email),
		FirstName: "Test",
		LastName:  "Test",
		Role:      &readonlyRole,
	}

	testClient, clientWithResponsesMock := NewTestClient(t)

	clientWithResponsesMock.On("TeamsCreateUserWithResponse",
		mock.AnythingOfType("*context.emptyCtx"),
		params,
		mock.AnythingOfType("client.RequestEditorFn"),
	).Return(&client.TeamsCreateUserResp{
		JSON200:      &models.TeamsCreateUserEvent{TaskId: &taskId},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}, nil)

	userResp, err := testClient.UserCreate(context.TODO(), params)
	assert.NoError(err)
	assert.Equal(&models.TeamsCreateUserEvent{TaskId: &taskId}, userResp.JSON200)
}

func TestClient_PermissionSetsReadByID(t *testing.T) {
	assert := require.New(t)
	permissionSetID := "b549185e-0fd7-44cf-a7b5-0751c720c0f0"

	testClient, clientWithResponsesMock := NewTestPermissionSetsClient(t)

	permissionSetRecord := &permissionssetsmodels.PermissionSetRecord{
		Id: uuid.MustParse(permissionSetID),
	}

	clientWithResponsesMock.On("GetPermissionSetWithResponse",
		mock.AnythingOfType("*context.emptyCtx"),
		uuid.MustParse(permissionSetID),
		mock.AnythingOfType("client.RequestEditorFn"),
	).Return(&permissionssetsclient.GetPermissionSetResponse{
		JSON200:      permissionSetRecord,
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}, nil)

	permissionSetResp, err := testClient.PermissionSetsReadByID(context.TODO(), permissionSetID)
	assert.NoError(err)
	assert.Equal(permissionSetRecord, permissionSetResp.JSON200)
}

func TestClient_PermissionSetsList(t *testing.T) {
	assert := require.New(t)
	permissionSetID := "b549185e-0fd7-44cf-a7b5-0751c720c0f0"

	testClient, clientWithResponsesMock := NewTestPermissionSetsClient(t)

	params := &permissionssetsmodels.ListPermissionSetsParams{}

	permissionSetRecord := permissionssetsmodels.PermissionSetRecord{
		Id: uuid.MustParse(permissionSetID),
	}

	clientWithResponsesMock.On("ListPermissionSetsWithResponse",
		mock.AnythingOfType("*context.emptyCtx"),
		params,
		mock.AnythingOfType("client.RequestEditorFn"),
	).Return(&permissionssetsclient.ListPermissionSetsResponse{
		JSON200: &permissionssetsmodels.ListPermissionSets{
			PermissionSets: []permissionssetsmodels.PermissionSetRecord{
				permissionSetRecord,
			},
		},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}, nil)

	permissionSetResp, err := testClient.PermissionSetsList(context.TODO(), params)
	assert.NoError(err)
	assert.Equal(&permissionssetsmodels.ListPermissionSets{PermissionSets: []permissionssetsmodels.PermissionSetRecord{permissionSetRecord}}, permissionSetResp.JSON200)
}

func TestClient_PermissionSetsCreate(t *testing.T) {
	assert := require.New(t)
	permissionSetID := "b549185e-0fd7-44cf-a7b5-0751c720c0f0"

	testClient, clientWithResponsesMock := NewTestPermissionSetsClient(t)

	params := permissionssetsmodels.CreatePermissionSetRecord{}

	permissionSetRecord := &permissionssetsmodels.PermissionSetRecord{
		Id: uuid.MustParse(permissionSetID),
	}

	clientWithResponsesMock.On("CreatePermissionSetWithResponse",
		mock.AnythingOfType("*context.emptyCtx"),
		params,
		mock.AnythingOfType("client.RequestEditorFn"),
	).Return(&permissionssetsclient.CreatePermissionSetResponse{
		JSON201:      permissionSetRecord,
		HTTPResponse: &http.Response{StatusCode: http.StatusCreated},
	}, nil)

	permissionSetResp, err := testClient.PermissionSetsCreate(context.TODO(), params)
	assert.NoError(err)
	assert.Equal(permissionSetRecord, permissionSetResp.JSON201)
}

func TestClient_PermissionSetsUpdate(t *testing.T) {
	assert := require.New(t)
	permissionSetID := "b549185e-0fd7-44cf-a7b5-0751c720c0f0"

	testClient, clientWithResponsesMock := NewTestPermissionSetsClient(t)

	params := permissionssetsmodels.UpdatePermissionSetRecord{}

	permissionSetRecord := &permissionssetsmodels.PermissionSetRecord{
		Id: uuid.MustParse(permissionSetID),
	}

	clientWithResponsesMock.On("UpdatePermissionSetWithResponse",
		mock.AnythingOfType("*context.emptyCtx"),
		uuid.MustParse(permissionSetID),
		params,
		mock.AnythingOfType("client.RequestEditorFn"),
	).Return(&permissionssetsclient.UpdatePermissionSetResponse{
		JSON200:      permissionSetRecord,
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}, nil)

	permissionSetResp, err := testClient.PermissionSetsUpdate(context.TODO(), permissionSetID, params)
	assert.NoError(err)
	assert.Equal(permissionSetRecord, permissionSetResp.JSON200)
}

func TestClient_PermissionSetsDelete(t *testing.T) {
	assert := require.New(t)
	permissionSetID := "b549185e-0fd7-44cf-a7b5-0751c720c0f0"

	testClient, clientWithResponsesMock := NewTestPermissionSetsClient(t)

	permissionSetRecord := &permissionssetsmodels.PermissionSetRecord{
		Id: uuid.MustParse(permissionSetID),
	}

	clientWithResponsesMock.On("DeletePermissionSetWithResponse",
		mock.AnythingOfType("*context.emptyCtx"),
		uuid.MustParse(permissionSetID),
		mock.AnythingOfType("client.RequestEditorFn"),
	).Return(&permissionssetsclient.DeletePermissionSetResponse{
		JSON200:      permissionSetRecord,
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}, nil)

	permissionSetResp, err := testClient.PermissionSetsDelete(context.TODO(), permissionSetID)
	assert.NoError(err)
	assert.Equal(permissionSetRecord, permissionSetResp.JSON200)
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

func NewTestPermissionSetsClient(t *testing.T) (*Client, *permissionssetsmocks.ClientWithResponsesInterface) {

	clientWithResponses := permissionssetsmocks.NewClientWithResponsesInterface(t)

	c := Client{
		permissionSetsClient: clientWithResponses,
		authFn:               testAuthFn,
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

func TestGetInstallationCoreAPIURL(t *testing.T) {
	tests := []struct {
		name                 string
		installation         string
		overrideEndpointURLs installationURLs
		want                 *installationURLs
		wantErr              bool
	}{
		{
			name:         "Valid AU1 installation",
			installation: "au1",
			want: &installationURLs{
				CoreAPIEndpointURL:        "https://api.au1.staxapp.cloud",
				PermissionSetsEndpointURL: "https://api.idam.au1.staxapp.cloud/20210321",
			},
			wantErr: false,
		},
		{
			name:         "Valid US1 installation",
			installation: "us1",
			want: &installationURLs{
				CoreAPIEndpointURL:        "https://api.us1.staxapp.cloud",
				PermissionSetsEndpointURL: "https://api.idam.us1.staxapp.cloud/20210321",
			},
			wantErr: false,
		},
		{
			name:         "Valid EU1 installation",
			installation: "eu1",
			want: &installationURLs{
				CoreAPIEndpointURL:        "https://api.eu1.staxapp.cloud",
				PermissionSetsEndpointURL: "https://api.idam.eu1.staxapp.cloud/20210321",
			},
			wantErr: false,
		},
		{
			name:         "Invalid installation",
			installation: "invalid",
			want:         nil,
			wantErr:      true,
		},
		{
			name:                 "Endpoint URL configured",
			overrideEndpointURLs: installationURLs{CoreAPIEndpointURL: "https://example.com", PermissionSetsEndpointURL: "https://idam.example.com"},
			want:                 &installationURLs{CoreAPIEndpointURL: "https://example.com", PermissionSetsEndpointURL: "https://idam.example.com"},
			wantErr:              false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)
			got, err := getInstallationURL(tt.installation, tt.overrideEndpointURLs)
			if (err != nil) != tt.wantErr {
				t.Errorf("getInstallationURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(tt.want, got)
		})
	}
}
