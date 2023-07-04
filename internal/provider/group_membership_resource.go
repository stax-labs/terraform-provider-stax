package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stax-labs/terraform-provider-stax/internal/api/helpers"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/client"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
	"golang.org/x/exp/slices"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &GroupMembershipResource{}
var _ resource.ResourceWithConfigure = &GroupMembershipResource{}
var _ resource.ResourceWithImportState = &GroupMembershipResource{}

type GroupMembershipResourceModel struct {
	ID       types.String `tfsdk:"id"`
	UsersIDs types.Set    `tfsdk:"user_ids"`
}

func NewGroupMembershipResource() resource.Resource {
	return &GroupMembershipResource{}
}

// GroupMembershipResource defines the resource implementation.
type GroupMembershipResource struct {
	client staxsdk.ClientInterface
}

func (r *GroupMembershipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_membership"
}

func (r *GroupMembershipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Stax Group Membership resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Group identifier",
			},
			"user_ids": schema.SetAttribute{
				MarkdownDescription: "Array of IDs of Stax Users belonging to the Group",
				ElementType:         types.StringType,
				Optional:            true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (r *GroupMembershipResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *GroupMembershipResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var addMemberUserIDs []string
	resp.Diagnostics.Append(data.UsersIDs.ElementsAs(ctx, &addMemberUserIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "assigning users to group", map[string]interface{}{
		"group_id": data.ID.ValueString(),
		"user_ids": addMemberUserIDs,
	})

	assignResp, err := r.client.GroupAssignUsers(ctx, data.ID.ValueString(), addMemberUserIDs, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to assign users to group, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "group assignment create response", map[string]interface{}{
		"JSON200": assignResp.JSON200,
	})

	taskResp, err := waitForTask(ctx, *assignResp.JSON200.Detail.TaskId, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to complete assign users task, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "assign users task response", map[string]interface{}{
		"JSON200": taskResp,
	})

	resp.Diagnostics.Append(r.readGroup(ctx, data.ID.ValueString(), data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *GroupMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *GroupMembershipResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	err := r.readGroup(ctx, data.ID.ValueString(), data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read group, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planData *GroupMembershipResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "update group membership", map[string]interface{}{
		"id": planData.ID.ValueString(),
	})

	assignResp, d := r.updateGroupMembership(ctx, planData.ID.ValueString(), planData)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "group assignment create response", map[string]interface{}{
		"JSON200": assignResp.JSON200,
	})

	taskResp, err := waitForTask(ctx, *assignResp.JSON200.Detail.TaskId, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to complete assign users task, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "assign users task response", map[string]interface{}{
		"JSON200": taskResp,
	})

	resp.Diagnostics.Append(r.readGroup(ctx, planData.ID.ValueString(), planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save planData into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)
}

func (r *GroupMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *GroupMembershipResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var deleteMemberUserIDs []string
	resp.Diagnostics.Append(data.UsersIDs.ElementsAs(ctx, &deleteMemberUserIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "removing users from group", map[string]interface{}{
		"group_id": data.ID.ValueString(),
		"user_ids": deleteMemberUserIDs,
	})

	assignResp, err := r.client.GroupAssignUsers(ctx, data.ID.ValueString(), nil, deleteMemberUserIDs)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to assign users to group, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "group assignment delete response", map[string]interface{}{
		"JSON200": assignResp.JSON200,
	})

	taskResp, err := waitForTask(ctx, *assignResp.JSON200.Detail.TaskId, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to complete remove users task, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "remove users task response", map[string]interface{}{
		"JSON200": taskResp,
	})

}

func (r *GroupMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *GroupMembershipResource) readGroup(ctx context.Context, groupID string, data *GroupMembershipResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	groupsResp, err := r.client.GroupReadByID(ctx, groupID)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read group, got error: %s", err))
		return diags
	}

	tflog.Info(ctx, "reading groups", map[string]interface{}{
		"count": len(groupsResp.JSON200.Groups),
	})

	for _, group := range groupsResp.JSON200.Groups {

		if group.Users != nil {
			slices.Sort(*group.Users)
		}

		usersList, d := types.SetValueFrom(ctx, types.StringType, group.Users)
		diags.Append(d...)

		data.ID = types.StringValue(*group.Id)
		data.UsersIDs = usersList
	}

	return diags
}

func (r *GroupMembershipResource) updateGroupMembership(ctx context.Context, groupID string, data *GroupMembershipResourceModel) (*client.TeamsUpdateGroupMembersResp, diag.Diagnostics) {
	var diags diag.Diagnostics

	// get the current state from the API
	groupsResp, err := r.client.GroupReadByID(ctx, groupID)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read group, got error: %s", err))
		return nil, diags
	}

	currentGroupUsers := groupsToUserIDMap(groupsResp.JSON200.Groups)

	userIDs := make([]string, 0)
	diags.Append(data.UsersIDs.ElementsAs(ctx, &userIDs, false)...)

	planGroupUsers := userIDMap(userIDs)

	addedUserIDs := helpers.Subtract(planGroupUsers, currentGroupUsers)
	removedUserIDs := helpers.Subtract(currentGroupUsers, planGroupUsers)

	tflog.Info(ctx, "group membership update", map[string]interface{}{
		"added":   addedUserIDs,
		"removed": removedUserIDs,
	})

	resp, err := r.client.GroupAssignUsers(ctx, groupID, addedUserIDs, removedUserIDs)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to assign users to group, got error: %s", err))
		return nil, diags
	}

	return resp, diags
}

func groupsToUserIDMap(groups []models.Group) map[string]bool {
	m := make(map[string]bool)

	for _, group := range groups {

		for _, user := range *group.Users {
			m[user] = true
		}
	}

	return m
}

func userIDMap(userIDs []string) map[string]bool {
	m := make(map[string]bool)

	for _, user := range userIDs {
		m[user] = true
	}

	return m
}
