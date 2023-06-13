//  Package staxsdk contains a client for interacting with the Stax API.

package staxsdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stax-labs/terraform-provider-stax/internal/api/apigw"
	"github.com/stax-labs/terraform-provider-stax/internal/api/auth"
	"github.com/stax-labs/terraform-provider-stax/internal/api/helpers"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/client"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/models"
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
	// AccountTypeRead reads account types and returns a client.AccountsReadAccountTypesResp.
	AccountTypeRead(ctx context.Context, accountTypeIDs []string) (*client.AccountsReadAccountTypesResp, error)
	// WorkloadDelete deletes a workload and returns a client.WorkloadsDeleteWorkloadResp.
	WorkloadDelete(ctx context.Context, workloadID string) (*client.WorkloadsDeleteWorkloadResp, error)
	//  GroupRead reads groups and returns a client.TeamsReadGroupsResp.
	GroupRead(ctx context.Context, groupIDs []string) (*client.TeamsReadGroupsResp, error)
	//	MonitorTask polls an asynchronous task and returns the final task response.
	MonitorTask(ctx context.Context, taskID string, callbackFunc func(context.Context, *client.TasksReadTaskResp) bool) (*client.TasksReadTaskResp, error)
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

func WithAuthRequestSigner(authRequestSigner client.RequestEditorFn) ClientOption {
	return func(c *Client) {
		c.authRequestSigner = authRequestSigner
	}
}

var _ ClientInterface = &Client{}

type Client struct {
	installation      string
	endpointURL       string
	userAgentVersion  string
	httpClient        *http.Client
	client            client.ClientWithResponsesInterface
	apiToken          *auth.APIToken
	authRequestSigner client.RequestEditorFn
	authFn            AuthFn
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

	url, err := getInstallationURL(c.installation, c.endpointURL)
	if err != nil {
		return nil, err
	}

	if c.client == nil {
		c.client, err = client.NewClientWithResponses(url, client.WithHTTPClient(c.httpClient), buildUserAgentRequestEditor(c.userAgentVersion))
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (cl *Client) Authenticate(ctx context.Context) error {
	authResponse, err := cl.authFn(ctx, cl.client, cl.apiToken)
	if err != nil {
		return err
	}

	cl.authRequestSigner, err = apigw.RequestSigner(authResponse.AWSConfig.Region, auth.AWSCredentialsRetrieverFn(authResponse.AWSConfig)), nil
	if err != nil {
		return err
	}

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

	tp := helpers.NewTaskPoller(taskID, func() (*client.TasksReadTaskResp, error) {
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

func getInstallationURL(installation, endpointURL string) (string, error) {
	if endpointURL != "" {
		return endpointURL, nil
	}

	switch installation {
	case "dev":
		return "https://api.core.dev.juma.cloud", nil
	case "test":
		return "https://api.core.test.juma.cloud", nil
	case "au1":
		return "https://api.au1.staxapp.cloud", nil
	case "us1":
		return "https://api.us1.staxapp.cloud", nil
	case "eu1":
		return "https://api.eu1.staxapp.cloud", nil
	}

	return "", ErrInvalidInstallation
}

func buildUserAgentRequestEditor(tag string) client.ClientOption {
	return client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("User-Agent", buildUserAgentVersion(tag))
		return nil
	})
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
