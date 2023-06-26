package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/client"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
	"golang.org/x/exp/slices"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PermissionSetAssignmentResource{}
var _ resource.ResourceWithConfigure = &PermissionSetAssignmentResource{}
var _ resource.ResourceWithImportState = &PermissionSetAssignmentResource{}

type PermissionSetAssignmentResourceModel struct {
	ID              types.String `tfsdk:"id"`
	PermissionSetID types.String `tfsdk:"permission_set_id"`
	AccountTypeID   types.String `tfsdk:"account_type_id"`
	GroupID         types.String `tfsdk:"group_id"`
	Status          types.String `tfsdk:"status"`
	CreatedBy       types.String `tfsdk:"created_by"`
	CreatedTS       types.String `tfsdk:"created_ts"`
}

func NewPermissionSetAssignmentResource() resource.Resource {
	return &PermissionSetAssignmentResource{}
}

// PermissionSetAssignmentResource defines the resource implementation.
type PermissionSetAssignmentResource struct {
	client staxsdk.ClientInterface
}

func (r *PermissionSetAssignmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission_set_assignment"
}

func (r *PermissionSetAssignmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a Stax Permission Set Assignment resource. This provides a mapping which links [Stax Permission Sets](https://support.stax.io/hc/en-us/articles/4453967433359-Permission-Sets), Stax Groups and Stax Account types.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Permission Set Assignment identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"permission_set_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the Permission Set associated with this Assignment",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"account_type_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the Account Type associated with this Assignment",
				Required:            true,
			},
			"group_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the Group associated with this Assignment",
				Required:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the stax Permission Set Assignment",
				Computed:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "The identifier of the stax user who created the Permission Set Assignment",
				Computed:            true,
			},
			"created_ts": schema.StringAttribute{
				MarkdownDescription: "The Permission Set Assignment was creation timestamp",
				Computed:            true,
			},
		},
	}
}

