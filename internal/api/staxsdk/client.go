//  Package staxsdk contains a client for interacting with the Stax API.

package staxsdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/uuid"
	"github.com/stax-labs/terraform-provider-stax/internal/api/apigw"
	"github.com/stax-labs/terraform-provider-stax/internal/api/auth"
	"github.com/stax-labs/terraform-provider-stax/internal/api/helpers"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/client"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/models"
	permissionssetsclient "github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/client"
	permissionssetsmodels "github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/models"
	"golang.org/x/exp/slices"
)

const (
	//  TaskStarted indicates an asynchronous task has started.
	TaskStarted = models.OperationStatus("STARTED")

	// TaskSucceeded indicates an asynchronous task completed successfully.
	TaskSucceeded = models.OperationStatus("SUCCEEDED")

	// TaskFailed indicates an asynchronous task failed.
	TaskFailed = models.OperationStatus("FAILED")
)

var (
	//  ErrAuthSessionEmpty is returned when an API call is made without first calling Authenticate.
	ErrAuthSessionEmpty = errors.New("auth session is empty, please call Authenticate to login")

	// ErrMissingAPIToken is returned when the client is constructed without an API token.
	ErrMissingAPIToken = errors.New("missing required api token, must provide WithAPIToken")

	//  ErrMissingTaskID is returned when MonitorTask is called without a task ID.
	ErrMissingTaskID = errors.New("missing task id, provided value is empty")

	//  ErrMissingTaskCallbackFunc is returned when MonitorTask is called without a callback function.
	ErrMissingTaskCallbackFunc = errors.New("missing task monitoring callback function")

	ErrInvalidInstallation = errors.New("invalid installation, url is unknown")
)

