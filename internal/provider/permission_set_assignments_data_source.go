package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
)

var _ datasource.DataSource = &PermissionSetAssignmentsDataSource{}

func NewPermissionSetAssignmentsDataSource() datasource.DataSource {
	return &PermissionSetAssignmentsDataSource{}
}

// PermissionSetAssignmentsDataSource defines the data source implementation.
type PermissionSetAssignmentsDataSource struct {
	client staxsdk.ClientInterface
}

type PermissionSetAssignmentDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	AccountTypeID  types.String `tfsdk:"account_type_id"`
	GroupID        types.String `tfsdk:"group_id"`
	Status         types.String `tfsdk:"status"`
	LastModifiedBy types.String `tfsdk:"last_modified_by"`
	LastModifiedTS types.String `tfsdk:"last_modified_ts"`
	CreatedBy      types.String `tfsdk:"created_by"`
	CreatedTS      types.String `tfsdk:"created_ts"`
}

// PermissionSetAssignmentsDataSourceModel describes the data source data model.
type PermissionSetAssignmentsDataSourceModel struct {
	PermissionSetID types.String                             `tfsdk:"permission_set_id"`
	Filters         *PermissionSetAssignmentsFiltersModel    `tfsdk:"filters"`
	Assignments     []PermissionSetAssignmentDataSourceModel `tfsdk:"assignments"`
}

type PermissionSetAssignmentsFiltersModel struct {
	Statuses types.List `tfsdk:"statuses"`
}

func (d *PermissionSetAssignmentsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission_set_assignments"
}

func (d *PermissionSetAssignmentsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Permission Set Assignments datasource",

		Attributes: map[string]schema.Attribute{
			"permission_set_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the stax Permission Set associated with the assignments",
				Required:            true,
			},
			"filters": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"statuses": schema.ListAttribute{
						MarkdownDescription: "A list of statuses used to filter stax Permission Sets",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"assignments": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The identifier of the stax Permission Set",
							Computed:            true,
						},
						"account_type_id": schema.StringAttribute{
							MarkdownDescription: "Permission Set Assignment Account Type identifier",
							Computed:            true,
						},
						"group_id": schema.StringAttribute{
							MarkdownDescription: "Permission Set Assignment Group identifier",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "The status of the stax Permission Set Assignment",
							Computed:            true,
						},
						"created_by": schema.StringAttribute{
							MarkdownDescription: "The Stax User who created the Permission Set Assignment",
							Computed:            true,
						},
						"created_ts": schema.StringAttribute{
							MarkdownDescription: "The Permission Set Assignment creation timestamp",
							Computed:            true,
						},
						"last_modified_by": schema.StringAttribute{
							MarkdownDescription: "The Stax User who last modified the Permission Set Assignment",
							Computed:            true,
						},
						"last_modified_ts": schema.StringAttribute{
							MarkdownDescription: "The Permission Set Assignment last modified timestamp",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *PermissionSetAssignmentsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*staxsdk.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *PermissionSetAssignmentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PermissionSetAssignmentsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	resp.Diagnostics.Append(d.readList(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read permissions sets from data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *PermissionSetAssignmentsDataSource) readList(ctx context.Context, data *PermissionSetAssignmentsDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	params := &models.ListPermissionSetAssignmentsParams{}

	if data.Filters != nil && !data.Filters.Statuses.IsNull() {
		statuses := []models.AssignmentRecordStatus{}
		diags.Append(data.Filters.Statuses.ElementsAs(ctx, &statuses, false)...)

		params.Status = &statuses
	}

	tflog.Info(ctx, "building permission sets filters", map[string]interface{}{
		"params": params,
	})

	listPermissionSets, err := d.client.PermissionSetAssignmentList(ctx, data.PermissionSetID.ValueString(), params)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read permissions sets, got error: %s", err))
		return diags
	}

	tflog.Info(ctx, "reading permissions sets", map[string]interface{}{
		"count": len(listPermissionSets.JSON200.Assignments),
	})

	for _, assignment := range listPermissionSets.JSON200.Assignments {
		data.Assignments = append(data.Assignments, assignmentAPIToTF(assignment))
	}

	return diags
}

func assignmentAPIToTF(assignment models.AssignmentRecord) PermissionSetAssignmentDataSourceModel {
	return PermissionSetAssignmentDataSourceModel{
		ID:             types.StringValue(assignment.Id.String()),
		AccountTypeID:  types.StringValue(assignment.AccountTypeId.String()),
		GroupID:        types.StringValue(assignment.GroupId.String()),
		Status:         types.StringValue(string(assignment.Status)),
		CreatedTS:      types.StringValue(assignment.CreatedTS.Format(time.RFC3339)),
		CreatedBy:      types.StringValue(assignment.CreatedBy.String()),
		LastModifiedBy: types.StringPointerValue(assignment.ModifiedBy),
		LastModifiedTS: types.StringPointerValue(timeToStringPtr(assignment.ModifiedTS)),
	}
}

func timeToStringPtr(ts *time.Time) *string {
	if ts == nil {
		return nil
	}

	s := (*ts).Format(time.RFC3339)

	return &s
}