func (r *PermissionSetAssignmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PermissionSetAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *PermissionSetAssignmentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := models.CreateAssignmentsRequest{}

	params = append(params, struct {
		AccountTypeId uuid.UUID "json:\"AccountTypeId\""
		GroupId       uuid.UUID "json:\"GroupId\""
	}{
		AccountTypeId: uuid.MustParse(data.AccountTypeID.ValueString()),
		GroupId:       uuid.MustParse(data.GroupID.ValueString()),
	})

	created, err := r.client.PermissionSetAssignmentCreate(ctx, data.PermissionSetID.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create permission set assignment, got error: %s", err))
		return
	}

	for _, assignment := range *created.JSON200 {
		assignmentAPIToTFResource(assignment, data)
	}

	deploymentCompletionStatuses := []models.AssignmentRecordStatus{models.DEPLOYMENTCOMPLETE, models.DEPLOYMENTFAILED}

	lastResponse, err := r.client.MonitorPermissionSetAssignments(ctx, data.PermissionSetID.ValueString(), data.ID.ValueString(), deploymentCompletionStatuses, &models.ListPermissionSetAssignmentsParams{}, func(ctx context.Context, lpsar *client.ListPermissionSetAssignmentsResponse) bool {
		tflog.Info(ctx, "polling complete for assignment", map[string]interface{}{
			"permission_set_id": data.PermissionSetID.ValueString(),
			"id":                data.ID.ValueString(),
		})

		return true
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create permission set assignment, got error: %s", err))
		return
	}

	status, ok := getAssignmentStatus(data.ID.ValueString(), lastResponse.JSON200.Assignments)
	if !ok {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create permission set assignment, unable to get status for id: %s", data.ID.ValueString()))
	}

	if status != models.DEPLOYMENTCOMPLETE {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create permission set assignment, ended with status: %s", status))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getAssignmentStatus(assignmentID string, assignments []models.AssignmentRecord) (models.AssignmentRecordStatus, bool) {
	for _, assignment := range assignments {
		if assignment.Id.String() == assignmentID {
			return assignment.Status, true
		}
	}

	return "", false
}

func (r *PermissionSetAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *PermissionSetAssignmentResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "reading permission set assignment", map[string]interface{}{
		"id":                data.ID.ValueString(),
		"permission_set_id": data.PermissionSetID.ValueString(),
	})

	read, err := r.client.PermissionSetAssignmentList(ctx, data.PermissionSetID.ValueString(), &models.ListPermissionSetAssignmentsParams{})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read permission set assignment, got error: %s", err))
		return
	}

	if !slices.ContainsFunc(read.JSON200.Assignments, containsAssignmentRecord(data.ID.ValueString())) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find assignment by permission set using id: %s", data.PermissionSetID.ValueString()))
		return
	}

	for _, assignment := range read.JSON200.Assignments {
		if assignment.Id.String() == data.ID.ValueString() {
			assignmentAPIToTFResource(assignment, data)
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionSetAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Client Error", "permission set assignments cannot be updated")
}

func (r *PermissionSetAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *PermissionSetAssignmentResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	_, err := r.client.PermissionSetAssignmentDelete(ctx, data.PermissionSetID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete permission set assignment, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "permission set assignment deleted", map[string]interface{}{
		"id":                data.ID.ValueString(),
		"permission_set_id": data.PermissionSetID.ValueString(),
	})

	deploymentCompletionStatuses := []models.AssignmentRecordStatus{models.DELETECOMPLETE, models.DELETEFAILED}

	lastResponse, err := r.client.MonitorPermissionSetAssignments(ctx, data.PermissionSetID.ValueString(), data.ID.ValueString(), deploymentCompletionStatuses, &models.ListPermissionSetAssignmentsParams{}, func(ctx context.Context, lpsar *client.ListPermissionSetAssignmentsResponse) bool {
		tflog.Info(ctx, "polling complete for assignment", map[string]interface{}{
			"permission_set_id": data.PermissionSetID.ValueString(),
			"id":                data.ID.ValueString(),
		})

		return true
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete permission set assignment, got error: %s", err))
		return
	}

	status, ok := getAssignmentStatus(data.ID.ValueString(), lastResponse.JSON200.Assignments)
	if !ok {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete permission set assignment, unable to get status for id: %s", data.ID.ValueString()))
	}

	if status != models.DELETECOMPLETE {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete permission set assignment, ended with status: %s", status))
		return
	}

}

func (r *PermissionSetAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	attrPath := path.Root("id")
	if attrPath.Equal(path.Empty()) {
		resp.Diagnostics.AddError(
			"Resource Import Passthrough Missing Attribute Path",
			"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				"Resource ImportState method call to ImportStatePassthroughID path must be set to a valid attribute path that can accept a string value.",
		)

		return
	}

	identifiers := strings.Split(req.ID, ":")

	if len(identifiers) != 2 {
		resp.Diagnostics.AddError(
			"Resource Import Failed",
			"Identifier must be in the format of \"permission_set_id:permission_set_assignment_id\" as both are required to import an assignment.")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("permission_set_id"), identifiers[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), identifiers[1])...)
}

func assignmentAPIToTFResource(assignment models.AssignmentRecord, data *PermissionSetAssignmentResourceModel) {
	data.ID = types.StringValue(assignment.Id.String())
	data.Status = types.StringValue(string(assignment.Status))
	data.GroupID = types.StringValue(assignment.GroupId.String())
	data.AccountTypeID = types.StringValue(assignment.AccountTypeId.String())
	data.CreatedBy = types.StringValue(assignment.CreatedBy.String())
	data.CreatedTS = types.StringValue(assignment.CreatedTS.Format(time.RFC3339))
}

func containsAssignmentRecord(id string) func(v models.AssignmentRecord) bool {
	return func(v models.AssignmentRecord) bool { return v.Id.String() == id }
}