// ClientInterface defines the interface for interacting with the Stax API.
type ClientInterface interface {
	// Authenticate authenticates the client and returns an error.
	Authenticate(ctx context.Context) error
	// PublicReadConfig reads the public configuration and returns a client.PublicReadConfigResp.
	PublicReadConfig(ctx context.Context) (*client.PublicReadConfigResp, error)
	// AccountCreate creates an account and returns a SyncResult containing the response and final task status.
	AccountCreate(ctx context.Context, createAccount models.AccountsCreateAccount) (*client.AccountsCreateAccountResp, error)
	// AccountReadByID reads an account by ID and returns a client.AccountsReadAccountResp.
	AccountReadByID(ctx context.Context, accountID string) (*client.AccountsReadAccountResp, error)
	// AccountRead reads accounts and returns a client.AccountsReadAccountsResp.
	AccountRead(ctx context.Context, accountIDs []string, accountNames []string) (*client.AccountsReadAccountsResp, error)
	// AccountUpdate updates an account and returns a SyncResult containing the response and final task status.
	AccountUpdate(ctx context.Context, accountID string, updateAccount models.AccountsUpdateAccount) (*client.AccountsUpdateAccountResp, error)
	// AccountClose closes an account and returns a SyncResult containing the response and final task status.
	AccountClose(ctx context.Context, accountID string) (*client.AccountsCloseAccountResp, error)
	// AccountTypeCreate creates an account type and returns a client.AccountsCreateAccountTypeResp.
	AccountTypeCreate(ctx context.Context, name string) (*client.AccountsCreateAccountTypeResp, error)
	// AccountTypeUpdate updates an account type and returns a client.AccountsUpdateAccountTypeResp.
	AccountTypeUpdate(ctx context.Context, accountTypeID, name string) (*client.AccountsUpdateAccountTypeResp, error)
	//  AccountTypeDelete deletes an account type and returns a client.AccountsDeleteAccountTypeResp.
	AccountTypeDelete(ctx context.Context, accountTypeID string) (*client.AccountsDeleteAccountTypeResp, error)
	// AccountTypeReadById reads an account type by ID and returns a client.AccountsReadAccountTypeResp.
	AccountTypeReadById(ctx context.Context, accountTypeID string) (*client.AccountsReadAccountTypeResp, error)
	// AccountTypeRead reads account types and returns a client.AccountsReadAccountTypesResp.
	AccountTypeRead(ctx context.Context, accountTypeIDs []string) (*client.AccountsReadAccountTypesResp, error)
	// WorkloadDelete deletes a workload and returns a client.WorkloadsDeleteWorkloadResp.
	WorkloadDelete(ctx context.Context, workloadID string) (*client.WorkloadsDeleteWorkloadResp, error)
	// GroupCreate create a group and returns a client.TeamsCreateGroupResp.
	GroupCreate(ctx context.Context, name string) (*client.TeamsCreateGroupResp, error)
	//  GroupUpdate updates a group and returns a client.TeamsUpdateGroupResp.
	GroupUpdate(ctx context.Context, groupID, name string) (*client.TeamsUpdateGroupResp, error)
	// GroupDelete deletes a group and returns a client.TeamsDeleteGroupResp.
	GroupDelete(ctx context.Context, groupID string) (*client.TeamsDeleteGroupResp, error)
	//  GroupReadByID reads a group by ID and returns a client.TeamsReadGroupResp.
	GroupReadByID(ctx context.Context, groupID string) (*client.TeamsReadGroupResp, error)
	//  GroupRead reads groups and returns a client.TeamsReadGroupsResp.
	GroupRead(ctx context.Context, groupIDs []string) (*client.TeamsReadGroupsResp, error)
	//  PermissionSetsList lists permission sets and returns a permissionssetsclient.ListPermissionSetsResponse.
	PermissionSetsList(ctx context.Context, params *permissionssetsmodels.ListPermissionSetsParams) (*permissionssetsclient.ListPermissionSetsResponse, error)
	//  PermissionSetsReadByID reads a permission set by ID and returns a permissionssetsclient.GetPermissionSetResponse.
	PermissionSetsReadByID(ctx context.Context, permissionSetId string) (*permissionssetsclient.GetPermissionSetResponse, error)
	//  PermissionSetsCreate creates a permission set and returns a permissionssetsclient.CreatePermissionSetResponse.
	PermissionSetsCreate(ctx context.Context, params permissionssetsmodels.CreatePermissionSetRecord) (*permissionssetsclient.CreatePermissionSetResponse, error)
	//  PermissionSetsUpdate updates a permission set and returns a permissionssetsclient.UpdatePermissionSetResponse.
	PermissionSetsUpdate(ctx context.Context, permissionSetId string, params permissionssetsmodels.UpdatePermissionSetRecord) (*permissionssetsclient.UpdatePermissionSetResponse, error)
	//  PermissionSetsDelete deletes a permission set and returns a permissionssetsclient.DeletePermissionSetResponse.
	PermissionSetsDelete(ctx context.Context, permissionSetId string) (*permissionssetsclient.DeletePermissionSetResponse, error)
	PermissionSetAssignmentCreate(ctx context.Context, permissionSetId string, params permissionssetsmodels.CreateAssignmentsRequest) (*permissionssetsclient.CreatePermissionSetAssignmentsResponse, error)
	PermissionSetAssignmentList(ctx context.Context, permissionSetId string, params *permissionssetsmodels.ListPermissionSetAssignmentsParams) (*permissionssetsclient.ListPermissionSetAssignmentsResponse, error)
	PermissionSetAssignmentDelete(ctx context.Context, permissionSetId string, assignmentId string) (*permissionssetsclient.DeletePermissionSetAssignmentResponse, error)
	//	MonitorTask polls an asynchronous task and returns the final task response.
	MonitorTask(ctx context.Context, taskID string, callbackFunc func(context.Context, *client.TasksReadTaskResp) bool) (*client.TasksReadTaskResp, error)
	//	MonitorPermissionSetAssignments polls an asynchronous assignment update and returns the final response.
	MonitorPermissionSetAssignments(ctx context.Context, permissionSetID, assignmentID string, completionStatuses []permissionssetsmodels.AssignmentRecordStatus, params *permissionssetsmodels.ListPermissionSetAssignmentsParams, callbackFunc func(context.Context, *permissionssetsclient.ListPermissionSetAssignmentsResponse) bool) (*permissionssetsclient.ListPermissionSetAssignmentsResponse, error)
}

//	AuthFn is the authentication function used to authenticate a client.
//
// It takes in a context, client client and auth.APIToken.
// It returns an auth.AuthResponse containing the AWS SDK config with credentials
// and Cognito IDP tokens.
type AuthFn func(ctx context.Context, client client.ClientWithResponsesInterface, apiToken *auth.APIToken) (*auth.AuthResponse, error)

type ClientOption func(*Client)

func WithClient(client client.ClientWithResponsesInterface) ClientOption {
	return func(c *Client) {
		c.client = client
	}
}

func WithPermissionSetsClient(client permissionssetsclient.ClientWithResponsesInterface) ClientOption {
	return func(c *Client) {
		c.permissionSetsClient = client
	}
}

func WithAuthFn(authFn AuthFn) ClientOption {
	return func(c *Client) {
		c.authFn = authFn
	}
}

// WithUserAgentVersion sets the user agent version used by the client.
func WithUserAgentVersion(userAgentVersion string) ClientOption {
	return func(c *Client) {
		c.userAgentVersion = userAgentVersion
	}
}

func WithHttpClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

//	WithInstallation sets the Stax installation used by the client.
//
// Options are "dev", "test", "au1", "eu1" or "us1".
func WithInstallation(installation string) ClientOption {
	return func(c *Client) {
		c.installation = installation
	}
}

func WithEndpointURL(endpointURL string) ClientOption {
	return func(c *Client) {
		c.endpointURL = endpointURL
	}
}

