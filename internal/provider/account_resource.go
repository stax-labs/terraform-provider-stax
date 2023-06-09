package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/client"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AccountResource{}
var _ resource.ResourceWithConfigure = &AccountResource{}
var _ resource.ResourceWithImportState = &AccountResource{}

type AccountResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Status          types.String `tfsdk:"status"`
	AWsAccountID    types.String `tfsdk:"aws_account_id"`
	AccountTypeID   types.String `tfsdk:"account_type_id"`
	AccountType     types.String `tfsdk:"account_type"`
	AwsAccountAlias types.String `tfsdk:"aws_account_alias"`
	Tags            types.Map    `tfsdk:"tags"`
}

func NewAccountResource() resource.Resource {
	return &AccountResource{}
}

// AccountResource defines the resource implementation.
type AccountResource struct {
	client staxsdk.ClientInterface
}

func (r *AccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (r *AccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Account resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Account identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the account",
				Required:            true,
			},
			"account_type_id": schema.StringAttribute{
				MarkdownDescription: "The account type identifier for the account",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_type": schema.StringAttribute{
				MarkdownDescription: "The account type for the account",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Account Status",
				Computed:            true,
			},
			"aws_account_id": schema.StringAttribute{
				MarkdownDescription: "The aws account identifier for the stax account",
				Computed:            true,
			},
			"aws_account_alias": schema.StringAttribute{
				MarkdownDescription: "The aws account alias for the stax account",
				Optional:            true,
			},
			"tags": schema.MapAttribute{
				MarkdownDescription: "Account tags",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *AccountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*staxsdk.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *AccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *AccountResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ac := models.AccountsCreateAccount_AccountType{}
	err := ac.FromRoUuidv4(data.AccountTypeID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create account type from UUID, got error: %s", err))
		return
	}

	staxTags := make(map[string]string)
	resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &staxTags, false)...)

	created, err := r.client.AccountCreate(ctx, models.AccountsCreateAccount{
		Name:        data.Name.ValueString(),
		AccountType: ac,
		Tags:        (*models.StaxTags)(&staxTags),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create account, got error: %s", err))
		return
	}

	taskResp, err := waitForTask(ctx, *created.JSON200.TaskId, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to complete task, got error: %s", err))
		return
	}

	accountsResp, err := r.client.AccountRead(ctx, *taskResp.Accounts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read account, got error: %s", err))
		return
	}

	tflog.Info(ctx, "reading accounts", map[string]interface{}{
		"count": len(accountsResp.JSON200.Accounts),
	})

	data.ID = types.StringValue(toString(accountsResp.JSON200.Accounts[0].Id))

	err = r.readAccount(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read account, got error: %s", err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *AccountResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readAccount(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read account, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *AccountResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	staxTags := make(map[string]string)
	resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &staxTags, false)...)

	tflog.Info(ctx, "update account", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	accountResp, err := r.client.AccountUpdate(ctx, data.ID.ValueString(), models.AccountsUpdateAccount{
		Name:        stringPtr(data.Name.ValueString()),
		AccountType: stringPtr(data.AccountTypeID.ValueString()),
		Tags:        (*models.StaxTags)(&staxTags),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read account, got error: %s", err))
		return
	}

	tflog.Info(ctx, "wait for account update", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	_, err = waitForTask(ctx, *accountResp.JSON200.TaskId, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to complete task, got error: %s", err))
		return
	}

	tflog.Info(ctx, "read updated account", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	err = r.readAccount(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read account, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *AccountResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddWarning(
		"Account Close Considerations",
		"Applying this resource destruction will only remove the resource from the Terraform state "+
			"and will not call the account close API due to AWS API limitations. Manually use the web "+
			"interface to fully close this account.",
	)

}

func (r *AccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *AccountResource) readAccount(ctx context.Context, data *AccountResourceModel) error {
	accountsResp, err := r.client.AccountRead(ctx, []string{data.ID.ValueString()}, nil)
	if err != nil {
		return err
	}

	tflog.Info(ctx, "reading accounts", map[string]interface{}{
		"count": len(accountsResp.JSON200.Accounts),
	})

	accountTypesResp, err := r.client.AccountTypeRead(ctx, []string{})
	if err != nil {
		return err
	}

	accountTypesMap := make(map[string]string)

	for _, accountType := range accountTypesResp.JSON200.AccountTypes {
		accountTypesMap[accountType.Name] = toString(accountType.Id)
	}

	for _, account := range accountsResp.JSON200.Accounts {
		data.ID = types.StringValue(*account.Id)
		data.Name = types.StringValue(account.Name)
		data.Status = types.StringValue(string(*account.Status))
		data.AWsAccountID = types.StringValue(toString(account.AwsAccountId))

		if account.AwsAccountAlias != nil {
			data.AwsAccountAlias = types.StringValue(toString(account.AwsAccountAlias))
		}

		tags := staxTagsToMapString(account.Tags)
		if len(tags) > 0 {
			data.Tags = types.MapValueMust(types.StringType, tags)
		}

		if account.AccountType != nil {
			if accountTypeID, ok := accountTypesMap[toString(account.AccountType)]; ok {
				data.AccountTypeID = types.StringValue(accountTypeID)
				data.AccountType = types.StringValue(toString(account.AccountType))
			}
		}
	}

	return nil
}

func waitForTask(ctx context.Context, taskID string, staxclient staxsdk.ClientInterface) (*models.TasksReadTask, error) {
	tflog.Debug(ctx, "waiting for task", map[string]interface{}{
		"taskID": len(taskID),
	})

	finalTaskStatus, err := staxclient.MonitorTask(ctx, taskID, func(ctx context.Context, taskRes *client.TasksReadTaskResp) bool {
		tflog.Debug(ctx, "read status of task", map[string]interface{}{
			"taskID": len(taskID),
		})

		return true
	})
	if err != nil {
		return nil, err
	}

	if finalTaskStatus.JSON200.Status != staxsdk.TaskSucceeded {
		tflog.Error(ctx, "task failed", map[string]interface{}{
			"task": finalTaskStatus.JSON200,
		})

		return nil, fmt.Errorf("something went wrong with task, final status: %s", finalTaskStatus.JSON200.Status)
	}

	return finalTaskStatus.JSON200, nil
}

func stringPtr(s string) *string {
	return &s
}

func toString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func staxTagsToMap(tags *models.StaxTags) map[string]string {
	accountTags := make(map[string]string)

	if tags == nil {
		return accountTags
	}

	for k, v := range *tags {
		accountTags[k] = v
	}

	return accountTags
}

func staxTagsToMapString(tags *models.StaxTags) map[string]attr.Value {
	accountTags := make(map[string]attr.Value)

	if tags == nil {
		return accountTags
	}

	for k, v := range *tags {
		accountTags[k] = types.StringValue(v)
	}

	return accountTags
}
