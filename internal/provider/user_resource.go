package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
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
var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithConfigure = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

type UserResourceModel struct {
	ID         types.String `tfsdk:"id"`
	FirstName  types.String `tfsdk:"first_name"`
	LastName   types.String `tfsdk:"last_name"`
	Email      types.String `tfsdk:"email"`
	Role       types.String `tfsdk:"role"`
	Status     types.String `tfsdk:"status"`
	AuthOrigin types.String `tfsdk:"auth_origin"`
	CreatedTS  types.String `tfsdk:"created_ts"`
	ModifiedTS types.String `tfsdk:"modified_ts"`
}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

// UserResource defines the resource implementation.
type UserResource struct {
	client staxsdk.ClientInterface
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Stax User resource. [Stax Users](https://support.stax.io/hc/en-us/articles/4445031773711-Manage-Users) allows you to manage users details for non federated logins.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "User identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"first_name": schema.StringAttribute{
				MarkdownDescription: "The first name of the stax user",
				Required:            true,
			},
			"last_name": schema.StringAttribute{
				MarkdownDescription: "The last name of the stax user",
				Required:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email of the stax user",
				Required:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role of the stax user, this can be one of `customer_admin`, `customer_user`, `customer_readonly` or `customer_costadmin`",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("customer_admin", "customer_user", "customer_readonly", "customer_costadmin"),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the stax user",
				Computed:            true,
			},
			"auth_origin": schema.StringAttribute{
				MarkdownDescription: "The authentication origin of the stax user",
				Computed:            true,
			},
			"created_ts": schema.StringAttribute{
				MarkdownDescription: "The created timestamp for the stax user",
				Computed:            true,
			},
			"modified_ts": schema.StringAttribute{
				MarkdownDescription: "The modified timestamp for the stax user",
				Computed:            true,
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *UserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.UserCreate(ctx, models.TeamsCreateUser{
		Email:     (openapi_types.Email)(data.Email.ValueString()),
		FirstName: data.FirstName.ValueString(),
		LastName:  data.LastName.ValueString(),
		Role:      (*models.IdamUserRole)(data.Role.ValueStringPointer()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "user create response", map[string]interface{}{
		"JSON200": created.JSON200,
	})

	taskResp, err := waitForTask(ctx, *created.JSON200.TaskId, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to complete task, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "task response", map[string]interface{}{
		"JSON200": taskResp,
	})

	userID, err := extractUserID(taskResp.Logs[0])
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to extract user id from task, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "extracted user id from logs", map[string]interface{}{
		"user_id": userID,
	})

	resp.Diagnostics.Append(r.userRead(ctx, userID, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *UserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	resp.Diagnostics.Append(r.userRead(ctx, data.ID.ValueString(), data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *UserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.AuthOrigin.ValueString() == "federated" {
		resp.Diagnostics.AddAttributeError(
			path.Root("auth_origin"),
			"Stax Users of type federated can't be modified",
			"The provider cannot update users which have a type of federated, these are managed by the federated service and are READONLY. ",
		)

		return
	}

	tflog.Info(ctx, "update user", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	params := models.TeamsUpdateUser{
		Email:     (*openapi_types.Email)(data.Email.ValueStringPointer()),
		FirstName: data.FirstName.ValueStringPointer(),
		LastName:  data.LastName.ValueStringPointer(),
		Role:      (*models.IdamUserRole)(data.Role.ValueStringPointer()),
	}

	userResp, err := r.client.UserUpdate(ctx, data.ID.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "user update response", map[string]interface{}{
		"JSON200": userResp.JSON200,
	})

	tflog.Info(ctx, "wait for user update", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	taskResp, err := waitForTask(ctx, *userResp.JSON200.TaskId, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to complete task, got error: %s", err))
		return
	}

	tflog.Info(ctx, "user update successful", map[string]interface{}{
		"id":       data.ID.ValueString(),
		"taskResp": taskResp,
	})

	resp.Diagnostics.Append(r.userRead(ctx, data.ID.ValueString(), data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *UserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if data.AuthOrigin.ValueString() == "federated" {
		resp.Diagnostics.AddAttributeError(
			path.Root("auth_origin"),
			"Stax Users of type federated can't be modified",
			"The provider cannot update users which have a type of federated, these are managed by the federated service and are READONLY. ",
		)

		return
	}

	deleteResp, err := r.client.UserDelete(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "user deleted", map[string]interface{}{
		"id":    data.ID.ValueString(),
		"users": deleteResp.JSON200.Users,
	})
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *UserResource) userRead(ctx context.Context, userID string, data *UserResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	userRead, err := r.client.UserReadByID(ctx, userID)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return diags
	}

	for _, user := range userRead.JSON200.Users {

		var email *string
		if user.Email != nil {
			email = (*string)(user.Email)
		}

		data.ID = types.StringValue(aws.ToString(user.Id))
		data.FirstName = types.StringPointerValue(user.FirstName)
		data.LastName = types.StringPointerValue(user.LastName)
		data.Role = types.StringPointerValue(userRoleToString(user.Role))
		data.Status = types.StringPointerValue(userStatusToString(user.Status))
		data.Email = types.StringPointerValue(email)
		data.AuthOrigin = types.StringPointerValue(user.AuthOrigin)
		data.CreatedTS = types.StringPointerValue(timeToStringPtr(user.CreatedTS))
		data.ModifiedTS = types.StringPointerValue(timeToStringPtr(user.ModifiedTS))
	}

	return diags
}

func extractUserID(message string) (string, error) {
	re := regexp.MustCompile(`^Successfully created user (.*)$`)
	matches := re.FindStringSubmatch(message)
	if len(matches) != 2 {
		return "", fmt.Errorf("unexpected task response: %s", message)
	}

	return matches[1], nil
}