func WithPermissionSetsEndpointURL(permissionSetsEndpointURL string) ClientOption {
	return func(c *Client) {
		c.permissionSetsEndpointURL = permissionSetsEndpointURL
	}
}

func WithAuthRequestSigner(authRequestSigner client.RequestEditorFn) ClientOption {
	return func(c *Client) {
		c.authRequestSigner = authRequestSigner
	}
}

var _ ClientInterface = &Client{}

type Client struct {
	installation              string
	endpointURL               string
	permissionSetsEndpointURL string
	userAgentVersion          string
	httpClient                *http.Client
	client                    client.ClientWithResponsesInterface
	permissionSetsClient      permissionssetsclient.ClientWithResponsesInterface
	apiToken                  *auth.APIToken
	authRequestSigner         func(ctx context.Context, req *http.Request) error
	authFn                    AuthFn
}

//	NewClient creates a new STAX API client.
//
// apiToken: The API token to use for authentication.
// opts: Optional configuration for the client.
//
// Returns:
// - client: The STAX API client.
// - err: Any error that occurred.
//
// The client can be configured using the opts. The available options are:
//
// - WithInstallation: Sets the STAX installation used by the client. Options are "dev", "test", "au1", "eu1" or "us1".
// - WithEndpointURL: Sets the endpoint URL for the STAX API.
// - WithAuthRequestSigner: Sets the request signer used to sign API Gateway requests.
// - WithUserAgentVersion: Sets the user agent version used in requests.
// - WithHttpClient: Sets the HTTP client used to make requests.
func NewClient(apiToken *auth.APIToken, opts ...ClientOption) (*Client, error) {
	c := &Client{
		apiToken:         apiToken,
		authFn:           auth.AuthAPIToken,
		httpClient:       http.DefaultClient,
		userAgentVersion: "stax-golang-sdk/0.0.1",
	}
	for _, opt := range opts {
		opt(c)
	}

	if c.apiToken == nil {
		return nil, ErrMissingAPIToken
	}

	installationURLs, err := getInstallationURL(c.installation, installationURLs{
		CoreAPIEndpointURL:        c.endpointURL,
		PermissionSetsEndpointURL: c.permissionSetsEndpointURL,
	})
	if err != nil {
		return nil, err
	}

	if c.client == nil {
		coreClient, err := client.NewClientWithResponses(installationURLs.CoreAPIEndpointURL, client.WithHTTPClient(c.httpClient), client.WithRequestEditorFn(buildUserAgentRequestEditor(c.userAgentVersion)))
		if err != nil {
			return nil, err
		}

		c.client = coreClient
	}

	if c.permissionSetsClient == nil {
		permissionSetsClient, err := permissionssetsclient.NewClientWithResponses(installationURLs.PermissionSetsEndpointURL, permissionssetsclient.WithHTTPClient(c.httpClient), permissionssetsclient.WithRequestEditorFn(buildUserAgentRequestEditor(c.userAgentVersion)))
		if err != nil {
			return nil, err
		}

		c.permissionSetsClient = permissionSetsClient
	}

	return c, nil
}

func (cl *Client) Authenticate(ctx context.Context) error {
	authResponse, err := cl.authFn(ctx, cl.client, cl.apiToken)
	if err != nil {
		return err
	}

	authRequestSigner, err := apigw.RequestSigner(authResponse.AWSConfig.Region, auth.AWSCredentialsRetrieverFn(authResponse.AWSConfig)), nil
	if err != nil {
		return err
	}

	cl.authRequestSigner = authRequestSigner

	return nil
}

//	PublicReadConfig reads the public configuration from the STAX API.
//
// ctx: The context to use for this request.
//
// Returns:
// - publicConfigResp: The response from the PublicReadConfig API call.
// - err: Any error that occurred.
func (cl *Client) PublicReadConfig(ctx context.Context) (*client.PublicReadConfigResp, error) {
	publicConfigResp, err := cl.client.PublicReadConfigWithResponse(ctx)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, publicConfigResp, string(publicConfigResp.Body))
	if err != nil {
		return nil, err
	}

	return publicConfigResp, nil
}

//	AccountCreate creates an account in STAX.
//
// ctx: The context to use for this request.
// createAccount: The account details to create.
//
// Returns:
// - createResp: The response from the AccountsCreateAccount API call.
// - err: Any error that occurred.
func (cl *Client) AccountCreate(ctx context.Context, createAccount models.AccountsCreateAccount) (*client.AccountsCreateAccountResp, error) {
	err := cl.checkSession(ctx)
	if err != nil {
		return nil, err
	}

	createResp, err := cl.client.AccountsCreateAccountWithResponse(ctx, createAccount, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, createResp, string(createResp.Body))
	if err != nil {
		return nil, err
	}

	return createResp, nil
}

