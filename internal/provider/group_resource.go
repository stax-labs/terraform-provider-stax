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
var _ resource.Resource = &GroupResource{}
var _ resource.ResourceWithConfigure = &GroupResource{}
var _ resource.ResourceWithImportState = &GroupResource{}

type GroupResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

// GroupResource defines the resource implementation.
type GroupResource struct {
	client staxsdk.ClientInterface
}

func (r *GroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Stax Group resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Group identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the group",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of Stax Group, this can be either `LOCAL` or `SCIM`. Note that groups with a type of `SCIM` cannot be modified by this provider.",
				Computed:            true,
			},
		},
	}
}

func (r *GroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *GroupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.GroupCreate(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create group, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "group create response", map[string]interface{}{
		"JSON200": created.JSON200,
	})

	taskResp, err := waitForTask(ctx, *created.JSON200.Detail.TaskId, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to complete task, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "task response", map[string]interface{}{
		"JSON200": taskResp,
	})

	err = r.readGroup(ctx, aws.ToString(created.JSON200.GroupId), data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read group, got error: %s", err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *GroupResourceModel

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

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *GroupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Type.ValueString() == "SCIM" {
		resp.Diagnostics.AddAttributeError(
			path.Root("type"),
			"Stax Groups of type SCIM can't be modified",
			"The provider cannot update groups which have a type of SCIM, these are managed by the SCIM service and are READONLY. ",
		)

		return
	}

	tflog.Info(ctx, "update group", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	groupResp, err := r.client.GroupUpdate(ctx, data.ID.ValueString(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update group, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "group update response", map[string]interface{}{
		"JSON200": groupResp.JSON200,
	})

	tflog.Info(ctx, "wait for group update", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	taskResp, err := waitForTask(ctx, *groupResp.JSON200.Detail.TaskId, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to complete task, got error: %s", err))
		return
	}

	tflog.Info(ctx, "group update successful", map[string]interface{}{
		"id":       data.ID.ValueString(),
		"taskResp": taskResp,
	})

	err = r.readGroup(ctx, data.ID.ValueString(), data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read group, got error: %s", err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *GroupResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if data.Type.ValueString() == "SCIM" {
		resp.Diagnostics.AddAttributeError(
			path.Root("type"),
			"Stax Groups of type SCIM can't be modified",
			"The provider cannot update groups which have a type of SCIM, these are managed by the SCIM service and are READONLY. ",
		)

		return
	}

	deleteResp, err := r.client.GroupDelete(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete group, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "group deleted", map[string]interface{}{
		"id":      data.ID.ValueString(),
		"task_id": deleteResp.JSON200.Detail.TaskId,
	})
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *GroupResource) readGroup(ctx context.Context, groupID string, data *GroupResourceModel) error {
	groupsResp, err := r.client.GroupReadByID(ctx, groupID)
	if err != nil {
		return err
	}

	tflog.Info(ctx, "reading groups", map[string]interface{}{
		"count": len(groupsResp.JSON200.Groups),
	})

	for _, group := range groupsResp.JSON200.Groups {
		data.ID = types.StringValue(*group.Id)
		data.Name = types.StringValue(group.Name)
		data.Type = types.StringValue(string(group.GroupType))
	}

	return nil
}
