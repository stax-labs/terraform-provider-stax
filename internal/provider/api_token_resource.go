package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &APITokenResource{}
var _ resource.ResourceWithConfigure = &APITokenResource{}
var _ resource.ResourceWithImportState = &APITokenResource{}

type APITokenResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Role       types.String `tfsdk:"role"`
	Status     types.String `tfsdk:"status"`
	CreatedTS  types.String `tfsdk:"created_ts"`
	ModifiedTS types.String `tfsdk:"modified_ts"`
	Tags       types.Map    `tfsdk:"tags"`
}

func NewAPITokenResource() resource.Resource {
	return &APITokenResource{}
}

// APITokenResource defines the resource implementation.
type APITokenResource struct {
	client staxsdk.ClientInterface
}

func (r *APITokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_token"
}

func (r *APITokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a Stax API Token resource. [Stax API Token](https://support.stax.io/hc/en-us/articles/4447315161231-About-API-Tokens) are security credentials that can be used to authenticate to the Stax API. Stax will create and store them securely in a customers security AWS account using [Systems Manager Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html)",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "API Token identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The first name of the stax api token",
				Required:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role of the stax api token, this can be one of `api_admin`, `api_user`, or `api_readonly`",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("api_admin", "api_user", "api_readonly"),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the stax stax api",
				Computed:            true,
			},
			"created_ts": schema.StringAttribute{
				MarkdownDescription: "The created timestamp for the stax stax api",
				Computed:            true,
			},
			"modified_ts": schema.StringAttribute{
				MarkdownDescription: "The modified timestamp for the stax stax api",
				Computed:            true,
			},
			"tags": schema.MapAttribute{
				MarkdownDescription: "The tags associated with the stax api token",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *APITokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *APITokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *APITokenResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	m := make(map[string]string)
	resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &m, false)...)

	created, err := r.client.APITokenCreate(ctx, models.TeamsCreateApiToken{
		Name:       data.Name.ValueString(),
		Role:       (models.ApiRole)(data.Role.ValueString()),
		StoreToken: aws.Bool(true), // Store the token in AWS Parameter store by default
		Tags:       (*models.Tags)(&m),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create api token, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "api token create response", map[string]interface{}{
		"JSON200": created.JSON200,
	})

	resp.Diagnostics.Append(r.apiTokenRead(ctx, aws.ToString(created.JSON200.ApiTokens[0].AccessKey), data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APITokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *APITokenResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	resp.Diagnostics.Append(r.apiTokenRead(ctx, data.ID.ValueString(), data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APITokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *APITokenResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "update user", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	m := make(map[string]string)
	resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &m, false)...)

	params := models.TeamsUpdateApiToken{
		Role: (*models.ApiRole)(data.Role.ValueStringPointer()),
		Tags: (*models.Tags)(&m),
	}

	apiTokenResp, err := r.client.ApiTokenUpdate(ctx, data.ID.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "user update response", map[string]interface{}{
		"JSON200": apiTokenResp.JSON200,
	})

	resp.Diagnostics.Append(r.apiTokenRead(ctx, data.ID.ValueString(), data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APITokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *APITokenResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	deleteResp, err := r.client.ApiTokenDelete(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "api token deleted", map[string]interface{}{
		"id":         data.ID.ValueString(),
		"api_tokens": deleteResp.JSON200.ApiTokens,
	})
}

func (r *APITokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *APITokenResource) apiTokenRead(ctx context.Context, userID string, data *APITokenResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	apiTokenRead, err := r.client.APITokenReadByID(ctx, userID)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return diags
	}

	for _, apiToken := range apiTokenRead.JSON200.ApiTokens {

		m := tokenTagsToMap(apiToken.Tags)

		tags, d := types.MapValueFrom(ctx, types.StringType, m)
		diags.Append(d...)

		data.ID = types.StringValue(aws.ToString(apiToken.AccessKey))
		data.Name = types.StringValue(apiToken.Name)
		data.Role = types.StringValue(string(apiToken.Role))
		data.Status = types.StringValue(string(apiToken.Status))
		data.CreatedTS = types.StringPointerValue(timeToStringPtr(apiToken.CreatedTS))
		data.ModifiedTS = types.StringPointerValue(timeToStringPtr(apiToken.ModifiedTS))
		data.Tags = tags
	}

	return diags
}