//	AccountReadByID reads an account by ID from STAX.
//
// ctx: The context to use for this request.
// accountID: The ID of the account to read.
//
// Returns:
// - readAccountRes: The response from the AccountsReadAccount API call.
// - err: Any error that occurred.
func (cl *Client) AccountReadByID(ctx context.Context, accountID string) (*client.AccountsReadAccountResp, error) {
	err := cl.checkSession(ctx)
	if err != nil {
		return nil, err
	}

	readAccountRes, err := cl.client.AccountsReadAccountWithResponse(ctx, accountID, &models.AccountsReadAccountParams{}, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, readAccountRes, string(readAccountRes.Body))
	if err != nil {
		return nil, err
	}

	// FIXME: API can return a list with zero entries if the identifier doesn't exist
	if len(readAccountRes.JSON200.Accounts) != 1 {
		return nil, fmt.Errorf("account not found for identifier: %s", accountID)
	}

	return readAccountRes, nil
}

//	AccountRead reads accounts from STAX.
//
// ctx: The context to use for this request.
// accountIDs: Optional list of account IDs to filter by.
// accountNames: Optional list of account names to filter by.
//
// Returns:
// - readAccountsRes: The response from the AccountsReadAccounts API call.
// - err: Any error that occurred.
func (cl *Client) AccountRead(ctx context.Context, accountIDs []string, accountNames []string) (*client.AccountsReadAccountsResp, error) {
	err := cl.checkSession(ctx)
	if err != nil {
		return nil, err
	}

	idFilter := helpers.CommaDelimitedOptionalValue(accountIDs)
	accountNamesFilter := helpers.CommaDelimitedOptionalValue(accountNames)

	// TODO: implement paginated results
	readAccountsRes, err := cl.client.AccountsReadAccountsWithResponse(ctx, &models.AccountsReadAccountsParams{
		Filter:       aws.String(string(models.AccountStatusACTIVE)),
		IncludeTags:  aws.Bool(true),
		IdFilter:     idFilter,
		AccountNames: accountNamesFilter,
	}, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, readAccountsRes, string(readAccountsRes.Body))
	if err != nil {
		return nil, err
	}

	return readAccountsRes, nil
}

//	AccountUpdate updates an account in STAX.
//
// ctx: The context to use for this request.
// accountID: The ID of the account to update.
// updateAccount: The account update parameters.
//
// Returns:
// - accountUpdateRes: The response from the AccountsUpdateAccount API call.
// - err: Any error that occurred.
func (cl *Client) AccountUpdate(ctx context.Context, accountID string, updateAccount models.AccountsUpdateAccount) (*client.AccountsUpdateAccountResp, error) {
	err := cl.checkSession(ctx)
	if err != nil {
		return nil, err
	}

	accountUpdateRes, err := cl.client.AccountsUpdateAccountWithResponse(ctx, accountID, updateAccount, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, accountUpdateRes, string(accountUpdateRes.Body))
	if err != nil {
		return nil, err
	}

	return accountUpdateRes, nil
}

//	AccountClose closes an account in STAX.
//
// ctx: The context to use for this request.
// accountID: The ID of the account to close.
//
// Returns:
// - accountCloseResp: The response from the AccountsCloseAccount API call.
// - err: Any error that occurred.
func (cl *Client) AccountClose(ctx context.Context, accountID string) (*client.AccountsCloseAccountResp, error) {
	err := cl.checkSession(ctx)
	if err != nil {
		return nil, err
	}

	accountCloseResp, err := cl.client.AccountsCloseAccountWithResponse(ctx, models.AccountsCloseAccount{
		Id: accountID,
	}, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, accountCloseResp, string(accountCloseResp.Body))
	if err != nil {
		return nil, err
	}

	return accountCloseResp, nil
}

//	AccountTypeReadById reads an account type by ID.
//
// ctx: The context for the request
// accountTypeID: The ID of the account type to read
//
// It returns:
//
// - *client.AccountsReadAccountTypeResp: The response containing the requested account type
// - error: Any error that occurred while making the request.
func (cl *Client) AccountTypeReadById(ctx context.Context, accountTypeID string) (*client.AccountsReadAccountTypeResp, error) {
	err := cl.checkSession(ctx)
	if err != nil {
		return nil, err
	}

	accountTypeResp, err := cl.client.AccountsReadAccountTypeWithResponse(ctx, accountTypeID, &models.AccountsReadAccountTypeParams{}, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, accountTypeResp, string(accountTypeResp.Body))
	if err != nil {
		return nil, err
	}

	// FIXME: API can return a list with zero entries if the identifier doesn't exist
	if len(accountTypeResp.JSON200.AccountTypes) != 1 {
		return nil, fmt.Errorf("account type not found for identifier: %s", accountTypeID)
	}

	return accountTypeResp, nil
}

