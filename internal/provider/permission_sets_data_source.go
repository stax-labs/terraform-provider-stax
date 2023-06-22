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

var _ datasource.DataSource = &PermissionSetsDataSource{}

func NewPermissionSetsDataSource() datasource.DataSource {
	return &PermissionSetsDataSource{}
}

// PermissionSetsDataSource defines the data source implementation.
type PermissionSetsDataSource struct {
	client staxsdk.ClientInterface
}

type PermissionSetDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Status             types.String `tfsdk:"status"`
	Description        types.String `tfsdk:"description"`
	CreatedBy          types.String `tfsdk:"created_by"`
	CreatedTS          types.String `tfsdk:"created_ts"`
	MaxSessionDuration types.Int64  `tfsdk:"max_session_duration"`
	Tags               types.Map    `tfsdk:"tags"`
}

// PermissionSetsDataSourceModel describes the data source data model.
type PermissionSetsDataSourceModel struct {
	ID             types.String                   `tfsdk:"id"`
	Filters        *PermissionSetsFiltersModel    `tfsdk:"filters"`
	PermissionSets []PermissionSetDataSourceModel `tfsdk:"permission_sets"`
}

type PermissionSetsFiltersModel struct {
	Statuses types.List `tfsdk:"statuses"`
}

func (d *PermissionSetsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission_sets"
}

func (d *PermissionSetsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Permission Sets datasource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Permission Set identifier used to select an group, this takes precedence over filters",
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
			"permission_sets": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The identifier of the stax Permission Set",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the stax Permission Set",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "The status of the stax Permission Set, can be ACTIVE or DELETED",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "The description of the stax Permission Set",
							Computed:            true,
						},
						"created_by": schema.StringAttribute{
							MarkdownDescription: "The identifier of the stax user who created the Permission Set",
							Computed:            true,
						},
						"created_ts": schema.StringAttribute{
							MarkdownDescription: "The Permission Set was creation timestamp",
							Computed:            true,
						},
						"max_session_duration": schema.Int64Attribute{
							MarkdownDescription: "The max session duration used by this Permission Set",
							Computed:            true,
						},
						"tags": schema.MapAttribute{
							MarkdownDescription: "Permission Set tags",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *PermissionSetsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PermissionSetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PermissionSetsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// if an id is provided retrieve that entry and return it, otherwise use the filter
	if !data.ID.IsNull() {
		permissionSet, err := d.client.PermissionSetsReadByID(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read permissions sets, got error: %s", err))
			return
		}

		permissionSetData, d := permissionSetAPIToTF(ctx, *permissionSet.JSON200)
		resp.Diagnostics.Append(d...)

		data.PermissionSets = append(data.PermissionSets, permissionSetData)
	} else {

		resp.Diagnostics.Append(d.readList(ctx, &data)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Trace(ctx, "read permissions sets from data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *PermissionSetsDataSource) readList(ctx context.Context, data *PermissionSetsDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	params := &models.ListPermissionSetsParams{}

	if data.Filters != nil && !data.Filters.Statuses.IsNull() {
		statuses := []models.PermissionSetStatus{}
		diags.Append(data.Filters.Statuses.ElementsAs(ctx, &statuses, false)...)

		params.Status = &statuses
	}

	tflog.Info(ctx, "building permission sets filters", map[string]interface{}{
		"params": params,
	})

	listPermissionSets, err := d.client.PermissionSetsList(ctx, params)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read permissions sets, got error: %s", err))
		return diags
	}

	tflog.Info(ctx, "reading permissions sets", map[string]interface{}{
		"count": len(listPermissionSets.JSON200.PermissionSets),
	})

	for _, permissionSet := range listPermissionSets.JSON200.PermissionSets {
		permissionSetData, d := permissionSetAPIToTF(ctx, permissionSet)
		diags.Append(d...)

		data.PermissionSets = append(data.PermissionSets, permissionSetData)
	}

	return diags
}

func permissionSetAPIToTF(ctx context.Context, permissionSet models.PermissionSetRecord) (PermissionSetDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	tags, d := types.MapValueFrom(ctx, types.StringType, permissionSet.Tags)
	diags.Append(d...)

	return PermissionSetDataSourceModel{
		ID:                 types.StringValue(permissionSet.Id.String()),
		Name:               types.StringValue(permissionSet.Name),
		Status:             types.StringValue(string(permissionSet.Status)),
		Description:        types.StringPointerValue(permissionSet.Description),
		CreatedBy:          types.StringValue(permissionSet.CreatedBy),
		CreatedTS:          types.StringValue(permissionSet.CreatedTS.Format(time.RFC3339)),
		MaxSessionDuration: types.Int64PointerValue(convertToI64Ptr(permissionSet.MaxSessionDuration)),
		Tags:               tags,
	}, diags
}

func convertToI64Ptr(i *int) *int64 {
	if i == nil {
		return nil
	}

	i64 := int64(*i)

	return &i64
}
