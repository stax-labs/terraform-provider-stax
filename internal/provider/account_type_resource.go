package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AccountTypeResource{}
var _ resource.ResourceWithConfigure = &AccountTypeResource{}
var _ resource.ResourceWithImportState = &AccountTypeResource{}

type AccountTypeResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Status types.String `tfsdk:"status"`
}

func NewAccountTypeResource() resource.Resource {
	return &AccountTypeResource{}
}

// AccountResource defines the resource implementation.
type AccountTypeResource struct {
	client staxsdk.ClientInterface
}

func (r *AccountTypeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_type"
}

func (r *AccountTypeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Account Type resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Account type identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the account type",
				Required:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the stax account type",
				Computed:            true,
			},
		},
	}
}

func (r *AccountTypeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AccountTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *AccountTypeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	accountTypeCreate, err := r.client.AccountTypeCreate(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create account type, got error: %s", err))
		return
	}

	accountTypesRead, err := r.client.AccountTypeRead(ctx, []string{aws.ToString(accountTypeCreate.JSON200.Detail.AccountType.Id)})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read account type, got error: %s", err))
		return
	}

	for _, accountType := range accountTypesRead.JSON200.AccountTypes {
		data.ID = types.StringValue(aws.ToString(accountType.Id))
		data.Name = types.StringValue(accountType.Name)
		data.Status = types.StringValue(fmt.Sprintf("%s", accountType.Status))
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccountTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *AccountTypeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	accountTypesRead, err := r.client.AccountTypeRead(ctx, []string{data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read account type, got error: %s", err))
		return
	}

	for _, accountType := range accountTypesRead.JSON200.AccountTypes {
		data.ID = types.StringValue(aws.ToString(accountType.Id))
		data.Name = types.StringValue(accountType.Name)
		data.Status = types.StringValue(fmt.Sprintf("%s", accountType.Status))
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccountTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *AccountTypeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	accountResp, err := r.client.AccountTypeUpdate(ctx, data.ID.ValueString(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update account type, got error: %s", err))
		return
	}

	accountTypesRead, err := r.client.AccountTypeRead(ctx, []string{aws.ToString(accountResp.JSON200.AccountTypes.Id)})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read account type, got error: %s", err))
		return
	}

	for _, accountType := range accountTypesRead.JSON200.AccountTypes {
		data.ID = types.StringValue(aws.ToString(accountType.Id))
		data.Name = types.StringValue(accountType.Name)
		data.Status = types.StringValue(fmt.Sprintf("%s", accountType.Status))
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccountTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *AccountTypeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	accountTypeDeleteResp, err := r.client.AccountTypeDelete(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete account type, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "account type deleted", map[string]interface{}{
		"id": aws.ToString(accountTypeDeleteResp.JSON200.AccountTypes.Id),
	})
}

func (r *AccountTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