//	AccountTypeRead reads account types from STAX.
//
// ctx: The context to use for this request.
// accountTypeIDs: A list of account type IDs to filter the results. If empty, all account types will be returned.
//
// Returns:
// - accountTypesResp: The response from the AccountsReadAccountTypes API call.
// - err: Any error that occurred.
func (cl *Client) AccountTypeRead(ctx context.Context, accountTypeIDs []string) (*client.AccountsReadAccountTypesResp, error) {
	err := cl.checkSession(ctx)
	if err != nil {
		return nil, err
	}

	accountTypesFilter := helpers.CommaDelimitedOptionalValue(accountTypeIDs)

	// TODO: implement paginated results
	accountTypesResp, err := cl.client.AccountsReadAccountTypesWithResponse(ctx, &models.AccountsReadAccountTypesParams{
		IdFilter: accountTypesFilter,
	}, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, accountTypesResp, string(accountTypesResp.Body))
	if err != nil {
		return nil, err
	}

	return accountTypesResp, nil
}

//	AccountTypeCreate creates an account type in STAX.
//
// ctx: The context to use for this request.
// name: The name of the account type to create.
//
// Returns:
// - accountTypeResp: The response from the AccountsCreateAccountType API call.
// - err: Any error that occurred.
func (cl *Client) AccountTypeCreate(ctx context.Context, name string) (*client.AccountsCreateAccountTypeResp, error) {
	accountTypeResp, err := cl.client.AccountsCreateAccountTypeWithResponse(ctx, models.AccountsCreateAccountType{
		Name: name,
	}, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, accountTypeResp, string(accountTypeResp.Body))
	if err != nil {
		return nil, err
	}

	return accountTypeResp, nil
}

//	AccountTypeUpdate updates an account type in STAX.
//
// ctx: The context to use for this request.
// accountTypeID: The ID of the account type to update.
// name: The new name of the account type.
//
// Returns:
// - accountTypeUpdateResp: The response from the AccountsUpdateAccountType API call.
// - err: Any error that occurred.
func (cl *Client) AccountTypeUpdate(ctx context.Context, accountTypeID, name string) (*client.AccountsUpdateAccountTypeResp, error) {
	accountTypeUpdateResp, err := cl.client.AccountsUpdateAccountTypeWithResponse(ctx, accountTypeID, models.AccountsUpdateAccountType{
		Name: name,
	}, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, accountTypeUpdateResp, string(accountTypeUpdateResp.Body))
	if err != nil {
		return nil, err
	}

	return accountTypeUpdateResp, nil
}

//	AccountTypeDelete deletes an account type in STAX.
//
// ctx: The context to use for this request.
// accountTypeID: The ID of the account type to delete.
//
// Returns:
// - accountTypeDeleteResp: The response from the AccountsDeleteAccountType API call.
// - err: Any error that occurred.
func (cl *Client) AccountTypeDelete(ctx context.Context, accountTypeID string) (*client.AccountsDeleteAccountTypeResp, error) {
	accountTypeDeleteResp, err := cl.client.AccountsDeleteAccountTypeWithResponse(ctx, accountTypeID, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, accountTypeDeleteResp, string(accountTypeDeleteResp.Body))
	if err != nil {
		return nil, err
	}

	return accountTypeDeleteResp, nil
}

//	WorkloadCreate creates a new workload in STAX.
//
// ctx: The context to use for this request.
// createWorkload: The details of the workload to create.
//
// Returns:
// - workloadCreateResp: The response from the WorkloadsCreateWorkload API call.
// - err: Any error that occurred.
func (cl *Client) WorkloadCreate(ctx context.Context, createWorkload models.WorkloadsCreateWorkload) (*client.WorkloadsCreateWorkloadResp, error) {
	workloadCreateResp, err := cl.client.WorkloadsCreateWorkloadWithResponse(ctx, createWorkload, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, workloadCreateResp, string(workloadCreateResp.Body))
	if err != nil {
		return nil, err
	}

	return workloadCreateResp, nil
}

//	WorkloadRead reads workloads from STAX.
//
// ctx: The context to use for this request.
// params: The parameters for filtering which workloads to read.
//
// Returns:
// - workloadsReadResp: The response from the WorkloadsReadWorkloads API call.
// - err: Any error that occurred.
func (cl *Client) WorkloadRead(ctx context.Context, params *models.WorkloadsReadWorkloadsParams) (*client.WorkloadsReadWorkloadsResp, error) {
	workloadsReadResp, err := cl.client.WorkloadsReadWorkloadsWithResponse(ctx, params, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, workloadsReadResp, string(workloadsReadResp.Body))
	if err != nil {
		return nil, err
	}

	return workloadsReadResp, nil
}

//	WorkloadDelete deletes a workload in STAX.
//
// ctx: The context to use for this request.
// workloadID: The ID of the workload to delete.
//
// Returns:
// - workloadDeleteResp: The response from the WorkloadsDeleteWorkload API call.
// - err: Any error that occurred.
func (cl *Client) WorkloadDelete(ctx context.Context, workloadID string) (*client.WorkloadsDeleteWorkloadResp, error) {
	workloadDeleteResp, err := cl.client.WorkloadsDeleteWorkloadWithResponse(ctx, workloadID, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, workloadDeleteResp, string(workloadDeleteResp.Body))
	if err != nil {
		return nil, err
	}

	return workloadDeleteResp, nil
}

//	GroupReadByID reads a group by ID from STAX.
//
// ctx is the context to use for this request.
// groupID is the ID of the group to read.
//
// Returns:
// - teamsReadResp: The response from the TeamsReadGroup API call.
// - err: Any error that occurred.
func (cl *Client) GroupReadByID(ctx context.Context, groupID string) (*client.TeamsReadGroupResp, error) {
	groupReadResp, err := cl.client.TeamsReadGroupWithResponse(ctx, groupID, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	if groupReadResp.StatusCode() == 404 {
		return nil, fmt.Errorf("group not found for identifier: %s", groupID)
	}

	err = checkResponse(ctx, groupReadResp, string(groupReadResp.Body))
	if err != nil {
		return nil, err
	}

	return groupReadResp, nil
}

//	GroupRead reads groups from STAX.
//
// ctx: The context to use for this request.
// groupIDs: A list of group IDs to filter the response by. Optional.
//
// Returns:
// - teamsReadResp: The response from the TeamsReadGroups API call.
// - err: Any error that occurred.
func (cl *Client) GroupRead(ctx context.Context, groupIDs []string) (*client.TeamsReadGroupsResp, error) {

	groupsFilter := helpers.CommaDelimitedOptionalValue(groupIDs)

	teamsReadResp, err := cl.client.TeamsReadGroupsWithResponse(ctx, &models.TeamsReadGroupsParams{
		IdFilter: groupsFilter,
	}, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, teamsReadResp, string(teamsReadResp.Body))
	if err != nil {
		return nil, err
	}

	return teamsReadResp, nil
}

func (cl *Client) GroupCreate(ctx context.Context, name string) (*client.TeamsCreateGroupResp, error) {
	createGroupResp, err := cl.client.TeamsCreateGroupWithResponse(ctx, models.TeamsCreateGroup{
		Name: name,
	}, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, createGroupResp, string(createGroupResp.Body))
	if err != nil {
		return nil, err
	}

	return createGroupResp, nil
}

func (cl *Client) GroupUpdate(ctx context.Context, groupID, name string) (*client.TeamsUpdateGroupResp, error) {
	updateGroupResp, err := cl.client.TeamsUpdateGroupWithResponse(ctx, groupID, models.TeamsUpdateGroup{
		Name: name,
	}, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, updateGroupResp, string(updateGroupResp.Body))
	if err != nil {
		return nil, err
	}

	return updateGroupResp, nil
}

func (cl *Client) GroupDelete(ctx context.Context, groupID string) (*client.TeamsDeleteGroupResp, error) {
	deleteGroupResp, err := cl.client.TeamsDeleteGroupWithResponse(ctx, groupID, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, deleteGroupResp, string(deleteGroupResp.Body))
	if err != nil {
		return nil, err
	}

	return deleteGroupResp, nil
}

func (cl *Client) UserCreate(ctx context.Context, params models.TeamsCreateUser) (*client.TeamsCreateUserResp, error) {
	createUserResp, err := cl.client.TeamsCreateUserWithResponse(ctx, params, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, createUserResp, string(createUserResp.Body))
	if err != nil {
		return nil, err
	}

	return createUserResp, nil
}

func (cl *Client) PermissionSetsReadByID(ctx context.Context, permissionSetId string) (*permissionssetsclient.GetPermissionSetResponse, error) {

	psetId, err := uuid.Parse(permissionSetId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse permission set id: %w", err)
	}

	readResp, err := cl.permissionSetsClient.GetPermissionSetWithResponse(ctx, psetId, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, readResp, string(readResp.Body))
	if err != nil {
		return nil, err
	}

	return readResp, nil
}

func (cl *Client) PermissionSetsList(ctx context.Context, params *permissionssetsmodels.ListPermissionSetsParams) (*permissionssetsclient.ListPermissionSetsResponse, error) {
	listResp, err := cl.permissionSetsClient.ListPermissionSetsWithResponse(ctx, params, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, listResp, string(listResp.Body))
	if err != nil {
		return nil, err
	}

	return listResp, nil
}

func (cl *Client) PermissionSetsCreate(ctx context.Context, params permissionssetsmodels.CreatePermissionSetRecord) (*permissionssetsclient.CreatePermissionSetResponse, error) {
	createResp, err := cl.permissionSetsClient.CreatePermissionSetWithResponse(ctx, params, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	if createResp.StatusCode() != http.StatusCreated {
		// TODO: split out each of the error types by status code
		return nil, fmt.Errorf("request failed, returned non 201 status: %s", createResp.Status())
	}

	return createResp, nil
}

func (cl *Client) PermissionSetsUpdate(ctx context.Context, permissionSetId string, params permissionssetsmodels.UpdatePermissionSetRecord) (*permissionssetsclient.UpdatePermissionSetResponse, error) {
	psetId, err := uuid.Parse(permissionSetId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse permission set id: %w", err)
	}

	updateResp, err := cl.permissionSetsClient.UpdatePermissionSetWithResponse(ctx, psetId, params, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	if updateResp.StatusCode() != http.StatusOK {
		// TODO: split out each of the error types by status code
		return nil, fmt.Errorf("request failed, returned non 200 status: %s", updateResp.Status())
	}

	return updateResp, nil
}

func (cl *Client) PermissionSetsDelete(ctx context.Context, permissionSetId string) (*permissionssetsclient.DeletePermissionSetResponse, error) {
	psetId, err := uuid.Parse(permissionSetId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse permission set id: %w", err)
	}

	deleteResp, err := cl.permissionSetsClient.DeletePermissionSetWithResponse(ctx, psetId, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	if deleteResp.StatusCode() != http.StatusOK {
		// TODO: split out each of the error types by status code
		return nil, fmt.Errorf("request failed, returned non 200 status: %s", deleteResp.Status())
	}

	return deleteResp, nil
}

func (cl *Client) PermissionSetAssignmentList(ctx context.Context, permissionSetId string, params *permissionssetsmodels.ListPermissionSetAssignmentsParams) (*permissionssetsclient.ListPermissionSetAssignmentsResponse, error) {

	psetId, err := uuid.Parse(permissionSetId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse permission set id: %w", err)
	}

	readResp, err := cl.permissionSetsClient.ListPermissionSetAssignmentsWithResponse(ctx, psetId, params, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	err = checkResponse(ctx, readResp, string(readResp.Body))
	if err != nil {
		return nil, err
	}

	return readResp, nil
}

func (cl *Client) PermissionSetAssignmentCreate(ctx context.Context, permissionSetId string, params permissionssetsmodels.CreateAssignmentsRequest) (*permissionssetsclient.CreatePermissionSetAssignmentsResponse, error) {

	psetId, err := uuid.Parse(permissionSetId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse permission set id: %w", err)
	}

	readResp, err := cl.permissionSetsClient.CreatePermissionSetAssignmentsWithResponse(ctx, psetId, params, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	if readResp.StatusCode() != http.StatusOK {
		// TODO: split out each of the error types by status code
		return nil, fmt.Errorf("request failed, returned non 200 status: %s", readResp.Status())
	}

	return readResp, nil
}

func (cl *Client) PermissionSetAssignmentDelete(ctx context.Context, permissionSetId string, assignmentId string) (*permissionssetsclient.DeletePermissionSetAssignmentResponse, error) {
	psetId, err := uuid.Parse(permissionSetId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse permission set id: %w", err)
	}

	assignId, err := uuid.Parse(assignmentId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse permission set id: %w", err)
	}

	deleteResp, err := cl.permissionSetsClient.DeletePermissionSetAssignmentWithResponse(ctx, psetId, assignId, cl.authRequestSigner)
	if err != nil {
		return nil, err
	}

	if deleteResp.StatusCode() != http.StatusOK {
		// TODO: split out each of the error types by status code
		return nil, fmt.Errorf("request failed, returned non 200 status: %s", deleteResp.Status())
	}

	return deleteResp, nil
}

//	MonitorTask polls an asynchronous task and returns the final task response.
//
// It uses a TaskPoller to poll the TasksReadTask API endpoint for the status of the task.
// It will continue polling until the task completes (succeeds or fails) or a timeout occurs.
// If the task fails or times out, an error is returned.
// Otherwise, the final client.TasksReadTaskResp is returned.
// taskID is the ID of the asynchronous task to monitor.
// callbackFunc is a function that will be called after each poll to determine whether polling should continue.
// It is passed the latest client.TasksReadTaskResp and should return true to continue polling or false to stop.
func (cl *Client) MonitorTask(ctx context.Context, taskID string, callbackFunc func(context.Context, *client.TasksReadTaskResp) bool) (*client.TasksReadTaskResp, error) {
	if taskID == "" {
		return nil, ErrMissingTaskID
	}

	// callback function used to report interim status events
	if callbackFunc == nil {
		return nil, ErrMissingTaskCallbackFunc
	}

	tp := helpers.NewTaskPoller(func() (*client.TasksReadTaskResp, error) {
		return cl.client.TasksReadTaskWithResponse(ctx, taskID, cl.authRequestSigner)
	})

	// loop for until deadline or
	for tp.Poll(ctx) {

		// the task poller checks the request success/failure so this result is always 200 OK
		taskRes := tp.Resp()

		// check whether it is OK to continue polling
		if ok := callbackFunc(ctx, taskRes); !ok {
			break
		}

		if isTaskComplete(taskRes.JSON200.Status) {
			break
		}

		time.Sleep(10 * time.Second) // TODO: check some timeout
	}

	if err := tp.Err(); err != nil {
		return nil, fmt.Errorf("task failed: %w", err)
	}

	return tp.Resp(), nil
}

func (cl *Client) MonitorPermissionSetAssignments(ctx context.Context, permissionSetID, assignmentID string, completionStatuses []permissionssetsmodels.AssignmentRecordStatus, params *permissionssetsmodels.ListPermissionSetAssignmentsParams, callbackFunc func(context.Context, *permissionssetsclient.ListPermissionSetAssignmentsResponse) bool) (*permissionssetsclient.ListPermissionSetAssignmentsResponse, error) {
	if permissionSetID == "" || assignmentID == "" {
		return nil, errors.New("missing permissionSetID or assignmentID")
	}

	// callback function used to report interim status events
	if callbackFunc == nil {
		return nil, errors.New("missing assignment monitoring callback function")
	}

	psetId, err := uuid.Parse(permissionSetID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse permission set id: %w", err)
	}

	tp := helpers.NewTaskPoller(func() (*permissionssetsclient.ListPermissionSetAssignmentsResponse, error) {
		return cl.permissionSetsClient.ListPermissionSetAssignmentsWithResponse(ctx, psetId, params, cl.authRequestSigner)
	})

	// loop for until deadline or
	for tp.Poll(ctx) {

		// the task poller checks the request success/failure so this result is always 200 OK
		taskRes := tp.Resp()

		// check whether it is OK to continue polling
		if ok := callbackFunc(ctx, taskRes); !ok {
			break
		}

		if isAssignmentComplete(assignmentID, completionStatuses, taskRes.JSON200.Assignments) {
			break
		}

		time.Sleep(10 * time.Second) // TODO: check some timeout
	}

	if err := tp.Err(); err != nil {
		return nil, fmt.Errorf("task failed: %w", err)
	}

	return tp.Resp(), nil

}

func isAssignmentComplete(assignmentID string, completionStatuses []permissionssetsmodels.AssignmentRecordStatus, assignments []permissionssetsmodels.AssignmentRecord) bool {
	for _, assignment := range assignments {
		if assignment.Id.String() == assignmentID {
			return slices.Contains(completionStatuses, assignment.Status)
		}
	}

	return false
}

func isTaskComplete(status models.OperationStatus) bool {
	return status == TaskFailed || status == TaskSucceeded
}

func (cl *Client) checkSession(_ context.Context) error {
	if cl.authRequestSigner == nil {
		return ErrAuthSessionEmpty
	}

	return nil
}

func checkResponse(_ context.Context, res helpers.HTTPResponse, _ string) error {
	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("request failed, returned non 200 status: %s", res.Status())
	}

	return nil
}

type installationURLs struct {
	CoreAPIEndpointURL        string
	PermissionSetsEndpointURL string
}

func getInstallationURL(installation string, overrideEndpointURLs installationURLs) (*installationURLs, error) {
	if !reflect.ValueOf(overrideEndpointURLs).IsZero() {
		return &overrideEndpointURLs, nil
	}

	switch installation {
	case "au1":
		return &installationURLs{
			CoreAPIEndpointURL:        "https://api.au1.staxapp.cloud",
			PermissionSetsEndpointURL: "https://api.idam.au1.staxapp.cloud/20210321",
		}, nil
	case "us1":
		return &installationURLs{
			CoreAPIEndpointURL:        "https://api.us1.staxapp.cloud",
			PermissionSetsEndpointURL: "https://api.idam.us1.staxapp.cloud/20210321",
		}, nil
	case "eu1":
		return &installationURLs{
			CoreAPIEndpointURL:        "https://api.eu1.staxapp.cloud",
			PermissionSetsEndpointURL: "https://api.idam.eu1.staxapp.cloud/20210321",
		}, nil
	}

	return nil, ErrInvalidInstallation
}

func buildUserAgentRequestEditor(tag string) func(ctx context.Context, req *http.Request) error {
	return func(ctx context.Context, req *http.Request) error {
		req.Header.Set("User-Agent", buildUserAgentVersion(tag))
		return nil
	}
}

// buildUserAgentVersion build user agent
//
// stax-golang-sdk/0.0.1 md/GOOS/linux md/GOARCH/amd64 lang/go/1.15.
func buildUserAgentVersion(userAgentVersion string) string {

	tokens := []string{userAgentVersion}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		tokens = append(tokens,
			fmt.Sprintf("md/GOOS/%s", runtime.GOOS),
			fmt.Sprintf("md/GOARCH/%s", runtime.GOARCH),
			fmt.Sprintf("lang/go/%s", buildInfo.GoVersion),
		)
	}

	return strings.Join(tokens, " ")
}
